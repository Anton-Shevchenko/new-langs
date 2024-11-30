package book_reader

import (
	"bufio"
	"log"
	"os"
	"strings"

	"gorm.io/gorm"
)

type BookPart struct {
	ID   int64  `gorm:"primaryKey;autoIncrement"`
	Text string `gorm:"type:text"`
}

type BookReader struct {
	db *gorm.DB
}

func NewBookReader(db *gorm.DB) *BookReader {
	return &BookReader{db: db}
}

func (br *BookReader) ReadAndSplitBook(filePath string, wordLimit int) []*BookPart {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open the book file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)

	return br.splitContent(scanner, wordLimit)
}

func (br *BookReader) SplitBookFromString(content string, wordLimit int) []*BookPart {
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Split(bufio.ScanWords)

	return br.splitContent(scanner, wordLimit)
}

func (br *BookReader) splitContent(scanner *bufio.Scanner, wordLimit int) []*BookPart {
	var parts []*BookPart
	var currentPart strings.Builder
	wordCount := 0

	for scanner.Scan() {
		word := scanner.Text()
		currentPart.WriteString(word + " ")
		wordCount++

		if wordCount >= wordLimit && strings.Contains(word, ".") {
			parts = append(parts, &BookPart{Text: strings.TrimSpace(currentPart.String())})
			currentPart.Reset()
			wordCount = 0
		}
	}

	if currentPart.Len() > 0 {
		parts = append(parts, &BookPart{Text: strings.TrimSpace(currentPart.String())})
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading the content: %v", err)
	}

	return parts
}
