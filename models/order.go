package models

import "time"

type Order struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CustomerID  uint      `json:"customer_id"`
	InventoryID uint      `json:"inventory_id"`
	OrderDate   time.Time `json:"order_date"`
	Quantity    int       `json:"quantity"`
}
