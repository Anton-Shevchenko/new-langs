package handlers

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"io"
	TGbot "langs/internal/infrastructure/platform/telegram/helper"
	"net/http"
	"os"
)

func (h *AppRouter) handleDocument(ctx context.Context, b *bot.Bot, update *models.Update) {
	user := h.getUserFromContext(ctx, b, update)
	document := update.Message.Document

	if document.MimeType != "text/plain" {
		TGbot.SendMessage(ctx, b, user.ChatId, "Book has to be in .txt format", nil)
		return
	}

	file, err := b.GetFile(ctx, &bot.GetFileParams{
		FileID: document.FileID,
	})

	if err != nil {
		fmt.Println("Error getting file URL:", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: user.ChatId,
			Text:   err.Error(),
		})
		return
	}
	fileURL := b.FileDownloadLink(file)
	filePath := "downloaded_" + document.FileName
	err = h.downloadFile(filePath, fileURL)
	if err != nil {
		fmt.Println("Error downloading file:", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: user.ChatId,
			Text:   "Не удалось скачать файл.",
		})
		return
	}

	book, err := h.readerService.AddBook(filePath, document.FileName, user)
	if err != nil {
		h.handleError(ctx, b, user.ChatId, err.Error())
		return
	}

	TGbot.SendMessage(ctx, b, user.ChatId, fmt.Sprintf("Saved \"%s\" to reader.", book.Name), nil)
}

func (h *AppRouter) downloadFile(filepath string, url string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
