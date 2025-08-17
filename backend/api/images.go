package api

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gen-library/backend/db"
	"gen-library/backend/scan"
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
			img = img.Where("nsfw = 0")
		case "only":
			img = img.Where("nsfw = 1")
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
		qimg := img.Order(sort + " " + strings.ToUpper(order)).
			Select("id, path, file_name, ext, width, height, model_name, prompt, nsfw").
			Limit(pageSize).Offset((page - 1) * pageSize)

		if err := qimg.Scan(&rows).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		for i := range rows {
			// Derive SHA to form thumb URL; get by extra query (lightweight)
			var sha string
			if err := gdb.Table("images").Select("sha256").Where("id=?", rows[i].ID).Scan(&sha).Error; err == nil && sha != "" {
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
		if err := gdb.Preload("Tags").Preload("UserMeta").First(&m, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, m)
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
		if err := gdb.Model(&db.Image{}).Where("id=?", id).Updates(payload).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
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
		for _, name := range body.Tags {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			var t db.Tag
			if err := gdb.Where("name=?", name).First(&t).Error; err != nil {
				// create if not exists
				gdb.Create(&db.Tag{Name: name})
				gdb.Where("name=?", name).First(&t)
			}
			gdb.Create(&db.ImageTag{ImageID: image.ID, TagID: t.ID})
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
		if len(body.Tags) > 0 {
			gdb.Exec(`DELETE FROM image_tags WHERE image_id = ? AND tag_id IN (SELECT id FROM tags WHERE name IN ?)`, image.ID, body.Tags)
		}
		getImage(gdb)(c)
	}
}

func deleteImage(gdb *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var path string
		if err := gdb.Table("images").Select("path").Where("id=?", id).Scan(&path).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		// Hard delete for MVP
		if err := gdb.Delete(&db.Image{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// Remove file best-effort
		_ = os.Remove(path)
		c.JSON(http.StatusOK, gin.H{"ok": true})
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
