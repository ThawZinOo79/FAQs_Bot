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
		api.POST("/register", controllers.Register) // 👈 Add this line
	}

	return r
}
