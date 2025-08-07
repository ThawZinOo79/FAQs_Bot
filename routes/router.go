package routes

import (
	"faqs-bot/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Keep API group for other routes
	api := r.Group("/api")
	{
		api.POST("/login", controllers.Login)
		api.POST("/register", controllers.Register)
	}

	// Direct webhook path for Facebook to access
	r.GET("/webhook", controllers.VerifyWebhook)
	r.POST("/webhook", controllers.HandleMessage)

	return r
}
