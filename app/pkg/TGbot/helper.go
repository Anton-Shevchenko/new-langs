package TGbot

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func SendPrompt(ctx context.Context, b *bot.Bot, msg *models.Message, text string, markup interface{}) {
	if len(msg.Entities) > 0 && msg.Entities[0].Type == models.MessageEntityTypeBotCommand {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      msg.Chat.ID,
			Text:        text,
			ReplyMarkup: markup,
		})

		if err != nil {
			fmt.Printf("Error sending message: %v\n", err)
		}
		return
	}

	_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      msg.Chat.ID,
		Text:        text,
		ReplyMarkup: markup,
		MessageID:   msg.ID,
	})
	if err != nil {
		fmt.Printf("Error sending message: %v\n", err)
	}
}

func SendMessage(ctx context.Context, b *bot.Bot, chatID int64, text string, markup interface{}) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
		LinkPreviewOptions: &models.LinkPreviewOptions{
			IsDisabled: bot.True(),
		},
		ReplyMarkup: markup,
	})
	if err != nil {
		return
	}
}

func GetChatIDFromUpdate(update *models.Update) int64 {
	if update.Message != nil {
		return update.Message.Chat.ID
	}
	return update.CallbackQuery.Message.Message.Chat.ID
}
