//go:build main
// +build main

package main

import (
	"errors"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"

	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"gen-library/backend/api"
	"gen-library/backend/db"
	"gen-library/backend/logger"
	"gen-library/backend/scan"
)

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		c.Set("RequestID", id)
		c.Writer.Header().Set("X-Request-ID", id)
		c.Next()
	}
}

func main() {
	logger.Init()
	defer logger.Close()

	if logger.Level() == zerolog.DebugLevel {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	dbConn, err := gorm.Open(sqlite.Open("library.db"), &gorm.Config{})
	if err != nil {
		logger.Error().Err(err).Msg("failed to open database")
		os.Exit(1)
	}

	if err := db.ApplyMigrations(dbConn); err != nil {
		logger.Error().Err(err).Msg("migrations failed")
		os.Exit(1)
	}

	var root string
	if err := dbConn.Table("settings").Select("value").Where("key=?", "library_path").Scan(&root).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Warn().Err(err).Msg("failed to read library_path")
	} else if root != "" {
		go scan.StartWatcher(root, dbConn)
	}

	r := gin.New()
	r.SetTrustedProxies(nil) // fix "trusted all proxies" warning
	r.Use(requestIDMiddleware())
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		evt := logger.Info().
			Str("method", param.Method).
			Str("path", param.Path).
			Int("status", param.StatusCode).
			Dur("latency", param.Latency).
			Str("client_ip", param.ClientIP)
		if id, ok := param.Keys["RequestID"].(string); ok {
			evt.Str("request_id", id)
		}
		evt.Msg("request")
		return ""
	}))
	r.Use(gin.Recovery())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5174"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
	}))

	r.StaticFS("/thumbs", http.Dir(".cache/thumbs"))
	api.RegisterRoutes(r, dbConn)

	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "8081" // new default backend port
	}
	logger.Info().Str("port", port).Msg("Backend listening")
	if err := r.Run(":" + port); err != nil {
		logger.Error().Err(err).Msg("server error")
		os.Exit(1)
	}
}
