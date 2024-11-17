package handlers

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"langs/internal/tg_bot/tg_keyboard"
	"strconv"
)

func (h *TGHandler) HandleBooks(ctx context.Context, b *bot.Bot, update *models.Update) {
	var listItems []*tg_keyboard.ListItem
	user := h.getUserFromContext(ctx, b, update)
	books, err := h.readerService.GetBooks(user)

	if err != nil {
		return
	}

	for _, book := range books {
		listItems = append(listItems, &tg_keyboard.ListItem{Id: book.ID, Data: book.Name})
	}

	booksKeyboard := h.tgKeyboard.GetListingKeyboard(
		b,
		listItems,
		h.HandleBook,
		h.HandlePaginationListing,
		0,
	)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Your books",
		ReplyMarkup: booksKeyboard,
	})
}

func (h *TGHandler) HandleBook(
	ctx context.Context,
	b *bot.Bot,
	mes models.MaybeInaccessibleMessage,
	data []byte,
) {
	user := h.getUserFromContextMsg(ctx, b, mes)

	bookId, err := strconv.Atoi(string(data))
	user.BookId = int64(bookId)
	h.userRepository.Update(user)
	if err != nil {
		return
	}
	bookPart, err := h.readerService.ReadBookPart(int64(bookId))

	if err != nil {
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: user.ChatID,
		Text:   bookPart.Text,
		ReplyMarkup: h.tgKeyboard.InitReaderKeyboard(
			b,
			h.handleBookNext,
			h.handleBookPrev,
			h.handleBookReset,
			h.HandleBack,
		),
	})
}

func (h *TGHandler) handleBookNavigation(ctx context.Context, b *bot.Bot, update *models.Update, navFunc func(userID int64)) {
	user := h.getUserFromContext(ctx, b, update)
	//TODO
	if user.BookId == 0 {
		user.BookID = 1
	}
	navFunc(user.BookID)

	bookPart, err := h.readerService.ReadBookPart(user.BookID)
	if err != nil {
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: user.ChatID,
		Text:   bookPart.Text,
		ReplyMarkup: h.tgKeyboard.InitReaderKeyboard(
			b,
			h.handleBookNext,
			h.handleBookPrev,
			h.handleBookReset,
			h.HandleBack,
		),
	})
}

func (h *TGHandler) handleBookNext(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.handleBookNavigation(ctx, b, update, h.readerService.NextPage)
}

func (h *TGHandler) handleBookPrev(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.handleBookNavigation(ctx, b, update, h.readerService.PrevPage)
}

func (h *TGHandler) handleBookReset(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.handleBookNavigation(ctx, b, update, func(userID int64) {})
}
