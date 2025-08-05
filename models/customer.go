package models

import "time"

type Customer struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	OrderDate   time.Time `json:"order_date"`
	AccountLink string    `json:"account_link"`
}
