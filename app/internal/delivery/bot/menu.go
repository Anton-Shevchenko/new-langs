package handlers

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"langs/internal/consts"
	TGbot "langs/internal/infrastructure/platform/telegram/helper"
	"langs/internal/infrastructure/platform/telegram/tg_keyboard"
)

func (h *AppRouter) OnBack(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.OnMainMenu(ctx, b, update)
}

func (h *AppRouter) OnMainMenu(ctx context.Context, b *bot.Bot, update *models.Update) {
	var chatID int64
	if update.Message != nil {
		chatID = update.Message.Chat.ID
	} else if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Message.Chat.ID
	}

	keyboard := h.tgKeyboard.InitMainMenuKeyboard(
		b,
		h.OnWordList,
		h.OnTestMe,
		h.OnSecondaryMenu,
	)

	TGbot.SendMessage(ctx, b, chatID, "Main Menu - Choose an option:", keyboard)
}

func (h *AppRouter) OnSecondaryMenu(ctx context.Context, b *bot.Bot, update *models.Update) {
	var chatID int64
	if update.Message != nil {
		chatID = update.Message.Chat.ID
	} else if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Message.Chat.ID
	}

	keyboard := h.tgKeyboard.InitSecondaryMenuKeyboard(
		b,
		h.OnMainMenu,
		h.OnBook,
		h.OnSettings,
		h.OnWordSearch,
		h.OnExport,
	)

	TGbot.SendMessage(ctx, b, chatID, "More Options - Choose an option:", keyboard)
}

func (h *AppRouter) OnWordList(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.WordList(ctx, b, TGbot.GetChatIDFromUpdate(update), 1, 0)
}

const wordsPerPage = 10

func (h *AppRouter) WordList(ctx context.Context, b *bot.Bot, chatId int64, page, msgId int) {
	totalWords := h.wordRepository.GetCountByChatId(chatId)
	if totalWords == 0 {
		h.tgMessage.SendOrEditMessage(ctx, chatId, msgId, "Word list is empty. Send me a word to add it.", nil)
		return
	}

	totalPages := int((totalWords + wordsPerPage - 1) / wordsPerPage)
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	var wordList []*tg_keyboard.ListItem
	words, _ := h.wordRepository.AllByChatId(chatId, wordsPerPage, (page-1)*wordsPerPage)

	for _, word := range words {
		wordList = append(wordList, &tg_keyboard.ListItem{
			Id:   word.ID,
			Data: word.Value + " " + consts.LangFlags[word.ValueLang] + " - " + word.Translation + " " + consts.LangFlags[word.TranslationLang],
		})
	}

	listKeyboard := h.tgKeyboard.GetListingKeyboard(
		b,
		wordList,
		h.HandleWordView,
		h.HandlePaginationListing,
		page,
		totalPages,
	)

	h.tgMessage.SendOrEditMessage(ctx, chatId, msgId, fmt.Sprintf("Word List (%d)", totalWords), listKeyboard)
}

func (h *AppRouter) OnBook(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.HandleBooks(ctx, b, update)
}
