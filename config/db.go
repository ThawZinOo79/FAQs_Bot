package config

import (
	"faqs-bot/models"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Build DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))

	// Configure custom logger
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // Output to console
		logger.Config{
			SlowThreshold:             time.Second, // Log only if query > 1s
			LogLevel:                  logger.Warn, // Show only warnings + slow queries
			IgnoreRecordNotFoundError: true,        // Ignore "record not found" errors
			Colorful:                  true,
		},
	)

	// Connect to MySQL
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	DB = database
	fmt.Println("📦 MySQL connected")

	// Run AutoMigrate
	err = DB.AutoMigrate(
		&models.User{},
		&models.Inventory{},
		&models.Customer{},
		&models.Stock{},
		&models.Faq{},
	)
	if err != nil {
		log.Fatal("Failed to migrate models:", err)
	}

	fmt.Println("✅ Database migrated")
}
