package routes

import (
	"faqs-bot/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	{
		api.POST("/login", controllers.Login)
		api.POST("/register", controllers.Register)
	}

	// Add Facebook webhook endpoint (no /api prefix)
	r.GET("/webhook", controllers.FBWebhookVerify)
	r.POST("/webhook", controllers.FBWebhookReceive)

	return r
}
