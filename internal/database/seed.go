package database

import (
	"log"

	"github.com/hail2skins/the-virtual-armory/internal/models"
	"gorm.io/gorm"
)

// SeedManufacturers populates the manufacturers table with a list of firearm manufacturers.
func SeedManufacturers(db *gorm.DB) {
	// List of manufacturers with an optional country.
	manufacturers := []models.Manufacturer{
		{Name: "Glock", Country: "Austria"},
		{Name: "Smith & Wesson", Country: "USA"},
		{Name: "Ruger", Country: "USA"},
		{Name: "Sig Sauer", Country: "Switzerland"},
		{Name: "Beretta", Country: "Italy"},
		{Name: "Colt", Country: "USA"},
		{Name: "Heckler & Koch", Country: "Germany"},
		{Name: "Remington", Country: "USA"},
		{Name: "Winchester", Country: "USA"},
		{Name: "Savage Arms", Country: "USA"},
		{Name: "Mossberg", Country: "USA"},
		{Name: "Browning", Country: "USA"},
		{Name: "CZ", Country: "Czech Republic"},
		{Name: "FN Herstal", Country: "Belgium"},
		{Name: "Taurus", Country: "Brazil"},
		{Name: "Walther", Country: "Germany"},
		{Name: "Springfield Armory", Country: "USA"},
		{Name: "Kimber", Country: "USA"},
		{Name: "Daniel Defense", Country: "USA"},
		{Name: "Benelli", Country: "Italy"},
		{Name: "Tikka", Country: "Finland"},
		{Name: "Sako", Country: "Finland"},
		{Name: "Weatherby", Country: "USA"},
		{Name: "Marlin", Country: "USA"},
		{Name: "Steyr", Country: "Austria"},
		{Name: "Barrett", Country: "USA"},
		{Name: "Kel-Tec", Country: "USA"},
		{Name: "LWRC", Country: "USA"},
		{Name: "Noveske", Country: "USA"},
		{Name: "Christensen Arms", Country: "USA"},
		{Name: "Bergara", Country: "Spain"},
		{Name: "Canik", Country: "Turkey"},
		{Name: "IWI", Country: "Israel"},
		{Name: "Stoeger", Country: "Italy"},
		{Name: "Henry", Country: "USA"},
		{Name: "Chiappa", Country: "Italy"},
		{Name: "Kahr", Country: "USA"},
		{Name: "Rossi", Country: "Brazil"},
		{Name: "Zastava", Country: "Serbia"},
		{Name: "Century Arms", Country: "USA"},
		{Name: "Palmetto State Armory", Country: "USA"},
		{Name: "Aero Precision", Country: "USA"},
		{Name: "Shadow Systems", Country: "USA"},
		{Name: "Wilson Combat", Country: "USA"},
		{Name: "Nighthawk Custom", Country: "USA"},
		{Name: "Les Baer", Country: "USA"},
		{Name: "Ed Brown", Country: "USA"},
		{Name: "STI", Country: "USA"},
		{Name: "Staccato", Country: "USA"},
		{Name: "Zev Technologies", Country: "USA"},
	}

	// Loop through each manufacturer and insert if it doesn't already exist.
	for _, m := range manufacturers {
		var count int64
		if err := db.Model(&models.Manufacturer{}).Where("name = ?", m.Name).Count(&count).Error; err != nil {
			log.Printf("Error checking manufacturer %s: %v", m.Name, err)
			continue
		}
		if count == 0 {
			if err := db.Create(&m).Error; err != nil {
				log.Printf("Error seeding manufacturer %s: %v", m.Name, err)
			} else {
				log.Printf("Seeded manufacturer: %s", m.Name)
			}
		} else {
			log.Printf("Manufacturer %s already exists - skipping", m.Name)
		}
	}
}

// SeedCalibers populates the calibers table with a list of firearm calibers.
func SeedCalibers(db *gorm.DB) {
	// List of calibers with an optional nickname.
	calibers := []models.Caliber{
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
		{Caliber: ".223 Remington", Nickname: "223"},
		{Caliber: ".22-250 Remington", Nickname: "22-250"},
		{Caliber: ".243 Winchester", Nickname: "243"},
		{Caliber: ".270 Winchester", Nickname: "270"},
		{Caliber: ".308 Winchester", Nickname: "308"},
		{Caliber: ".30-06 Springfield", Nickname: "30-06"},
		{Caliber: ".300 Winchester Magnum", Nickname: "300 WM"},
		{Caliber: "5.56×45mm NATO", Nickname: "5.56"},
		{Caliber: "6.5 Creedmoor", Nickname: "6.5"},
		{Caliber: "7.62×39mm", Nickname: "7.62"},
		{Caliber: "7.62×51mm NATO", Nickname: "7.62 NATO"},
		{Caliber: "7.62×54mm R", Nickname: "7.62 R"},
		{Caliber: ".300 AAC Blackout", Nickname: "300 BLK"},
		{Caliber: "6.8 SPC", Nickname: "6.8 SPC"},
		{Caliber: "6mm Creedmoor", Nickname: "6 Creedmoor"},
		{Caliber: ".338 Lapua Magnum", Nickname: "338 Lapua"},
		{Caliber: ".375 H&H Magnum", Nickname: "375 H&H"},
		{Caliber: ".458 Winchester Magnum", Nickname: "458 WM"},
		{Caliber: ".416 Rigby", Nickname: "416 Rigby"},
		{Caliber: ".500 S&W Magnum", Nickname: "500 S&W"},
		{Caliber: ".338 Federal", Nickname: "338 Fed"},
		{Caliber: "12 Gauge", Nickname: "12"},
		{Caliber: "20 Gauge", Nickname: "20"},
		{Caliber: "28 Gauge", Nickname: "28"},
		{Caliber: ".410 Bore", Nickname: "410"},
		{Caliber: "10 Gauge", Nickname: "10"},
		{Caliber: "16 Gauge", Nickname: "16"},
	}

	// Loop through each caliber and insert if it doesn't already exist.
	for _, c := range calibers {
		var count int64
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
		} else {
			log.Printf("Caliber %s already exists - skipping", c.Caliber)
		}
	}
}

// SeedWeaponTypes populates the weapon_types table with a list of firearm categories.
func SeedWeaponTypes(db *gorm.DB) {
	// List of firearm types with an optional nickname.
	weaponTypes := []models.WeaponType{
		// Handguns (pistols & revolvers)
		{Type: "Handgun", Nickname: "Pistol"},
		{Type: "Revolver", Nickname: "Revolver"},

		// Long guns
		{Type: "Shotgun", Nickname: "Shotgun"},
		{Type: "Rifle", Nickname: "Rifle"},

		// Subcategories of rifles
		{Type: "Assault Rifle", Nickname: "AR"}, // Although many ARs are labeled assault rifles, you might change wording if needed.
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
	}

	// Loop through each weapon type and insert if it doesn't already exist.
	for _, wt := range weaponTypes {
		var count int64
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
