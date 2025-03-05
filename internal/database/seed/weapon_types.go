package seed

import (
	"log"

	"github.com/hail2skins/the-virtual-armory/internal/models"
	"gorm.io/gorm"
)

// SeedWeaponTypes seeds the database with common weapon types
func SeedWeaponTypes(db *gorm.DB) {
	// Define common weapon types
	weaponTypes := []models.WeaponType{
		// Handguns (pistols & revolvers)
		{Type: "Handgun", Nickname: "Pistol"},
		{Type: "Revolver", Nickname: "Revolver"},

		// Long guns
		{Type: "Shotgun", Nickname: "Shotgun"},
		{Type: "Rifle", Nickname: "Rifle"},

		// Subcategories of rifles
		{Type: "Semi-Automatic Rifle", Nickname: "AR"},
		{Type: "Carbine", Nickname: "Carbine"},
		{Type: "Sniper Rifle", Nickname: "Sniper"},
		{Type: "Designated Marksman Rifle", Nickname: "DMR"},

		// Machine and Automatic Weapons
		{Type: "Machine Gun", Nickname: "MG"},
		{Type: "Submachine Gun", Nickname: "SMG"},
		{Type: "Personal Defense Weapon", Nickname: "PDW"},

		// Other specialist types
		{Type: "Anti-Materiel Rifle", Nickname: "AMR"},
		{Type: "Battle Rifle", Nickname: "Battle Rifle"},       // Typically more robust than standard rifles.
		{Type: "Precision Rifle", Nickname: "Precision Rifle"}, // For highly accurate rifles.
		{Type: "Lever-Action Rifle", Nickname: "Lever Rifle"},
		{Type: "Bolt-Action Rifle", Nickname: "Bolt Rifle"},
		{Type: "Semi-Automatic Rifle", Nickname: "Semi-Auto Rifle"},
		{Type: "Semi-Automatic Shotgun", Nickname: "Semi-Auto Shotgun"},
		{Type: "Pump-Action Shotgun", Nickname: "Pump Shotgun"},
		// Add additional categories as desired.
	}

	// Loop through each weapon type
	for _, wt := range weaponTypes {
		var count int64
		// Check if the record exists (by Type)
		if err := db.Model(&models.WeaponType{}).Where("type = ?", wt.Type).Count(&count).Error; err != nil {
			log.Printf("Error checking weapon type %s: %v", wt.Type, err)
			continue
		}
		if count == 0 {
			if err := db.Create(&wt).Error; err != nil {
				log.Printf("Error seeding weapon type %s: %v", wt.Type, err)
			} else {
				log.Printf("Seeded weapon type: %s", wt.Type)
			}
		} else {
			log.Printf("Weapon type %s already exists - skipping", wt.Type)
		}
	}
}
