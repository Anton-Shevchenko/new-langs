package service

import (
	"context"
	"fmt"
	"langs/internal/domain"
	"langs/internal/interfaces"
	"langs/pkg/nlp/localizer_lib"
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

	if err == nil && user != nil && user.InterfaceLang != "" {
		localizer_lib.LoadLang(user.InterfaceLang)
	} else {
		localizer_lib.LoadLang(defaultInterfaceLang)
	}

	return user
}

func (s *userService) Upsert(user *model.User) error {
	if existing, err := s.repository.First(user.ChatId); err == nil && existing != nil {
		return s.repository.Update(user)
	}

	return s.repository.Create(user)
}

func (s *userService) GetUserFromContext(ctx context.Context) (*model.User, error) {
	user, ok := ctx.Value("user").(*model.User)
	if !ok || user == nil {
		return nil, fmt.Errorf("user not found in context")
	}
	return user, nil
}
