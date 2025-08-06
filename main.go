package main

import (
	"faqs-bot/config"
	"faqs-bot/routes"
	"faqs-bot/controllers"
)

func main() {
	config.ConnectDB() // Connect and migrate once
	// godotenv.Load()
    controllers.LoadTexts()
	r := routes.SetupRouter()
	r.Run(":8080")
}
