package main

import (
	"os"
	"urlShortener/initializers"
	"urlShortener/routes"

	"github.com/gin-gonic/gin"
)

func init() {
	initializers.LoadEnv()
	initializers.ConnectDB()
	initializers.SyncDb()
	initializers.ConnectRedis()
}

func main() {

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	gin.SetMode(gin.ReleaseMode) //disables debug logging for better performance

	router := gin.New()
	router.Use(gin.Recovery()) //keeps crash protection without per-request logging

	routes.UrlRoute(router)
	router.Run(":" + port)
}
