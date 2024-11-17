package jobs

import (
	"github.com/go-telegram/ui/keyboard/inline"
	"langs/internal/interfaces"
	"langs/internal/model"
	"math/rand"
	"time"
)

type SendWordJob struct {
	wordRepository   interfaces.WordRepository
	userRepository   interfaces.UserRepository
	messengerService interfaces.MessengerService
}

func NewSendWordJob(
	wordRepository interfaces.WordRepository,
	messengerService interfaces.MessengerService,
	userRepository interfaces.UserRepository,
) *SendWordJob {
	return &SendWordJob{
		wordRepository:   wordRepository,
		messengerService: messengerService,
		userRepository:   userRepository,
	}
}

type WordOption struct {
	Word            string
	Translation     string
	TranslationLang string
}

func (j *SendWordJob) Execute(handle inline.OnSelect) {
	users, err := j.userRepository.GetAllByInterval(30)
	if err != nil {
		return
	}

	for _, user := range users {
		j.sendTest(user, handle)
	}
}

func (j *SendWordJob) sendTest(user *model.User, handle inline.OnSelect) {
	word, err := j.wordRepository.GetRandomWordByChatId(user.ChatID)

	if err != nil {
		return
	}

	pair := [2]*WordOption{
		{Word: word.Value, Translation: word.Translation, TranslationLang: word.TranslationLang},
		{Word: word.Translation, Translation: word.Value, TranslationLang: word.ValueLang},
	}

	randomWord := j.getRandomWordOption(pair)
	translations, err := j.wordRepository.GetRandomTranslationsByChatId(
		user.ChatID,
		randomWord.Translation,
		randomWord.TranslationLang,
		3,
	)
	translations = append(translations, randomWord.Translation)

	if err != nil {
		return
	}

	j.messengerService.SendWordTest(user.ChatID, handle, randomWord.Word, j.shuffleTranslations(translations))
}

func (j *SendWordJob) getRand() *rand.Rand {
	source := rand.NewSource(time.Now().UnixNano())
	return rand.New(source)
}

func (j *SendWordJob) getRandomWordOption(pair [2]*WordOption) *WordOption {
	return pair[j.getRand().Intn(len(pair))]
}

func (j *SendWordJob) shuffleTranslations(slice []string) []string {
	j.getRand().Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})

	return slice
}
