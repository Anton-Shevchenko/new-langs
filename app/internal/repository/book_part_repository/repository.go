package book_part_repository

import (
	"gorm.io/gorm"
	"langs/internal/model"
)

type BookPartRepository struct {
	db *gorm.DB
}

func NewBookPartRepository(db *gorm.DB) *BookPartRepository {
	return &BookPartRepository{
		db: db,
	}
}

func (r *BookPartRepository) Find(number, bookId int64) (*model.BookPart, error) {
	var bookPart model.BookPart
	err := r.db.Where("number = ? and book_id = ?", number, bookId).First(&bookPart).Error

	if err != nil {
		return nil, err
	}

	return &bookPart, nil
}

func (r *BookPartRepository) Add(part *model.BookPart) error {
	result := r.db.Model(&model.BookPart{}).Create(map[string]interface{}{
		"book_id": part.BookId,
		"text":    part.Text,
		"number":  gorm.Expr("(SELECT COALESCE(MAX(number), 0) + 1 FROM book_parts WHERE book_id = ?)", part.BookId),
	})

	return result.Error
}
