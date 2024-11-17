package handlers

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"langs/internal/interfaces"
	"langs/internal/model"
	"langs/internal/repository/word_repository"
	"langs/internal/service"
	"langs/internal/tg_bot/tg_keyboard"
	"langs/internal/tg_bot/tg_msg"
	"langs/pkg/TGbot"
	"langs/pkg/wordTranslator"
)

type TGHandlerInterface interface {
	DefaultHandler(ctx context.Context, b *bot.Bot, update *models.Update)
	handleText(ctx context.Context, b *bot.Bot, update *models.Update)
	getUserFromContext(ctx context.Context, b *bot.Bot, update *models.Update) *model.User
	processCustomTranslation(ctx context.Context, b *bot.Bot, update *models.Update, user *model.User)
	processNewWord(ctx context.Context, b *bot.Bot, update *models.Update, user *model.User)
	processTranslation(ctx context.Context, b *bot.Bot, sourceWord string, sourceWordLang string, update *models.Update, translation *wordTranslator.TranslateResult)
	handleError(ctx context.Context, b *bot.Bot, chatID int64, errorMsg string)
	handleSpellingMistakes(ctx context.Context, b *bot.Bot, update *models.Update) bool
	handleReplacement(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte)
	handleCustomTranslation(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte)
	OnBack(ctx context.Context, b *bot.Bot, update *models.Update)
	OnBook(ctx context.Context, b *bot.Bot, update *models.Update)
	OnWordList(ctx context.Context, b *bot.Bot, update *models.Update)
	WordList(ctx context.Context, b *bot.Bot, chatId int64, page, msgId int)
	OnSelectTranslateOption(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, translation []byte)
	HandleWordView(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte)
	HandlePaginationListing(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte)
}

type TGHandler struct {
	tgKeyboard     *tg_keyboard.TGKeyboard
	tgMessage      *tg_msg.TGMessageService
	wordService    *service.WordService
	wordRepository *word_repository.WordRepository
	userRepository interfaces.UserRepository
	userService    interfaces.UserService
	readerService  *service.ReaderService
}

func NewTGHandler(
	tgKeyboard *tg_keyboard.TGKeyboard,
	tgMessage *tg_msg.TGMessageService,
	wordService *service.WordService,
	wordRepository *word_repository.WordRepository,
	userRepository interfaces.UserRepository,
	userService interfaces.UserService,
	readerService *service.ReaderService,
) *TGHandler {
	return &TGHandler{
		tgKeyboard:     tgKeyboard,
		tgMessage:      tgMessage,
		wordService:    wordService,
		wordRepository: wordRepository,
		userRepository: userRepository,
		userService:    userService,
		readerService:  readerService,
	}
}

func (h *TGHandler) DefaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message.Document != nil {
		h.handleDocument(ctx, b, update)
		return
	}

	if update.Message == nil {
		return
	}

	h.handleText(ctx, b, update)
}

func (h *TGHandler) handleError(ctx context.Context, b *bot.Bot, chatID int64, errorMsg string) {
	TGbot.SendMessage(ctx, b, chatID, "Error occurred: "+errorMsg)
}

func (h *TGHandler) handleText(ctx context.Context, b *bot.Bot, update *models.Update) {
	user := h.getUserFromContext(ctx, b, update)

	if user == nil {
		fmt.Println("No user")
		return
	}
	if user.IsAwaitingInput() {
		h.processCustomTranslation(ctx, b, update, user)
		return
	}

	h.processNewWord(ctx, b, update, user)
}

func (h *TGHandler) getUserFromContext(ctx context.Context, b *bot.Bot, update *models.Update) *model.User {
	user, err := h.userService.GetUserFromContext(ctx)
	if err != nil {
		h.handleError(ctx, b, update.Message.Chat.ID, "User context is invalid")
		panic("User context is invalid")
	}
	return user
}

func (h *TGHandler) getUserFromContextMsg(ctx context.Context, b *bot.Bot, msg models.MaybeInaccessibleMessage) *model.User {
	user, err := h.userService.GetUserFromContext(ctx)
	if err != nil {
		h.handleError(ctx, b, msg.Message.Chat.ID, "User context is invalid")
		panic("User context is invalid")
	}
	return user
}
