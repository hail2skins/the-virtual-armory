package models

import (
	"gorm.io/gorm"
)

// Caliber represents an ammunition caliber in the system
type Caliber struct {
	gorm.Model
	Caliber    string `gorm:"size:100;not null;unique" json:"caliber"`
	Nickname   string `gorm:"size:50" json:"nickname"`
	Popularity int    `gorm:"default:0" json:"popularity"` // Higher values appear first in dropdowns
}
