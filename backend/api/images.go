package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gen-library/backend/db"
	"gen-library/backend/scan"
	"gen-library/backend/util"
)

type imageDTO struct {
	ID        uint    `json:"id"`
	Path      string  `json:"path"`
	FileName  string  `json:"fileName"`
	Ext       string  `json:"ext"`
	Width     *int    `json:"width"`
	Height    *int    `json:"height"`
	ModelName *string `json:"modelName"`
	Prompt    *string `json:"prompt"`
	NSFW      bool    `json:"nsfw"`
	ThumbURL  string  `json:"thumbUrl"`
}

func listImages(gdb *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))
		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 200 {
			pageSize = 50
		}

		nsfwMode := strings.ToLower(c.DefaultQuery("nsfw", "hide")) // hide|show|only
		q := c.Query("q")
		csvTags := c.Query("tags")
		var tags []string
		if csvTags != "" {
			tags = splitNonEmpty(csvTags, ",")
		}

		sort := c.DefaultQuery("sort", "imported_at")
		order := c.DefaultQuery("order", "desc")
		if !inSet(sort, []string{"created_time", "imported_at", "file_name"}) {
			sort = "imported_at"
		}
		if !inSet(strings.ToLower(order), []string{"asc", "desc"}) {
			order = "desc"
		}

		// Base query
		img := gdb.Table("images")
		// NSFW filter
		switch nsfwMode {
		case "hide":
			img = img.Where("images.nsfw = 0")
		case "only":
			img = img.Where("images.nsfw = 1")
		}

		// FTS join if q
		if strings.TrimSpace(q) != "" {
			img = img.Joins("JOIN images_fts ON images_fts.rowid = images.id").Where("images_fts MATCH ?", q)
		}

		// Tag filter: require ALL tags
		if len(tags) > 0 {
			sub := gdb.Table("image_tags it").
				Select("it.image_id").
				Joins("JOIN tags t ON t.id = it.tag_id").
				Where("t.name IN ?", tags).
				Group("it.image_id").
				Having("COUNT(DISTINCT t.name) = ?", len(tags))
			img = img.Where("images.id IN (?)", sub)
		}

		// Count total
		var total int64
		if err := img.Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Select page
		rows := []imageDTO{}
		qimg := img.Order("images." + sort + " " + strings.ToUpper(order)).
			Select("images.id, images.path, images.file_name, images.ext, images.width, images.height, images.model_name, images.prompt, images.nsfw").
			Limit(pageSize).Offset((page - 1) * pageSize)

		if err := qimg.Scan(&rows).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var root string
		gdb.Table("settings").Select("value").Where("key=?", "library_path").Scan(&root)
		for i := range rows {
			var sha string
			if err := gdb.Table("images").Select("sha256").Where("id=?", rows[i].ID).Scan(&sha).Error; err == nil && sha != "" {
				src := rows[i].Path
				if root != "" && !filepath.IsAbs(src) {
					src = filepath.Join(root, src)
				}
				_, _ = util.EnsureThumb(sha, src, 400)
				rows[i].ThumbURL = "/thumbs/" + sha + "_400.jpg"
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"page":     page,
			"pageSize": pageSize,
			"total":    total,
			"items":    rows,
		})
	}
}

func getImage(gdb *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var m db.Image
		if err := gdb.Preload("Tags").Preload("UserMeta").Preload("Loras").First(&m, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, m)
	}
}

func serveImage(gdb *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var img db.Image
		if err := gdb.First(&img, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		var root string
		gdb.Table("settings").Select("value").Where("key=?", "library_path").Scan(&root)
		path := img.Path
		if root != "" && !filepath.IsAbs(path) {
			path = filepath.Join(root, path)
		}
		c.File(path)
	}
}

func updateMetadata(gdb *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var payload map[string]any
		if err := c.BindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var (
			loras        []db.Lora
			lorasPresent bool
		)
		if raw, ok := payload["loras"]; ok {
			lorasPresent = true
			delete(payload, "loras")
			if arr, ok := raw.([]any); ok {
				for _, r := range arr {
					obj, ok := r.(map[string]any)
					if !ok {
						continue
					}
					name, _ := obj["name"].(string)
					hash, _ := obj["hash"].(string)
					loras = append(loras, db.Lora{Name: name, Hash: hash})
				}
			}
		}

		updates := make(map[string]any, len(payload))
		for k, v := range payload {
			updates[camelToSnake(k)] = v
		}

		if err := gdb.Model(&db.Image{}).Where("id=?", id).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if lorasPresent {
			uid, err := strconv.ParseUint(id, 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if err := gdb.Where("image_id = ?", uid).Delete(&db.Lora{}).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if len(loras) > 0 {
				for i := range loras {
					loras[i].ImageID = uint(uid)
				}
				if err := gdb.Create(&loras).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}
		}

		getImage(gdb)(c)
	}
}

func addTags(gdb *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var body struct {
			Tags []string `json:"tags"`
		}
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var image db.Image
		if err := gdb.First(&image, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		seen := make(map[string]struct{})
		for _, name := range body.Tags {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}

			var t db.Tag
			if err := gdb.Where("name = ?", name).First(&t).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					if err := gdb.Create(&db.Tag{Name: name}).Error; err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
						return
					}
					if err := gdb.Where("name = ?", name).First(&t).Error; err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
						return
					}
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}

			rel := db.ImageTag{ImageID: image.ID, TagID: t.ID}
			if err := gdb.FirstOrCreate(&rel, rel).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		getImage(gdb)(c)
	}
}

func removeTags(gdb *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var body struct {
			Tags []string `json:"tags"`
		}
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var image db.Image
		if err := gdb.Select("id").First(&image, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		seen := make(map[string]struct{})
		for _, name := range body.Tags {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}

			var t db.Tag
			if err := gdb.Where("name = ?", name).First(&t).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if err := gdb.Where("image_id = ? AND tag_id = ?", image.ID, t.ID).Delete(&db.ImageTag{}).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			var count int64
			if err := gdb.Model(&db.ImageTag{}).Where("tag_id = ?", t.ID).Count(&count).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if count == 0 {
				if err := gdb.Delete(&db.Tag{}, t.ID).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}
		}

		getImage(gdb)(c)
	}
}

func deleteImage(gdb *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		mode := strings.ToLower(c.DefaultQuery("mode", "trash"))

		var token string
		if mode == "hard" {
			var body struct {
				Token string `json:"token"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			token = body.Token
		}

		err := gdb.Transaction(func(tx *gorm.DB) error {
			var img db.Image
			if err := tx.First(&img, id).Error; err != nil {
				return err
			}

			if err := tx.Delete(&db.Image{}, id).Error; err != nil {
				return err
			}

			// Resolve the file path against the configured library
			// root (if any) and ensure we work with an absolute
			// path. This avoids platform specific resolution issues
			// when moving files to the trash. If any of the
			// conversions fail, propagate the error so the
			// transaction is rolled back and surfaced to the
			// caller.
			absPath := img.Path
			var root string
			if err := tx.Table("settings").Select("value").Where("key=?", "library_path").Scan(&root).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if root != "" && !filepath.IsAbs(absPath) {
				absPath = filepath.Join(root, absPath)
			}
			absPath, err := filepath.Abs(absPath)
			if err != nil {
				return err
			}

			switch mode {
			case "trash":
				return moveToTrash(absPath)
			case "hard":
				expected := fmt.Sprintf("%d", img.ID)
				if token != expected {
					return fmt.Errorf("invalid token")
				}
				return os.Remove(absPath)
			default:
				return fmt.Errorf("unknown mode")
			}
		})

		if err != nil {
			switch {
			case errors.Is(err, gorm.ErrRecordNotFound):
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			case strings.Contains(err.Error(), "invalid token"):
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			case strings.Contains(err.Error(), "unknown mode"):
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

func moveToTrash(path string) error {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("powershell", "-NoProfile", "-Command",
			fmt.Sprintf(`Add-Type -AssemblyName Microsoft.VisualBasic; [Microsoft.VisualBasic.FileIO.FileSystem]::DeleteFile(%q, [Microsoft.VisualBasic.FileIO.UIOption]::OnlyErrorDialogs, [Microsoft.VisualBasic.FileIO.RecycleOption]::SendToRecycleBin)`, path))
		// Capture stdout/stderr so any PowerShell errors are surfaced to the caller.
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("powershell recycle failed: %v: %s", err, strings.TrimSpace(string(out)))
		}
		return nil
	case "darwin":
		script := fmt.Sprintf(`tell application \"Finder\" to delete POSIX file %q`, path)
		cmd := exec.Command("osascript", "-e", script)
		return cmd.Run()
	default: // freedesktop trash spec
		abs, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		filesDir := filepath.Join(home, ".local/share/Trash/files")
		infoDir := filepath.Join(home, ".local/share/Trash/info")
		if err := os.MkdirAll(filesDir, 0o755); err != nil {
			return err
		}
		if err := os.MkdirAll(infoDir, 0o755); err != nil {
			return err
		}
		base := filepath.Base(abs)
		dest := filepath.Join(filesDir, base)
		for i := 1; ; i++ {
			if _, err := os.Stat(dest); os.IsNotExist(err) {
				break
			}
			dest = filepath.Join(filesDir, fmt.Sprintf("%s.%d", base, i))
		}
		if err := os.Rename(abs, dest); err != nil {
			return err
		}
		infoPath := filepath.Join(infoDir, filepath.Base(dest)+".trashinfo")
		u := url.PathEscape(abs)
		ts := time.Now().Format("2006-01-02T15:04:05")
		content := fmt.Sprintf("[Trash Info]\nPath=%s\nDeletionDate=%s\n", u, ts)
		return os.WriteFile(infoPath, []byte(content), 0o644)
	}
}

func scanFolder(gdb *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			Root string `json:"root"`
		}
		if err := c.ShouldBindJSON(&body); err != nil && !errors.Is(err, io.EOF) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		root := body.Root
		if root == "" {
			var s db.Setting
			if err := gdb.First(&s, "key = ?", "library_path").Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "library path not set"})
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				}
				return
			}
			root = s.Value
		}
		if root == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no root provided"})
			return
		}
		n, err := scan.ScanFolder(gdb, root)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"count": n})
	}
}

func inSet(v string, arr []string) bool {
	for _, a := range arr {
		if v == a {
			return true
		}
	}
	return false
}

func splitNonEmpty(s, sep string) []string {
	parts := strings.Split(s, sep)
	res := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			res = append(res, t)
		}
	}
	return res
}

func camelToSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
