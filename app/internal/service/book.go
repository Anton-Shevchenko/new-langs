package service

import (
	"fmt"
	"langs/internal/interfaces"
	"langs/internal/model"
	"langs/internal/repository/book_part_repository"
	"langs/internal/repository/book_progress_repository"
	"langs/pkg/book_reader"
)

type ReaderService struct {
	bookRepository         interfaces.BookRepository
	bookPartRepository     *book_part_repository.BookPartRepository
	reader                 *book_reader.BookReader
	bookProgressRepository *book_progress_repository.BookProgressRepository
}

func NewReaderService(
	bookRepository interfaces.BookRepository,
	bookPartRepository *book_part_repository.BookPartRepository,
	reader *book_reader.BookReader,
	bookProgressRepository *book_progress_repository.BookProgressRepository,
) *ReaderService {
	return &ReaderService{
		bookRepository:         bookRepository,
		bookPartRepository:     bookPartRepository,
		reader:                 reader,
		bookProgressRepository: bookProgressRepository,
	}
}

func (s *ReaderService) ReadBookPart(bookId int64) (*model.BookPart, error) {
	bookPartId, err := s.bookProgressRepository.FindOrCreate(bookId)
	if err != nil {
		return nil, err
	}
	return s.bookPartRepository.Find(bookPartId.BookPartId, bookId)
}

func (s *ReaderService) GetBooks(user *model.User) ([]*model.Book, error) {
	books, err := s.bookRepository.AllByChatId(user.ChatId)

	if err != nil {
		return nil, err
	}

	return books, nil
}

func (s *ReaderService) AddBook(filePath, name string, user *model.User) (*model.Book, error) {
	book := &model.Book{
		ChatId: user.ChatId,
		Name:   name,
	}
	err := s.bookRepository.Create(book)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	s.bookProgressRepository.Create(&model.BookProgress{
		BookId:     book.ID,
		BookPartId: book.ID,
	})

	parts := s.reader.ReadAndSplitBook(filePath, 100)
	s.batchInsert(parts, book, 100)
	book.Len = len(parts)
	s.bookRepository.Update(book)

	return nil, nil
}

func (s *ReaderService) batchInsert(parts []*book_reader.BookPart, book *model.Book, batchSize int) {
	for _, part := range parts {
		s.bookPartRepository.Add(&model.BookPart{
			BookId: book.ID,
			Text:   part.Text,
		})
	}
}

func (s *ReaderService) NextPage(bookId int64) {
	s.bookProgressRepository.Increment(bookId)
}

func (s *ReaderService) PrevPage(bookId int64) {
	s.bookProgressRepository.Decrement(bookId)
}
