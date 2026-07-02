package db

import (
	"gorm.io/gorm"
)

type DB struct {
	DB *gorm.DB
}

func NewDB(connection *gorm.DB) DB {
	return DB{
		DB: connection,
	}
}
