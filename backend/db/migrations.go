package db

import (
	"fmt"

	"gorm.io/gorm"
)

// ApplyMigrations creates tables and FTS structures idempotently using raw SQL.
func ApplyMigrations(gdb *gorm.DB) error {
	stmts := []string{
		// Pragma & tables
		"PRAGMA foreign_keys = ON;",
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
			model_name TEXT,
			model_hash TEXT,
			prompt TEXT,
			negative_prompt TEXT,
			sampler TEXT,
			steps INTEGER,
			cfg_scale REAL,
			seed TEXT,
			scheduler TEXT,
			clip_skip INTEGER,
			nsfw INTEGER DEFAULT 0,
			hidden INTEGER DEFAULT 0,
			raw_metadata TEXT
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
		`CREATE TABLE IF NOT EXISTS user_metadata (
                        id INTEGER PRIMARY KEY,
                        image_id INTEGER NOT NULL,
                        key TEXT NOT NULL,
                        value TEXT NOT NULL,
                        UNIQUE(image_id, key),
                        FOREIGN KEY (image_id) REFERENCES images(id) ON DELETE CASCADE
                );`,
		`CREATE TABLE IF NOT EXISTS settings (
                        key TEXT PRIMARY KEY,
                        value TEXT NOT NULL
                );`,
		// Indexes
		`CREATE INDEX IF NOT EXISTS images_nsfw_idx ON images(nsfw);`,
		`CREATE INDEX IF NOT EXISTS images_model_idx ON images(model_name);`,
		`CREATE INDEX IF NOT EXISTS image_tags_image_idx ON image_tags(image_id);`,
		`CREATE INDEX IF NOT EXISTS image_tags_tag_idx ON image_tags(tag_id);`,
		// FTS5 virtual table (content-linked)
		`CREATE VIRTUAL TABLE IF NOT EXISTS images_fts USING fts5(
			file_name, model_name, prompt, negative_prompt, raw_metadata,
			content='images', content_rowid='id'
		);`,
		// Triggers to sync FTS
		`CREATE TRIGGER IF NOT EXISTS images_ai AFTER INSERT ON images BEGIN
			INSERT INTO images_fts(rowid, file_name, model_name, prompt, negative_prompt, raw_metadata)
			VALUES (new.id, new.file_name, new.model_name, new.prompt, new.negative_prompt, new.raw_metadata);
		END;`,
		`CREATE TRIGGER IF NOT EXISTS images_ad AFTER DELETE ON images BEGIN
			INSERT INTO images_fts(images_fts, rowid, file_name) VALUES('delete', old.id, old.file_name);
		END;`,
		`CREATE TRIGGER IF NOT EXISTS images_au AFTER UPDATE ON images BEGIN
			INSERT INTO images_fts(images_fts, rowid, file_name) VALUES('delete', old.id, old.file_name);
			INSERT INTO images_fts(rowid, file_name, model_name, prompt, negative_prompt, raw_metadata)
			VALUES (new.id, new.file_name, new.model_name, new.prompt, new.negative_prompt, new.raw_metadata);
		END;`,
	}

	for _, s := range stmts {
		if err := gdb.Exec(s).Error; err != nil {
			return fmt.Errorf("migration failed on: %s\nerr: %w", s, err)
		}
	}
	return nil
}
