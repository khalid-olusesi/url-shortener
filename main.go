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

	router := gin.Default()

	routes.UrlRoute(router)
	router.Run(":" + port)
}
