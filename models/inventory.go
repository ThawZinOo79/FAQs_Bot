package models

import (
	"time"

	"gorm.io/datatypes"
)

type Inventory struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `json:"name"`
	Price         float64        `json:"price"`
	Stock         int            `json:"stock"`
	Description   string         `json:"description"`
	Category      string         `json:"category"`   // e.g., "Makeup", "Home Accessories"
	Attributes    datatypes.JSON `json:"attributes"` // flexible fields in JSON
	AvailableTime time.Time      `json:"available_time"`
	EstimateTime  time.Time      `json:"estimate_time"`
}
