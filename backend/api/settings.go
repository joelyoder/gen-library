package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"gen-library/backend/db"
)

func getSetting(gdb *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Param("key")
		var s db.Setting
		if err := gdb.First(&s, "key = ?", key).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusOK, gin.H{"value": ""})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		c.JSON(http.StatusOK, gin.H{"value": s.Value})
	}
}

func setSetting(gdb *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Param("key")
		var body struct {
			Value string `json:"value"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		s := db.Setting{Key: key, Value: body.Value}
		if err := gdb.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			UpdateAll: true,
		}).Create(&s).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"value": s.Value})
	}
}
