package main

import (
	"faqs-bot/config"
	"faqs-bot/models"
	"faqs-bot/routes"
)

func main() {
	config.ConnectDB()
	config.DB.AutoMigrate(&models.User{})

	r := routes.SetupRouter()
	r.Run(":8080")
}
