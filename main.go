package main

import (
	"faqs-bot/config"
	"faqs-bot/routes"

	"log"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or failed to load")
	}

	config.ConnectDB() // Now environment variables are loaded

	r := routes.SetupRouter()
	r.Run(":8080")
}
