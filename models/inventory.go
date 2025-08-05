package models

import "time"

type Inventory struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `json:"name"`
	Price         float64   `json:"price"`
	Stock         int       `json:"stock"`
	Description   string    `json:"description"`
	AvailableTime time.Time `json:"available_time"`
	EstimateTime  time.Time `json:"estimate_time"`
}
