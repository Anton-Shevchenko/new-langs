package model

type BookPart struct {
	ID     int64 `gorm:"primaryKey;autoIncrement"`
	Number int64
	BookId int64
	Text   string `gorm:"type:text"`
}
