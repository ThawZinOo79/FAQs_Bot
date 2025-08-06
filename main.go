package main

import (
	"faqs-bot/config"
	"faqs-bot/routes"
)

func main() {
	config.ConnectDB() // Connect and migrate once
	r := routes.SetupRouter()
	r.Run(":8080")
}
