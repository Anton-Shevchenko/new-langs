package handlers

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"langs/internal/consts"
	"langs/internal/tg_bot/tg_keyboard"
	"langs/pkg/TGbot"
)

func (h *TGHandler) HandleBack(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Choose an option:",
		ReplyMarkup: h.tgKeyboard.InitMenuKeyboard(
			b,
			h.OnWordList,
			h.HandleBack,
			h.HandleBooks,
			h.OnTestMe,
		),
	})
}

func (h *TGHandler) OnWordList(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.WordList(ctx, b, TGbot.GetChatIDFromUpdate(update), 1, 0)
}

func (h *TGHandler) WordList(ctx context.Context, b *bot.Bot, chatId int64, page, msgId int) {
	var wordList []*tg_keyboard.ListItem
	words, _ := h.wordRepository.AllByChatId(chatId, 10, (page-1)*10)

	for _, word := range words {
		wordList = append(wordList, &tg_keyboard.ListItem{
			Id: word.ID,
			Data: word.Value + " " + consts.LangFlags[word.ValueLang] +
				" - " +
				word.Translation + " " + consts.LangFlags[word.TranslationLang],
		})
	}

	listKeyboard := h.tgKeyboard.GetListingKeyboard(
		b,
		wordList,
		h.HandleWordView,
		h.HandlePaginationListing,
		page,
	)

	h.tgMessage.SendOrEditMessage(ctx, chatId, msgId, "Word List", listKeyboard)
}
