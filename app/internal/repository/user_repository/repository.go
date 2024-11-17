package user_repository

import (
	"gorm.io/gorm"
	"langs/internal/interfaces"
	"langs/internal/model"
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

		if err != nil {
			return nil, err
		}

		if dbUser.ChatID != 0 {
			return dbUser, nil
		}
	}

	newUser := &model.User{ChatID: chatId}
	err := r.Create(newUser)

	return newUser, err
}

func (r *userRepository) Update(user *model.User) error {
	return r.db.Save(&user).Error
}

func (r *userRepository) GetAllByInterval(interval uint16) ([]*model.User, error) {
	var users []*model.User

	//return users, r.db.Where("test_interval = ?", interval).Find(&users).Error
	return users, r.db.Find(&users).Error
}

func (r *userRepository) GetBookPart(id int64) (*model.BookPart, error) {
	var bookPart model.BookPart

	//return users, r.db.Where("test_interval = ?", interval).Find(&users).Error
	return &bookPart, r.db.Where("id = ?", id).Find(&bookPart).Error
}

func (r *userRepository) GetAllChatIDs() ([]int64, error) {
	var chatIDs []int64

	err := r.db.Model(&model.User{}).Pluck("chat_id", &chatIDs).Error
	if err != nil {
		return nil, err
	}

	return chatIDs, nil
}
