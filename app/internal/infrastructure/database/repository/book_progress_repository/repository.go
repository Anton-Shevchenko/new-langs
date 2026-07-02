package book_progress_repository

import (
	"errors"

	"gorm.io/gorm"
	"langs/internal/domain"
)

type BookProgressRepository struct {
	db *gorm.DB
}

func NewBookProgressRepository(db *gorm.DB) *BookProgressRepository {
	return &BookProgressRepository{
		db: db,
	}
}

func (r *BookProgressRepository) FindOrCreate(bookId int64) (*model.BookProgress, error) {
	progress, err := r.Find(bookId)
	if err == nil {
		return progress, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		newProgress := &model.BookProgress{
			BookId:     bookId,
			BookPartId: 1,
		}
		if err := r.Create(newProgress); err != nil {
			return nil, err
		}
		return newProgress, nil
	}

	return nil, err
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

func (r *BookProgressRepository) SetPage(bookId int64, page int64) error {
	return r.db.
		Model(&model.BookProgress{}).
		Where("book_id = ?", bookId).
		Update("book_part_id", page).
		Error
}

func (r *BookProgressRepository) Increment(bookId int64) (*model.BookProgress, error) {
	var progress model.BookProgress

	err := r.db.
		Model(&model.BookProgress{}).
		Where("book_id = ?", bookId).
		Where("book_part_id < (SELECT MAX(number) FROM book_parts WHERE book_id = ?)", bookId).
		Update("book_part_id", gorm.Expr("book_part_id + 1")).
		Error

	return &progress, err
}

func (r *BookProgressRepository) Decrement(bookId int64) (*model.BookProgress, error) {
	var progress model.BookProgress
	err := r.db.
		Model(&model.BookProgress{}).
		Where("book_id = ?", bookId).
		Where("book_part_id > (SELECT MIN(number) FROM book_parts WHERE book_id = ?)", bookId).
		Update("book_part_id", gorm.Expr("book_part_id - 1")).
		Error
	return &progress, err
}
