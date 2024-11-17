package book_progress_repository

import (
	"gorm.io/gorm"
	"langs/internal/model"
)

type BookProgressRepository struct {
	db *gorm.DB
}

func NewBookProgressRepository(db *gorm.DB) *BookProgressRepository {
	return &BookProgressRepository{
		db: db,
	}
}

func (r *BookProgressRepository) Find(bookId int64) (*model.BookProgress, error) {
	var progress model.BookProgress
	err := r.db.Where("book_id = ?", bookId).First(&progress).Error

	if err != nil {
		return nil, err
	}

	return &progress, nil
}

func (r *BookProgressRepository) Create(progress *model.BookProgress) error {
	return r.db.Create(&progress).Error
}

func (r *BookProgressRepository) Increment(bookId int64) (*model.BookProgress, error) {
	var progress model.BookProgress
	err := r.db.
		Model(&model.BookProgress{}).
		Where("book_id = ?", bookId).
		Update("book_part_id", gorm.Expr("book_part_id + ?", 1)).
		Error

	return &progress, err
}

func (r *BookProgressRepository) Decrement(bookId int64) (*model.BookProgress, error) {
	var progress model.BookProgress
	err := r.db.
		Model(&model.BookProgress{}).
		Where("book_id = ?", bookId).
		Update("book_part_id", gorm.Expr("book_part_id - ?", 1)).
		Error

	return &progress, err
}
