package models

type Stock struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	CustomerID  uint   `json:"customer_id"`
	InventoryID uint   `json:"inventory_id"`
	AccountLink string `json:"account_link"`
	OrderType   string `json:"order_type"`

	Customer  Customer  `gorm:"foreignKey:CustomerID"`
	Inventory Inventory `gorm:"foreignKey:InventoryID"`
}
