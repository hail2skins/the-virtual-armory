package seed

import (
	"log"

	"github.com/hail2skins/the-virtual-armory/internal/models"
	"gorm.io/gorm"
)

// SeedCalibers seeds the database with common calibers
func SeedCalibers(db *gorm.DB) {
	// Define common calibers
	calibers := []models.Caliber{
		// Small-caliber rimfire and pistol rounds:
		{Caliber: ".22 Long Rifle", Nickname: "22 LR"},
		{Caliber: ".22 Magnum", Nickname: "22 Mag"},
		{Caliber: ".25 ACP", Nickname: "25 ACP"},
		{Caliber: ".32 ACP", Nickname: "32 ACP"},
		{Caliber: ".32 S&W", Nickname: "32 S&W"},
		{Caliber: ".380 ACP", Nickname: "380"},
		{Caliber: "9mm Parabellum", Nickname: "9"},
		{Caliber: "9×19mm", Nickname: "9"},
		{Caliber: ".38 Special", Nickname: "38"},
		{Caliber: ".357 Magnum", Nickname: "357"},
		{Caliber: ".40 S&W", Nickname: "40"},
		{Caliber: ".44 Special", Nickname: "44"},
		{Caliber: ".44 Magnum", Nickname: "44 Mag"},
		{Caliber: ".45 ACP", Nickname: "45"},
		{Caliber: ".50 AE", Nickname: "50 AE"},

		// Common rifle calibers:
		{Caliber: ".223 Remington", Nickname: "223"},
		{Caliber: ".22-250 Remington", Nickname: "22-250"},
		{Caliber: ".243 Winchester", Nickname: "243"},
		{Caliber: ".270 Winchester", Nickname: "270"},
		{Caliber: ".308 Winchester", Nickname: "308"},
		{Caliber: ".30-06 Springfield", Nickname: "30-06"},
		{Caliber: ".300 Winchester Magnum", Nickname: "300 WM"},
		{Caliber: "5.56×45mm NATO", Nickname: "5.56"},
		{Caliber: "6.5 Creedmoor", Nickname: "6.5"}, // Often written as 6.5mm

		// Intermediate and less common rifle rounds:
		{Caliber: "7.62×39mm", Nickname: "7.62"},
		{Caliber: "7.62×51mm NATO", Nickname: "7.62 NATO"},
		{Caliber: "7.62×54mm R", Nickname: "7.62 R"},
		{Caliber: ".300 AAC Blackout", Nickname: "300 BLK"},
		{Caliber: "6.8 SPC", Nickname: "6.8 SPC"},
		{Caliber: "6mm Creedmoor", Nickname: "6 Creedmoor"},

		// Big bore and magnum calibers:
		{Caliber: ".338 Lapua Magnum", Nickname: "338 Lapua"},
		{Caliber: ".375 H&H Magnum", Nickname: "375 H&H"},
		{Caliber: ".458 Winchester Magnum", Nickname: "458 WM"},
		{Caliber: ".416 Rigby", Nickname: "416 Rigby"},
		{Caliber: ".500 S&W Magnum", Nickname: "500 S&W"},
		{Caliber: ".338 Federal", Nickname: "338 Fed"},

		// Common Shotgun Shells:
		{Caliber: "12 Gauge", Nickname: "12"},
		{Caliber: "20 Gauge", Nickname: "20"},
		{Caliber: "28 Gauge", Nickname: "28"},
		{Caliber: ".410 Bore", Nickname: "410"},

		// Less-common / Historical Gauges:
		{Caliber: "10 Gauge", Nickname: "10"},
		{Caliber: "16 Gauge", Nickname: "16"},
	}

	// Loop through each caliber
	for _, c := range calibers {
		var count int64
		// Check if the record exists (by Caliber)
		if err := db.Model(&models.Caliber{}).Where("caliber = ?", c.Caliber).Count(&count).Error; err != nil {
			log.Printf("Error checking caliber %s: %v", c.Caliber, err)
			continue
		}
		if count == 0 {
			if err := db.Create(&c).Error; err != nil {
				log.Printf("Error seeding caliber %s: %v", c.Caliber, err)
			} else {
				log.Printf("Seeded caliber: %s", c.Caliber)
			}
		}
	}
}
