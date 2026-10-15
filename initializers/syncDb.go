package initializers

import (
	"log"
	"urlShortener/models"
)

func SyncDb() {
	err := DB.AutoMigrate(&models.URL{})
	if err != nil {
		log.Fatal("failed to create a table", err)
	}
}
