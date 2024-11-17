package db

import (
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

func NewPostgresConnection(config string) *gorm.DB {
	dbURL := "host=db user=postgres password=password dbname=my_database port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	//db.Session(&gorm.Session{FullSaveAssociations: false})

	if err != nil {
		log.Fatalln(err)
	}

	return db
}
