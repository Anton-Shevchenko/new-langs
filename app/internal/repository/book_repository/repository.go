package book_repository

import (
	"gorm.io/gorm"
	"langs/internal/model"
)

type BookRepository struct {
	db *gorm.DB
}

func NewBookRepository(db *gorm.DB) *BookRepository {
	return &BookRepository{
		db: db,
	}
}

func (r *BookRepository) AllByChatId(chatId int64) ([]*model.Book, error) {
	var books []*model.Book

	return books, r.db.
		Where("chat_id = ?", chatId).
		Order("id desc").
		Find(&books).Error
}

func (r *BookRepository) Find(id int64) (*model.Book, error) {
	var bookPart model.Book
	err := r.db.Where("id = ?", id).First(&bookPart).Error

	if err != nil {
		return nil, err
	}

	return &bookPart, nil
}

func (r *BookRepository) Create(book *model.Book) error {
	return r.db.Create(&book).Error
}

func (r *BookRepository) Update(book *model.Book) error {
	return r.db.Save(&book).Error
}
