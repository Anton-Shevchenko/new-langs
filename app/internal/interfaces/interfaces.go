package interfaces

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/keyboard/inline"
	"langs/internal/model"
)

type UserService interface {
	Upsert(user *model.User) error
	InitUser(chatId int64) *model.User
	GetUserFromContext(ctx context.Context) (*model.User, error)
}

type UserRepository interface {
	First(chatId int64) (*model.User, error)
	FirstOrCreate(chatId int64) (*model.User, error)
	Update(user *model.User) error
	Create(user *model.User) error
	GetAllByInterval(interval uint16) ([]*model.User, error)
	GetAllChatIDs() ([]int64, error)

	GetBookPart(id int64) (*model.BookPart, error)
}

type BookRepository interface {
	Find(id int64) (*model.Book, error)
	Create(book *model.Book) error
	AllByChatId(chatId int64) ([]*model.Book, error)
	Update(book *model.Book) error
}

type WordRepository interface {
	Create(word *model.Word) error
	AllByChatId(chatId int64, limit, offset int) ([]*model.Word, error)
	GetRandomWordByChatId(chatId int64) (*model.Word, error)
	GetRandomTranslationsByChatId(chatId int64, exception, lang string, limit int) ([]string, error)
	CheckSimilarWord(word *model.Word) (*model.Word, error)
	Delete(id int64) error
	GetCountByChatId(chatId int64) int64
}

type MessengerService interface {
	SendWordTest(chatId int64, handle inline.OnSelect, word string, translations []string)
}

type CommandSetInterface interface {
	Start(ctx context.Context, b *bot.Bot, update *models.Update)
	AskNativeLang(ctx context.Context, b *bot.Bot, update *models.Update)
	AskLangToStudy(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage)
}
