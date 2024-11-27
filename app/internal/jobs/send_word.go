package jobs

import (
	"github.com/go-telegram/ui/keyboard/inline"
	"langs/internal/interfaces"
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

func (j *SendWordJob) Execute(handle inline.OnSelect) {
	users, err := j.userRepository.GetAllByInterval(30)
	if err != nil {
		return
	}

	for _, user := range users {
		j.sendTest(user, handle)
	}
}
