package service

import (
	"langs/internal/model"
	"langs/internal/repository/word_repository"
	"langs/pkg/language_detector"
	"strings"
)

type WordService struct {
	wordRepository *word_repository.WordRepository
}

func NewWordService(wordRepository *word_repository.WordRepository) *WordService {
	return &WordService{
		wordRepository: wordRepository,
	}
}

func (s *WordService) AddWord(source, translation string, user *model.User) (*model.Word, error) {
	sourceLang, err := language_detector.Detect(source, user.GetUserLangs())

	if err != nil {
		return nil, err
	}

	translateLang, err := language_detector.Detect(translation, user.GetUserLangs())

	if err != nil {
		return nil, err
	}

	fields := strings.Fields(source)

	word := &model.Word{
		IsSimpleWord:    len(fields) == 1,
		Value:           source,
		Translation:     translation,
		ChatId:          user.ChatId,
		ValueLang:       sourceLang,
		TranslationLang: translateLang,
	}

	savedWord, err := s.wordRepository.CheckSimilarWord(word)

	if err != nil {
		return nil, err
	}

	if savedWord.ID != 0 {
		return word, nil
	}

	err = s.wordRepository.Create(word)

	if err != nil {
		return nil, err
	}

	return nil, nil
}
