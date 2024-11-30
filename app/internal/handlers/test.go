package handlers

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"langs/internal/consts"
	"langs/internal/tg_bot/helper"
)

func (h *TGHandler) OnTestMe(ctx context.Context, b *bot.Bot, update *models.Update) {
	user := h.getUserFromContext(ctx, b, update)
	h.wordService.SendTest(user, h.OnTestAnswer)
}

func (h *TGHandler) OnTestAnswer(
	ctx context.Context,
	b *bot.Bot,
	mes models.MaybeInaccessibleMessage,
	answer []byte,
) {
	answerData := helper.StringToData(string(answer))
	wordID, ok := extractInt(answerData, 0, "answerData[0] is not an int")
	if !ok {
		return
	}

	word, err := h.wordRepository.First(int64(wordID))
	if err != nil {
		h.handleError(ctx, b, mes.Message.Chat.ID, "Failed to fetch word from repository.")
		return
	}

	valueString, ok := extractString(answerData, 1, "answerData[1] is not a string")
	if !ok {
		return
	}

	user := h.getUserFromContextMsg(ctx, b, mes)

	correct := word.Value == valueString || word.Translation == valueString
	if correct {
		word.Rate++
	} else {
		word.Rate++
	}

	h.wordRepository.Save(word)

	statusEmoji := "✅"
	if !correct {
		statusEmoji = "❌"
	}

	message := fmt.Sprintf(
		"%s %s - %s %s %s (%d/%d)",
		consts.LangFlags[word.ValueLang],
		word.Value,
		consts.LangFlags[word.TranslationLang],
		word.Translation,
		statusEmoji,
		word.Rate,
		user.TargetRate,
	)

	h.tgMessage.SendOrEditMessage(
		ctx,
		mes.Message.Chat.ID,
		mes.Message.ID,
		message,
		nil,
	)
}

func extractInt(data []interface{}, index int, errorMsg string) (int, bool) {
	if index >= len(data) {
		return 0, false
	}
	value, ok := data[index].(int)
	if !ok {
		panic(errorMsg)
	}
	return value, true
}

func extractString(data []interface{}, index int, errorMsg string) (string, bool) {
	if index >= len(data) {
		return "", false
	}
	value, ok := data[index].(string)
	if !ok {
		panic(errorMsg)
	}
	return value, true
}
