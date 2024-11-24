package handlers

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"io"
	"langs/pkg/TGbot"
	"net/http"
	"os"
)

func (h *TGHandler) handleDocument(ctx context.Context, b *bot.Bot, update *models.Update) {
	user := h.getUserFromContext(ctx, b, update)
	document := update.Message.Document

	if document.MimeType != "text/plain" {
		TGbot.SendMessage(ctx, b, user.ChatID, "Book has to be in .txt format", nil)

	}

	file, err := b.GetFile(ctx, &bot.GetFileParams{
		FileID: document.FileID,
	})

	if err != nil {
		fmt.Println("Error getting file URL:", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: user.ChatID,
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
			ChatID: user.ChatID,
			Text:   "Не удалось скачать файл.",
		})
		return
	}

	h.readerService.AddBook(filePath, document.FileName, user)
}

func (h *TGHandler) downloadFile(filepath string, url string) error {
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
