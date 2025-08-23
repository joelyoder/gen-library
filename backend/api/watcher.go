package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gen-library/backend/scan"
)

func getWatcherStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"running": scan.IsWatcherRunning()})
	}
}

func startWatcher(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var root string
		if err := db.Table("settings").Select("value").Where("key=?", "library_path").Scan(&root).Error; err != nil || root == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "library path not set"})
			return
		}
		go scan.StartWatcher(root, db)
		c.JSON(http.StatusOK, gin.H{"running": scan.IsWatcherRunning()})
	}
}

func stopWatcher() gin.HandlerFunc {
	return func(c *gin.Context) {
		scan.StopWatcher()
		c.JSON(http.StatusOK, gin.H{"running": scan.IsWatcherRunning()})
	}
}
