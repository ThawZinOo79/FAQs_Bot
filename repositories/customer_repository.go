package repositories

import (
	"faqs-bot/config"
	"faqs-bot/models"
)

// CreateCustomer inserts a new customer record
func CreateCustomer(customer *models.Customer) error {
	return config.DB.Create(customer).Error
}
