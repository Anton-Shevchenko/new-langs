package handlers

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"langs/internal/domain"
	TGbot "langs/internal/infrastructure/platform/telegram/helper"
	"langs/internal/infrastructure/platform/telegram/tg_keyboard"
	"langs/internal/infrastructure/platform/telegram/tg_msg"
	"langs/internal/interfaces"
	"langs/internal/usecase"
	"langs/pkg/nlp/wordTranslator"
	"strings"
)

type AppRouterInterface interface {
	DefaultHandler(ctx context.Context, b *bot.Bot, update *models.Update)
	handleText(ctx context.Context, b *bot.Bot, update *models.Update, user *model.User)
	getUserFromContext(ctx context.Context, b *bot.Bot, update *models.Update) *model.User
	getUserFromContextMsg(ctx context.Context, b *bot.Bot, msg models.MaybeInaccessibleMessage) *model.User
	processCustomTranslation(ctx context.Context, b *bot.Bot, update *models.Update, user *model.User)
	processNewWord(ctx context.Context, b *bot.Bot, update *models.Update, user *model.User)
	processTranslation(ctx context.Context, b *bot.Bot, sourceWord string, sourceWordLang string, update *models.Update, translation *wordTranslator.TranslateResult)
	handleError(ctx context.Context, b *bot.Bot, chatID int64, errorMsg string)
	handleSpellingMistakes(ctx context.Context, b *bot.Bot, update *models.Update, word string, lang string) bool
	handleReplacement(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte)
	handleCustomTranslation(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte)
	handleDocument(ctx context.Context, b *bot.Bot, update *models.Update)
	OnBack(ctx context.Context, b *bot.Bot, update *models.Update)
	OnBook(ctx context.Context, b *bot.Bot, update *models.Update)
	OnWordList(ctx context.Context, b *bot.Bot, update *models.Update)
	OnWordSearch(ctx context.Context, b *bot.Bot, update *models.Update)
	processWordSearch(ctx context.Context, b *bot.Bot, update *models.Update, user *model.User)
	WordList(ctx context.Context, b *bot.Bot, chatId int64, page, msgId int)
	OnSelectTranslateOption(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, translation []byte)
	HandleWordView(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte)
	HandlePaginationListing(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte)
	HandleBooks(ctx context.Context, b *bot.Bot, update *models.Update)
	HandleBook(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte)
	OnTestMe(ctx context.Context, b *bot.Bot, update *models.Update)
	OnTestAnswer(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, answerDataBytes []byte)
	OnWriteTestAnswer(ctx context.Context, b *bot.Bot, update *models.Update, user *model.User)
	OnWriteTestOption(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte)
	handleDeleteWord(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte)
	handleBookNavigation(ctx context.Context, b *bot.Bot, update *models.Update, navFunc func(userID int64))
	handleBookNext(ctx context.Context, b *bot.Bot, update *models.Update)
	handleBookPrev(ctx context.Context, b *bot.Bot, update *models.Update)
	handleBookReset(ctx context.Context, b *bot.Bot, update *models.Update)
	OnSettings(ctx context.Context, b *bot.Bot, update *models.Update)
	HandleSettingsCallback(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte)
	processTimezoneInput(ctx context.Context, b *bot.Bot, update *models.Update, user *model.User)
	processStartTimeInput(ctx context.Context, b *bot.Bot, update *models.Update, user *model.User)
	processEndTimeInput(ctx context.Context, b *bot.Bot, update *models.Update, user *model.User)
	HandleSettingsCallbacks(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte)
	HandleLangChange(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte)
	handleCallback(ctx context.Context, b *bot.Bot, update *models.Update)
	OnMainMenu(ctx context.Context, b *bot.Bot, update *models.Update)
	OnSecondaryMenu(ctx context.Context, b *bot.Bot, update *models.Update)
	OnExport(ctx context.Context, b *bot.Bot, update *models.Update)
	HandleExportCallback(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte)
	SetCommandSet(set interfaces.CommandSetInterface)
}

type AppRouter struct {
	tgKeyboard     *tg_keyboard.TGKeyboard
	tgMessage      *tg_msg.TGMessageService
	wordService    *service.WordService
	wordRepository interfaces.WordRepository
	userRepository interfaces.UserRepository
	userService    interfaces.UserService
	readerService  *service.ReaderService
	commandSet     interfaces.CommandSetInterface
}

func NewAppRouter(
	tgKeyboard *tg_keyboard.TGKeyboard,
	tgMessage *tg_msg.TGMessageService,
	wordService *service.WordService,
	wordRepository interfaces.WordRepository,
	userRepository interfaces.UserRepository,
	userService interfaces.UserService,
	readerService *service.ReaderService,
) *AppRouter {
	return &AppRouter{
		tgKeyboard:     tgKeyboard,
		tgMessage:      tgMessage,
		wordService:    wordService,
		wordRepository: wordRepository,
		userRepository: userRepository,
		userService:    userService,
		readerService:  readerService,
	}
}

func (h *AppRouter) SetCommandSet(set interfaces.CommandSetInterface) {
	h.commandSet = set
}

func (h *AppRouter) DefaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil {
		return
	}

	if update.CallbackQuery != nil {
		h.handleCallback(ctx, b, update)
		return
	}

	if update.Message == nil {
		return
	}

	if update.Message.Document != nil {
		h.handleDocument(ctx, b, update)
		return
	}

	user := h.getUserFromContext(ctx, b, update)

	if user == nil {
		fmt.Println("No user")
		return
	}

	langs := user.GetUserLangs()

	if len(langs) < 2 || (len(langs) > 1 && (langs[0] == "" || langs[1] == "")) {
		h.commandSet.Start(ctx, b, update)
		return
	}

	h.handleText(ctx, b, update, user)
}

func (h *AppRouter) handleCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil || len(update.CallbackQuery.Data) == 0 {
		return
	}

	callbackData := string(update.CallbackQuery.Data)
	mes := models.MaybeInaccessibleMessage{
		Message: update.CallbackQuery.Message.Message,
	}

	if strings.HasPrefix(callbackData, "settings_") ||
		strings.HasPrefix(callbackData, "timezone_") ||
		strings.HasPrefix(callbackData, "interval_") ||
		strings.HasPrefix(callbackData, "day_") ||
		strings.HasPrefix(callbackData, setNativeLangPrefix) ||
		strings.HasPrefix(callbackData, setTargetLangPrefix) ||
		strings.HasPrefix(callbackData, setIfaceLangPrefix) ||
		callbackData == "days_all" ||
		callbackData == "days_weekdays" ||
		callbackData == "days_weekend" {
		h.HandleSettingsCallbacks(ctx, b, mes, []byte(update.CallbackQuery.Data))
		return
	}

	if strings.HasPrefix(callbackData, "word_") {
		h.HandleWordView(ctx, b, mes, []byte(update.CallbackQuery.Data))
		return
	}

	if strings.HasPrefix(callbackData, "list_") {
		h.HandlePaginationListing(ctx, b, mes, []byte(update.CallbackQuery.Data))
		return
	}

	if strings.HasPrefix(callbackData, "book_") {
		h.HandleBook(ctx, b, mes, []byte(update.CallbackQuery.Data))
		return
	}

	if strings.HasPrefix(callbackData, exportCallbackPrefix) {
		h.HandleExportCallback(ctx, b, mes, []byte(update.CallbackQuery.Data))
		return
	}

	if strings.HasPrefix(callbackData, "test_") {
		h.OnTestAnswer(ctx, b, mes, []byte(update.CallbackQuery.Data))
		return
	}

	if strings.HasPrefix(callbackData, "write_") {
		h.OnWriteTestOption(ctx, b, mes, []byte(update.CallbackQuery.Data))
		return
	}

	if strings.HasPrefix(callbackData, "translate_") {
		h.OnSelectTranslateOption(ctx, b, mes, []byte(update.CallbackQuery.Data))
		return
	}

	if strings.HasPrefix(callbackData, "custom_") {
		h.handleCustomTranslation(ctx, b, mes, []byte(update.CallbackQuery.Data))
		return
	}

	if strings.HasPrefix(callbackData, "replacement_") {
		h.handleReplacement(ctx, b, mes, []byte(update.CallbackQuery.Data))
		return
	}
}

func (h *AppRouter) handleError(ctx context.Context, b *bot.Bot, chatID int64, errorMsg string) {
	TGbot.SendMessage(ctx, b, chatID, "Error occurred: "+errorMsg, nil)
}

func (h *AppRouter) handleText(ctx context.Context, b *bot.Bot, update *models.Update, user *model.User) {
	if user.IsAwaitingInput() {
		if user.StateData.Scenario == model.CustomTranslationScenario {
			h.processCustomTranslation(ctx, b, update, user)
			return
		}

		if user.StateData.Scenario == model.WritingExamScenario {
			h.OnWriteTestAnswer(ctx, b, update, user)
			return
		}

		if user.StateData.Scenario == model.WordSearchScenario {
			h.processWordSearch(ctx, b, update, user)
			return
		}

		if user.StateData.Scenario == TimezoneScenario {
			h.processTimezoneInput(ctx, b, update, user)
			return
		}

		if user.StateData.Scenario == QuietHoursStartScenario {
			h.processStartTimeInput(ctx, b, update, user)
			return
		}

		if user.StateData.Scenario == QuietHoursEndScenario {
			h.processEndTimeInput(ctx, b, update, user)
			return
		}
	}

	h.processNewWord(ctx, b, update, user)
}

func (h *AppRouter) getUserFromContext(ctx context.Context, b *bot.Bot, update *models.Update) *model.User {
	user, err := h.userService.GetUserFromContext(ctx)
	if err != nil {
		h.handleError(ctx, b, update.Message.Chat.ID, "User context is invalid")
		panic("User context is invalid")
	}
	return user
}

func (h *AppRouter) getUserFromContextMsg(ctx context.Context, b *bot.Bot, msg models.MaybeInaccessibleMessage) *model.User {
	user, err := h.userService.GetUserFromContext(ctx)

	if err != nil {
		h.handleError(ctx, b, msg.Message.Chat.ID, "User context is invalid")
		panic("User context is invalid")
	}
	return user
}
