package db

import (
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

func NewPostgresConnection(config string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	//db.Session(&gorm.Session{FullSaveAssociations: false})

	if err != nil {
		log.Fatalln(err)
	}

	return db
}
