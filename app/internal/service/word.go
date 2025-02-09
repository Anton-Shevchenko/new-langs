package service

import (
	"github.com/go-telegram/ui/keyboard/inline"
	"langs/internal/interfaces"
	"langs/internal/model"
	"langs/internal/repository/word_repository"
	"langs/pkg/language_detector"
	"math/rand"
	"strings"
	"time"
)

const defaultWriteExamsCount = 2

type WordService struct {
	wordRepository   *word_repository.WordRepository
	messengerService interfaces.MessengerService
}

func NewWordService(
	wordRepository *word_repository.WordRepository,
	messengerService interfaces.MessengerService,
) *WordService {
	return &WordService{
		wordRepository:   wordRepository,
		messengerService: messengerService,
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

func (s *WordService) SendTest(user *model.User, handle, handleWrite inline.OnSelect) {
	word, err := s.wordRepository.GetRandomWordByChatIdAndRateLimit(user.ChatId, user.TargetRate)

	if err != nil {
		return
	}

	restOfAnswersToTarget := user.TargetRate - uint16(word.Rate)
	pair := [2]*model.WordOption{
		{
			WordID:          word.ID,
			Word:            word.Value,
			Translation:     word.Translation,
			TranslationLang: word.TranslationLang,
			WordLang:        word.ValueLang,
		},
		{
			WordID:          word.ID,
			Word:            word.Translation,
			Translation:     word.Value,
			TranslationLang: word.ValueLang,
			WordLang:        word.TranslationLang,
		},
	}

	if restOfAnswersToTarget <= defaultWriteExamsCount {
		var nativeOption *model.WordOption
		if pair[0].WordLang == user.NativeLang {
			nativeOption = pair[0]
		} else {
			nativeOption = pair[1]
		}

		s.messengerService.SendExam(user.ChatId, handleWrite, nativeOption)
		return
	}

	randomWord := s.getRandomWordOption(pair)

	translations, err := s.wordRepository.GetRandomTranslationsByChatId(
		user.ChatId,
		randomWord.Translation,
		randomWord.TranslationLang,
		3,
	)
	translations = append(translations, randomWord.Translation)

	if err != nil {
		return
	}

	s.messengerService.SendWordTest(
		user.ChatId,
		handle,
		randomWord,
		s.shuffleTranslations(translations),
	)
}

func (s *WordService) getRand() *rand.Rand {
	source := rand.NewSource(time.Now().UnixNano())
	return rand.New(source)
}

func (s *WordService) getRandomWordOption(pair [2]*model.WordOption) *model.WordOption {
	return pair[s.getRand().Intn(len(pair))]
}

func (s *WordService) shuffleTranslations(slice []string) []string {
	s.getRand().Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})

	return slice
}
