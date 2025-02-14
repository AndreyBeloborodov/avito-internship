package models

import "gorm.io/gorm"

// Transaction - структура для хранения информации о покупке
type Transaction struct {
	gorm.Model
	SenderId   uint `gorm:"not null" json:"senderId"`
	ReceiverId uint `gorm:"not null" json:"receiverId"`
	Amount     int  `gorm:"not null" json:"amount"`
}
