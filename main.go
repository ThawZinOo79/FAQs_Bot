package main

import (
	"faqs-bot/config"
	"faqs-bot/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	config.ConnectDB()

	r := gin.Default()

	// Public Routes
	routes.AuthRoutes(r)

	// Protected Routes
	routes.CustomerRoutes(r)

	r.Run(":8080")
}
