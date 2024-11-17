package service

import (
	"context"
	"fmt"
	"langs/internal/interfaces"
	"langs/internal/model"
	"langs/pkg/localizer_lib"
)

const defaultInterfaceLang = "en"

type userService struct {
	repository interfaces.UserRepository
}

func NewUserService(
	repository interfaces.UserRepository,
) interfaces.UserService {
	return &userService{
		repository: repository,
	}
}

func (s *userService) InitUser(chatId int64) *model.User {
	user, err := s.repository.First(chatId)

	if err == nil && user != nil {
		fmt.Println("localized")
		localizer_lib.LoadLang(user.InterfaceLang)
	} else {
		localizer_lib.LoadLang(defaultInterfaceLang)
	}

	return user
}

func (s *userService) Upsert(user *model.User) error {
	if user.ChatID != 0 {
		s.repository.Update(user)
	}

	s.repository.Create(user)

	return nil
}

func (s *userService) GetUserFromContext(ctx context.Context) (*model.User, error) {
	user, ok := ctx.Value("user").(*model.User)
	if !ok || user == nil {
		return nil, fmt.Errorf("user not found in context")
	}
	return user, nil
}
