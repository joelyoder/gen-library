package db

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// ApplyMigrations creates tables and FTS structures idempotently using raw SQL.
func ApplyMigrations(gdb *gorm.DB) error {
	stmts := []string{
		// Pragma & tables
		"PRAGMA foreign_keys = ON;",
		`CREATE TABLE IF NOT EXISTS models (
                       id INTEGER PRIMARY KEY,
                       name TEXT UNIQUE NOT NULL,
                       hash TEXT
               );`,
		`CREATE TABLE IF NOT EXISTS images (
                        id INTEGER PRIMARY KEY,
                        path TEXT UNIQUE NOT NULL,
                        file_name TEXT NOT NULL,
                        ext TEXT NOT NULL,
                        size_bytes INTEGER NOT NULL,
                        sha256 TEXT UNIQUE NOT NULL,
                        width INTEGER,
                        height INTEGER,
                        created_time DATETIME,
                        imported_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                        source_app TEXT,
                        model_id INTEGER,
                        prompt TEXT,
                        negative_prompt TEXT,
                        sampler TEXT,
                        steps INTEGER,
                        cfg_scale REAL,
                        seed TEXT,
                        scheduler TEXT,
                        clip_skip INTEGER,
                        variation_seed INTEGER,
                        variation_seed_strength REAL,
                        aspect_ratio TEXT,
                        refiner_control_percentage REAL,
                        refiner_upscale REAL,
                        refiner_upscale_method TEXT,
                        rating INTEGER DEFAULT 0,
                        nsfw INTEGER DEFAULT 0,
                        hidden INTEGER DEFAULT 0,
                        raw_metadata TEXT,
                        FOREIGN KEY (model_id) REFERENCES models(id)
                );`,
		`CREATE TABLE IF NOT EXISTS tags (
			id INTEGER PRIMARY KEY,
			name TEXT UNIQUE NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS image_tags (
			image_id INTEGER NOT NULL,
			tag_id INTEGER NOT NULL,
			PRIMARY KEY (image_id, tag_id),
			FOREIGN KEY (image_id) REFERENCES images(id) ON DELETE CASCADE,
			FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS settings (
                       key TEXT PRIMARY KEY,
                       value TEXT NOT NULL
               );`,
		`CREATE TABLE IF NOT EXISTS loras (
                       id INTEGER PRIMARY KEY,
                       name TEXT UNIQUE NOT NULL,
                       hash TEXT
               );`,
		`CREATE TABLE IF NOT EXISTS image_loras (
                       image_id INTEGER NOT NULL,
                       lora_id INTEGER NOT NULL,
                       weight REAL,
                       PRIMARY KEY (image_id, lora_id),
                       FOREIGN KEY (image_id) REFERENCES images(id) ON DELETE CASCADE,
                       FOREIGN KEY (lora_id) REFERENCES loras(id) ON DELETE CASCADE
               );`,
		`DROP TABLE IF EXISTS embeddings;`,
		`CREATE TABLE IF NOT EXISTS embeddings (
                       id INTEGER PRIMARY KEY,
                       name TEXT UNIQUE NOT NULL,
                       hash TEXT
               );`,
		`CREATE TABLE IF NOT EXISTS image_embeddings (
                       image_id INTEGER NOT NULL,
                       embedding_id INTEGER NOT NULL,
                       PRIMARY KEY (image_id, embedding_id),
                       FOREIGN KEY (image_id) REFERENCES images(id) ON DELETE CASCADE,
                       FOREIGN KEY (embedding_id) REFERENCES embeddings(id) ON DELETE CASCADE
               );`,
		// Indexes
		`CREATE INDEX IF NOT EXISTS images_nsfw_idx ON images(nsfw);`,
		`CREATE INDEX IF NOT EXISTS images_rating_idx ON images(rating);`,
		`CREATE INDEX IF NOT EXISTS images_model_idx ON images(model_id);`,
		`CREATE INDEX IF NOT EXISTS models_hash_idx ON models(hash);`,
		`CREATE INDEX IF NOT EXISTS image_tags_image_idx ON image_tags(image_id);`,
		`CREATE INDEX IF NOT EXISTS image_tags_tag_idx ON image_tags(tag_id);`,
		`CREATE INDEX IF NOT EXISTS loras_hash_idx ON loras(hash);`,
		`CREATE INDEX IF NOT EXISTS image_loras_image_idx ON image_loras(image_id);`,
		`CREATE INDEX IF NOT EXISTS image_loras_lora_idx ON image_loras(lora_id);`,
		`CREATE INDEX IF NOT EXISTS embeddings_hash_idx ON embeddings(hash);`,
		`CREATE INDEX IF NOT EXISTS image_embeddings_image_idx ON image_embeddings(image_id);`,
		`CREATE INDEX IF NOT EXISTS image_embeddings_embedding_idx ON image_embeddings(embedding_id);`,
	}

	for _, s := range stmts {
		if err := gdb.Exec(s).Error; err != nil {
			return fmt.Errorf("migration failed on: %s\nerr: %w", s, err)
		}
	}

	// Conditional schema adjustments for legacy databases
	if exists, err := columnExists(gdb, "loras", "image_id"); err != nil {
		return err
	} else if exists {
		if err := gdb.Exec(`ALTER TABLE loras DROP COLUMN image_id;`).Error; err != nil {
			return fmt.Errorf("failed dropping loras.image_id: %w", err)
		}
	}

	if exists, err := columnExists(gdb, "image_loras", "weight"); err != nil {
		return err
	} else if !exists {
		if err := gdb.Exec(`ALTER TABLE image_loras ADD COLUMN weight REAL;`).Error; err != nil {
			return fmt.Errorf("failed adding image_loras.weight: %w", err)
		}
	}

	if exists, err := columnExists(gdb, "images", "model_id"); err != nil {
		return err
	} else if !exists {
		if err := gdb.Exec(`ALTER TABLE images ADD COLUMN model_id INTEGER;`).Error; err != nil {
			return fmt.Errorf("failed adding images.model_id: %w", err)
		}
	}

	if exists, err := columnExists(gdb, "images", "model_name"); err != nil {
		return err
	} else if exists {
		if err := gdb.Exec(`ALTER TABLE images DROP COLUMN model_name;`).Error; err != nil {
			return fmt.Errorf("failed dropping images.model_name: %w", err)
		}
	}

	if exists, err := columnExists(gdb, "images", "model_hash"); err != nil {
		return err
	} else if exists {
		if err := gdb.Exec(`ALTER TABLE images DROP COLUMN model_hash;`).Error; err != nil {
			return fmt.Errorf("failed dropping images.model_hash: %w", err)
		}
	}

	// Optional FTS5 setup; ignore if module unavailable
	ftsStmts := []string{
		`CREATE VIRTUAL TABLE IF NOT EXISTS images_fts USING fts5(
                       file_name, model_name, prompt, negative_prompt, raw_metadata,
                       content='images', content_rowid='id'
               );`,
		`CREATE TRIGGER IF NOT EXISTS images_ai AFTER INSERT ON images BEGIN
                       INSERT INTO images_fts(rowid, file_name, model_name, prompt, negative_prompt, raw_metadata)
                       VALUES (new.id, new.file_name, (SELECT name FROM models WHERE id = new.model_id), new.prompt, new.negative_prompt, new.raw_metadata);
               END;`,
		`CREATE TRIGGER IF NOT EXISTS images_ad AFTER DELETE ON images BEGIN
                       INSERT INTO images_fts(images_fts, rowid, file_name) VALUES('delete', old.id, old.file_name);
               END;`,
		`CREATE TRIGGER IF NOT EXISTS images_au AFTER UPDATE ON images BEGIN
                       INSERT INTO images_fts(images_fts, rowid, file_name) VALUES('delete', old.id, old.file_name);
                       INSERT INTO images_fts(rowid, file_name, model_name, prompt, negative_prompt, raw_metadata)
                       VALUES (new.id, new.file_name, (SELECT name FROM models WHERE id = new.model_id), new.prompt, new.negative_prompt, new.raw_metadata);
               END;`,
	}
	for _, s := range ftsStmts {
		if err := gdb.Exec(s).Error; err != nil {
			if strings.Contains(err.Error(), "no such module: fts5") {
				break
			}
			return fmt.Errorf("migration failed on: %s\nerr: %w", s, err)
		}
	}

	return nil
}

// columnExists checks whether a column is present on a table.
func columnExists(gdb *gorm.DB, table, column string) (bool, error) {
	type col struct{ Name string }
	var cols []col
	if err := gdb.Raw(fmt.Sprintf("PRAGMA table_info(%s);", table)).Scan(&cols).Error; err != nil {
		return false, err
	}
	for _, c := range cols {
		if c.Name == column {
			return true, nil
		}
	}
	return false, nil
}
