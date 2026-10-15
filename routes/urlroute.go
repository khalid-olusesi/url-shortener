package routes

import (
	"urlShortener/controllers"

	"github.com/gin-gonic/gin"
)

func UrlRoute(router *gin.Engine) {
	router.POST("/url", controllers.CreateURL)
	router.GET("/:shortcode", controllers.GetURL) //: was used because the url will be dynamic
}
