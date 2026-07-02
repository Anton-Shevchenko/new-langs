package service

import (
	"fmt"
	"langs/internal/domain"
	"langs/internal/infrastructure/book_reader"
	"langs/internal/infrastructure/database/repository/book_part_repository"
	"langs/internal/infrastructure/database/repository/book_progress_repository"
	"langs/internal/interfaces"
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

	bookPart, err := s.bookPartRepository.Find(bookPartId.BookPartId, bookId)
	if err == nil {
		return bookPart, nil
	}

	firstPart, fallbackErr := s.bookPartRepository.Find(1, bookId)
	if fallbackErr != nil {
		return nil, err
	}

	if err := s.bookProgressRepository.SetPage(bookId, 1); err != nil {
		return nil, err
	}

	return firstPart, nil
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
		BookPartId: 1,
	})

	parts := s.reader.ReadAndSplitBook(filePath, 100)
	s.batchInsert(parts, book, 100)
	book.Len = len(parts)
	s.bookRepository.Update(book)

	return book, nil
}

func (s *ReaderService) AddLongRead(longRead, name string, user *model.User) (*model.Book, error) {
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
		BookPartId: 1,
	})

	parts := s.reader.SplitBookFromString(longRead, 100)
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
