package model

type Book struct {
	ID     int64 `gorm:"primaryKey;autoIncrement"`
	Name   string
	Parts  []*BookPart
	FileId int
	ChatId int64
	Len    int
}
