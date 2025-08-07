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
	Category      string         `json:"category"`
	Attributes    datatypes.JSON `json:"attributes"`
	ImageURL      string         `json:"image_url"` // <- S3 image link field
	AvailableTime time.Time      `json:"available_time"`
	EstimateTime  time.Time      `json:"estimate_time"`
}
