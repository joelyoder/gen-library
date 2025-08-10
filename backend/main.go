package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"

	"gorm.io/gorm"

	"gen-library/backend/api"
	"gen-library/backend/db"
)

func main() {
	dbConn, err := gorm.Open(sqlite.Open("library.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	if err := db.ApplyMigrations(dbConn); err != nil {
		log.Fatalf("migrations failed: %v", err)
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
	log.Println("Backend listening on :" + port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
