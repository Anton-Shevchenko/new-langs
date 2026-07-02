package handlers

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"langs/internal/infrastructure/platform/telegram/tg_keyboard"
	"strconv"
)

func (h *AppRouter) HandleBooks(ctx context.Context, b *bot.Bot, update *models.Update) {
	var listItems []*tg_keyboard.ListItem
	user := h.getUserFromContext(ctx, b, update)
	books, err := h.readerService.GetBooks(user)

	if err != nil {
		h.handleError(ctx, b, user.ChatId, err.Error())
		return
	}

	if len(books) == 0 {
		h.tgMessage.SendOrEditMessage(ctx, user.ChatId, 0, "Your reader is empty. Send me a .txt file or a long text to add a book.", nil)
		return
	}

	for _, book := range books {
		listItems = append(listItems, &tg_keyboard.ListItem{
			Id:   book.ID,
			Data: fmt.Sprintf("%s (%d pages)", book.Name, book.Len),
		})
	}

	booksKeyboard := h.tgKeyboard.GetListingKeyboard(
		b,
		listItems,
		h.HandleBook,
		h.HandlePaginationListing,
		1,
		1,
	)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Your books",
		ReplyMarkup: booksKeyboard,
	})
}

func (h *AppRouter) HandleBook(
	ctx context.Context,
	b *bot.Bot,
	mes models.MaybeInaccessibleMessage,
	data []byte,
) {
	user := h.getUserFromContextMsg(ctx, b, mes)

	bookId, err := strconv.Atoi(string(data))
	if err != nil {
		h.handleError(ctx, b, user.ChatId, err.Error())
		return
	}

	user.BookId = int64(bookId)
	if err := h.userRepository.Update(user); err != nil {
		h.handleError(ctx, b, user.ChatId, err.Error())
		return
	}

	bookPart, err := h.readerService.ReadBookPart(int64(bookId))

	if err != nil {
		h.handleError(ctx, b, user.ChatId, err.Error())
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: user.ChatId,
		Text:   bookPart.Text,
		ReplyMarkup: h.tgKeyboard.InitReaderKeyboard(
			b,
			h.handleBookNext,
			h.handleBookPrev,
			h.handleBookReset,
			h.OnBack,
		),
	})
}

func (h *AppRouter) handleBookNavigation(ctx context.Context, b *bot.Bot, update *models.Update, navFunc func(userID int64)) {
	user := h.getUserFromContext(ctx, b, update)
	if user.BookId == 0 {
		h.HandleBooks(ctx, b, update)
		return
	}

	navFunc(user.BookId)

	bookPart, err := h.readerService.ReadBookPart(user.BookId)
	if err != nil {
		h.handleError(ctx, b, user.ChatId, err.Error())
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: user.ChatId,
		Text:   bookPart.Text,
		ReplyMarkup: h.tgKeyboard.InitReaderKeyboard(
			b,
			h.handleBookNext,
			h.handleBookPrev,
			h.handleBookReset,
			h.OnBack,
		),
	})
}

func (h *AppRouter) handleBookNext(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.handleBookNavigation(ctx, b, update, h.readerService.NextPage)
}

func (h *AppRouter) handleBookPrev(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.handleBookNavigation(ctx, b, update, h.readerService.PrevPage)
}

func (h *AppRouter) handleBookReset(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.handleBookNavigation(ctx, b, update, func(userID int64) {})
}
