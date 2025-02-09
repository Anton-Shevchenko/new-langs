package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"langs/internal/consts"
	"langs/internal/model"
	"langs/internal/tg_bot/helper"
)

func (h *TGHandler) OnTestMe(ctx context.Context, b *bot.Bot, update *models.Update) {
	user := h.getUserFromContext(ctx, b, update)
	h.wordService.SendTest(user, h.OnTestAnswer, h.OnWriteTestOption)
}

func (h *TGHandler) OnTestAnswer(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, answerDataBytes []byte) {
	data := helper.StringToData(string(answerDataBytes))

	wordID, err := extractInt(data, 0)
	if err != nil {
		h.handleError(ctx, b, mes.Message.Chat.ID, "Failed to parse wordID.")
		return
	}
	chosenValue, err := extractString(data, 1)
	if err != nil {
		h.handleError(ctx, b, mes.Message.Chat.ID, "Failed to parse chosen value.")
		return
	}
	word, err := h.wordRepository.First(int64(wordID))
	if err != nil {
		h.handleError(ctx, b, mes.Message.Chat.ID, "Failed to fetch word from repository.")
		return
	}
	user := h.getUserFromContextMsg(ctx, b, mes)
	correct := isCorrect(chosenValue, word.Value) || isCorrect(chosenValue, word.Translation)
	updateWordRate(h, word, correct)
	h.sendTestResultMessage(ctx, mes.Message.Chat.ID, mes.Message.ID, word, user, correct)
}

func (h *TGHandler) OnWriteTestAnswer(ctx context.Context, b *bot.Bot, update *models.Update, user *model.User) {
	wordID, err := strconv.Atoi(user.StateData.Value)
	if err != nil {
		h.handleError(ctx, b, update.Message.Chat.ID, "Failed to parse wordID.")
		return
	}
	word, err := h.wordRepository.First(int64(wordID))
	if err != nil {
		h.handleError(ctx, b, update.Message.Chat.ID, "Failed to fetch word from repository.")
		return
	}

	fmt.Println(word.Value, word.Translation, word.TranslationLang, user.NativeLang)
	targetValue := word.Value
	if user.NativeLang == word.ValueLang {
		targetValue = word.Translation
	}
	user.StateData.Scenario = ""
	user.StateData.Value = ""
	if err := h.userRepository.Update(user); err != nil {
		h.handleError(ctx, b, update.Message.Chat.ID, "Failed to update user.")
		return
	}
	userAnswer := update.Message.Text
	correct := isCorrect(userAnswer, targetValue)
	updateWordRate(h, word, correct)
	h.sendTestResultMessage(ctx, update.Message.Chat.ID, 0, word, user, correct)
}

func (h *TGHandler) OnWriteTestOption(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	wordID, err := strconv.Atoi(string(data))
	if err != nil {
		h.handleError(ctx, b, mes.Message.Chat.ID, "Failed to parse wordID.")
		return
	}
	word, err := h.wordRepository.First(int64(wordID))

	if err != nil {
		h.handleError(ctx, b, mes.Message.Chat.ID, "Failed to fetch word from repository.")
		return
	}
	user := h.getUserFromContextMsg(ctx, b, mes)
	promptWord := fmt.Sprintf("%s%s", word.Value, consts.LangFlags[word.ValueLang])
	if user.NativeLang == word.TranslationLang {
		promptWord = fmt.Sprintf("%s%s", word.Translation, consts.LangFlags[word.TranslationLang])
	}
	user.StateData.Scenario = model.WritingExamScenario
	user.StateData.Value = string(data)

	if err := h.userRepository.Update(user); err != nil {
		h.handleError(ctx, b, mes.Message.Chat.ID, "Failed to update user.")
		return
	}
	h.tgMessage.SendOrEditMessage(ctx, mes.Message.Chat.ID, 0, fmt.Sprintf("Write the translation for - %s", promptWord), nil)
}

func (h *TGHandler) sendTestResultMessage(ctx context.Context, chatID int64, messageID int, word *model.Word, user *model.User, correct bool) {
	statusEmoji := "✅"
	if !correct {
		statusEmoji = "❌"
	}
	message := fmt.Sprintf("%s %s - %s %s %s (%d/%d)", consts.LangFlags[word.ValueLang], word.Value, consts.LangFlags[word.TranslationLang], word.Translation, statusEmoji, word.Rate, user.TargetRate)
	h.tgMessage.SendOrEditMessage(ctx, chatID, messageID, message, nil)
}

func updateWordRate(h *TGHandler, word *model.Word, correct bool) {
	if correct {
		word.Rate++
	} else {
		word.Rate--
	}
	_ = h.wordRepository.Save(word)
}

func isCorrect(input, target string) bool {
	fmt.Println(strings.TrimSpace(input), strings.TrimSpace(target))
	return strings.EqualFold(strings.TrimSpace(input), strings.TrimSpace(target))
}

func extractInt(data []interface{}, index int) (int, error) {
	if index < 0 || index >= len(data) {
		return 0, fmt.Errorf("index out of range")
	}
	val, ok := data[index].(int)
	if !ok {
		return 0, fmt.Errorf("failed to cast data[%d] to int", index)
	}
	return val, nil
}

func extractString(data []interface{}, index int) (string, error) {
	if index < 0 || index >= len(data) {
		return "", fmt.Errorf("index out of range")
	}
	val, ok := data[index].(string)
	if !ok {
		return "", fmt.Errorf("failed to cast data[%d] to string", index)
	}
	return val, nil
}
