package api

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    "gen-library/backend/scan"
)

func getLibraryFolder(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var path string
        db.Raw("SELECT value FROM settings WHERE key = ?", "libraryFolder").Scan(&path)
        c.JSON(http.StatusOK, gin.H{"path": path})
    }
}

func setLibraryFolder(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var body struct{ Path string `json:"path"` }
        if err := c.BindJSON(&body); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        if err := db.Exec("INSERT INTO settings(key,value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value", "libraryFolder", body.Path).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"path": body.Path})
    }
}

func importLibrary(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var path string
        db.Raw("SELECT value FROM settings WHERE key = ?", "libraryFolder").Scan(&path)
        added, updated, err := scan.ScanFolder(db, path)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"added": added, "updated": updated})
    }
}

