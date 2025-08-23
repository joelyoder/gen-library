package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	api := r.Group("/api")
	{
		api.GET("/images", listImages(db))
		api.GET("/images/:id", getImage(db))
		api.GET("/images/:id/file", serveImage(db))
		api.PUT("/images/:id/metadata", updateMetadata(db))
		api.POST("/images/:id/tags", addTags(db))
		api.DELETE("/images/:id/tags", removeTags(db))
		api.DELETE("/images/:id", deleteImage(db))
		api.POST("/scan", scanFolder(db))
		api.GET("/settings/:key", getSetting(db))
		api.PUT("/settings/:key", setSetting(db))
		api.GET("/watcher", getWatcherStatus())
		api.POST("/watcher/start", startWatcher(db))
		api.POST("/watcher/stop", stopWatcher())
	}
}
