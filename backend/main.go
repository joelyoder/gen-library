package main

import (
	"errors"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"

	"gorm.io/gorm"

	"gen-library/backend/api"
	"gen-library/backend/db"
	"gen-library/backend/logger"
	"gen-library/backend/scan"
)

func main() {
	logger.Init()

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

	r := gin.Default()
	r.SetTrustedProxies(nil) // fix "trusted all proxies" warning

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
