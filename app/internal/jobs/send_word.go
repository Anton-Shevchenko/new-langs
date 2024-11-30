package jobs

import (
	"github.com/go-telegram/ui/keyboard/inline"
	"langs/internal/interfaces"
	"langs/internal/service"
)

type SendWordJob struct {
	wordService    *service.WordService
	userRepository interfaces.UserRepository
}

func NewSendWordJob(
	wordService *service.WordService,
	userRepository interfaces.UserRepository,
) *SendWordJob {
	return &SendWordJob{
		wordService:    wordService,
		userRepository: userRepository,
	}
}

func (j *SendWordJob) Execute(handle inline.OnSelect) {
	users, err := j.userRepository.GetAllByInterval(30)
	if err != nil {
		return
	}

	for _, user := range users {
		j.wordService.SendTest(user, handle)
	}
}
