package user_repository

import (
	"errors"

	"gorm.io/gorm"
	"langs/internal/domain"
	"langs/internal/interfaces"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) interfaces.UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) First(chatId int64) (*model.User, error) {
	var user model.User
	err := r.db.Where("chat_id = ?", chatId).First(&user).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(&user).Error
}

func (r *userRepository) FirstOrCreate(chatId int64) (*model.User, error) {
	if chatId != 0 {
		dbUser, err := r.First(chatId)

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		if dbUser != nil && dbUser.ChatId != 0 {
			return dbUser, nil
		}
	}

	newUser := &model.User{ChatId: chatId}
	err := r.Create(newUser)

	return newUser, err
}

func (r *userRepository) Update(user *model.User) error {
	return r.db.Where("chat_id = ?", user.ChatId).Save(&user).Error
}

func (r *userRepository) GetAll() ([]*model.User, error) {
	var users []*model.User

	return users, r.db.Find(&users).Error
}

func (r *userRepository) GetAllChatIDs() ([]int64, error) {
	var chatIDs []int64

	err := r.db.Model(&model.User{}).Pluck("chat_id", &chatIDs).Error
	if err != nil {
		return nil, err
	}

	return chatIDs, nil
}
