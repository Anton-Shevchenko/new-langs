package jobs

import (
	"github.com/go-telegram/ui/keyboard/inline"
	"langs/internal/domain"
	"langs/internal/interfaces"
	"langs/internal/usecase"
)

const defaultTestsCount = 3

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

func (j *SendWordJob) Execute(handle, writeTestHandle inline.OnSelect) {
	users, err := j.userRepository.GetAll()
	if err != nil {
		return
	}

	for _, user := range users {
		j.executeWordTests(user, handle, writeTestHandle, defaultTestsCount)
	}
}

func (j *SendWordJob) executeWordTests(
	user *model.User,
	handle,
	writeTestHandle inline.OnSelect,
	testCount int,
) {
	if user.IsInQuietHours() {
		return
	}

	for i := 0; i < testCount; i++ {
		j.wordService.SendTest(user, handle, writeTestHandle)
	}
}
