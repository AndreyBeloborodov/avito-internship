package models

import "gorm.io/gorm"

// User - структура пользователя
type User struct {
	gorm.Model
	Username string `gorm:"unique;not null" json:"username"`
	Password string `gorm:"not null" json:"-"`
	Coins    int    `json:"coins"`
}
