package db

import (
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"time"
)

func NewPostgresConnection(config string) *gorm.DB {
	var db *gorm.DB
	var err error

	for i := 0; i < 5; i++ {
		db, err = gorm.Open(postgres.Open(config), &gorm.Config{})
		if err == nil {
			log.Printf("Successfully connected to database on attempt %d", i+1)
			return db
		}

		log.Printf("Database connection attempt %d failed: %v", i+1, err)
		if i < 4 {
			log.Printf("Retrying in 2 seconds...")
			time.Sleep(2 * time.Second)
		}
	}

	log.Printf("Failed to connect to database after 5 attempts")
	return nil
}
