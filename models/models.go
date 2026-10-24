package models

import (
	"time"

	"gorm.io/gorm"
)

type Device struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Token     string    `gorm:"uniqueIndex;size:512" json:"token"`
	Provider  string    `gorm:"size:255" json:"provider"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TokenRecord struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Token     string    `gorm:"uniqueIndex;size:512" json:"token"`
	Provider  string    `gorm:"size:255" json:"provider"`
	Active    bool      `gorm:"default:true" json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
