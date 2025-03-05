package seed

import (
	"log"

	"github.com/hail2skins/the-virtual-armory/internal/models"
	"gorm.io/gorm"
)

// SeedManufacturers is our seed function that checks and inserts manufacturers.
func SeedManufacturers(db *gorm.DB) {
	// List of manufacturers to seed.
	manufacturers := []models.Manufacturer{
		// American Manufacturers:
		{Name: "Smith & Wesson", Country: "USA", Nickname: "S&W"},
		{Name: "Colt's Manufacturing Company", Country: "USA", Nickname: "Colt"},
		{Name: "Remington Arms", Country: "USA", Nickname: "Remington"},
		{Name: "Winchester Repeating Arms", Country: "USA", Nickname: "Winchester"},
		{Name: "Sturm, Ruger & Co.", Country: "USA", Nickname: "Ruger"},
		{Name: "Browning", Country: "USA", Nickname: "Browning"},
		{Name: "Taurus", Country: "Brazil/USA", Nickname: "Taurus"},
		{Name: "Kimber Manufacturing", Country: "USA", Nickname: "Kimber"},
		{Name: "Springfield Armory", Country: "USA", Nickname: "Springfield"},
		{Name: "Sig Sauer", Country: "Germany/USA", Nickname: "Sig"},
		{Name: "Heckler & Koch", Country: "Germany", Nickname: "H&K"},
		{Name: "Barrett Firearms Manufacturing", Country: "USA", Nickname: "Barrett"},
		{Name: "Bushmaster Firearms International", Country: "USA", Nickname: "Bushmaster"},
		{Name: "Franklin Armory", Country: "USA", Nickname: "Franklin"},
		{Name: "Accuracy International", Country: "UK", Nickname: "AI"},

		// International Manufacturers:
		{Name: "Glock", Country: "Austria", Nickname: "Glock"},
		{Name: "Beretta", Country: "Italy", Nickname: "Beretta"},
		{Name: "Česká zbrojovka (CZ)", Country: "Czech Republic", Nickname: "CZ"},
		{Name: "FN Herstal", Country: "Belgium", Nickname: "FN"},
		{Name: "Steyr Mannlicher", Country: "Austria", Nickname: "Steyr"},
		{Name: "Walther", Country: "Germany", Nickname: "Walther"},
		{Name: "IWI (Israel Weapon Industries)", Country: "Israel", Nickname: "IWI"},

		// Smaller or Specialty Manufacturers:
		{Name: "Kel-Tec", Country: "USA", Nickname: "Kel-Tec"},
		{Name: "Rossi", Country: "USA/Brazil", Nickname: "Rossi"},
		{Name: "Charter Arms", Country: "USA", Nickname: "Charter"},
		{Name: "Uberti", Country: "Italy/USA", Nickname: "Uberti"},
		{Name: "ArmaLite", Country: "USA", Nickname: "ArmaLite"},
		{Name: "Magnum Research", Country: "USA", Nickname: "Magnum"},

		// Classic/Historical Brands:
		{Name: "Mauser", Country: "Germany", Nickname: "Mauser"},
		{Name: "Luger", Country: "Germany", Nickname: "Luger"},
		{Name: "Webley", Country: "UK", Nickname: "Webley"},
		{Name: "Enfield", Country: "UK", Nickname: "Enfield"},

		// Custom / High-End Specialty Manufacturers:
		{Name: "Wilson Combat", Country: "USA", Nickname: "Wilson"},
		{Name: "Les Baer", Country: "USA", Nickname: "Baer"},
		{Name: "Nighthawk Custom", Country: "USA", Nickname: "Nighthawk"},
		// You can also consider other custom shops:
		{Name: "Taran Tactical Innovations", Country: "USA", Nickname: "Taran"},
		{Name: "Ed Brown Products", Country: "USA", Nickname: "Ed Brown"},
		{Name: "CCI (Cascade Cartridge Inc.)", Country: "USA", Nickname: "CCI"},
	}

	// Loop through each manufacturer
	for _, m := range manufacturers {
		var count int64
		// Check if the record exists (by Name)
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
