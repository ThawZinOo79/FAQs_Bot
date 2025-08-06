package repositories

import (
	"faqs-bot/config"
	"faqs-bot/models"
)

// CreateOrder saves an order to the stocks table
func CreateOrder(customerID, inventoryID uint, accountLink, orderType string) (*models.Stock, error) {
	stock := models.Stock{
		CustomerID:  customerID,
		InventoryID: inventoryID,
		AccountLink: accountLink,
		OrderType:   orderType,
	}

	if err := config.DB.Create(&stock).Error; err != nil {
		return nil, err
	}

	return &stock, nil
}
