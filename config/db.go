package config

import (
	"faqs-bot/models"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))

	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	DB = database
	fmt.Println("📦 MySQL connected")

	err = DB.AutoMigrate(
		&models.User{},
		&models.Inventory{},
		&models.Customer{},
		&models.Stock{},
	)
	if err != nil {
		log.Fatal("Failed to migrate models:", err)
	}

	fmt.Println("✅ Database migrated")
}
