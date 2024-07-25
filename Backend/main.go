package main

import (
	"time"

	"github.com/TenJit/SE/Backend/configs"
	"github.com/TenJit/SE/Backend/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func init() {
	configs.Connect_to_mongodb()
}

func main() {
	app := gin.Default()
	app.Static("/public", "./public")
	corsConfig := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	app.Use(cors.New(corsConfig))
	routes.UserRoute(app)
	routes.ImageRoute(app)
	app.Run(":8080")
}
