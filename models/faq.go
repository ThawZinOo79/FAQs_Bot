package models

import "gorm.io/gorm"

type Faq struct {
	gorm.Model
	Question string `json:"question"`
	Answer   string `json:"answer"`
}
