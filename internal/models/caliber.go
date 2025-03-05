package models

import (
	"gorm.io/gorm"
)

// Caliber represents an ammunition caliber in the system
type Caliber struct {
	gorm.Model
	Caliber  string `gorm:"size:100;not null;unique" json:"caliber"`
	Nickname string `gorm:"size:50" json:"nickname"`
}
