package models

import "gorm.io/gorm"

// Merch - структура пользователя
type Merch struct {
	gorm.Model
	Name  string `gorm:"unique;not null" json:"name"`
	Price int    `json:"price"`
}
