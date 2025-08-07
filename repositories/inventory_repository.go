package repositories

import (
	"faqs-bot/config"
	"faqs-bot/models"
)

// GetProductsByCategory retrieves all products in a category
func GetProductsByCategory(category string) ([]models.Inventory, error) {
	var products []models.Inventory
	result := config.DB.Where("category = ?", category).Find(&products)
	return products, result.Error
}

// GetDistinctCategories returns unique product categories
func GetDistinctCategories(categories *[]string) error {
	return config.DB.Model(&models.Inventory{}).Distinct().Pluck("category", categories).Error
}

// GetProductByName retrieves a single product by its name
func GetProductByName(name string) (*models.Inventory, error) {
	var product models.Inventory
	result := config.DB.Where("name = ?", name).First(&product)
	return &product, result.Error
}

// GetProductByIDString handles string input IDs
func GetProductByIDString(id string) (*models.Inventory, error) {
	var product models.Inventory
	result := config.DB.First(&product, id)
	return &product, result.Error
}

// GetProductByID retrieves a single product by its ID
func GetProductByID(id uint) (*models.Inventory, error) {
	var product models.Inventory
	result := config.DB.First(&product, id)
	return &product, result.Error
}

func GetAllProducts() ([]models.Inventory, error) {
	var products []models.Inventory
	result := config.DB.Find(&products)
	return products, result.Error
}


