package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gen-library/backend/db"
	"gen-library/backend/scan"
)

func getLibraryFolder(gdb *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var s db.Setting
		if err := gdb.First(&s, "key = ?", "libraryFolder").Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusOK, gin.H{"path": ""})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"path": s.Value})
	}
}

func setLibraryFolder(gdb *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			Path string `json:"path"`
		}
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := gdb.Save(&db.Setting{Key: "libraryFolder", Value: body.Path}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

func importLibrary(gdb *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var s db.Setting
		if err := gdb.First(&s, "key = ?", "libraryFolder").Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusOK, gin.H{"added": 0, "updated": 0})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		added, updated, err := scan.ScanFolder(gdb, s.Value)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"added": added, "updated": updated})
	}
}
