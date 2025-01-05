package handlers

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"langs/internal/consts"
	"langs/internal/model"
	"langs/pkg/language_detector"
	"langs/pkg/wordTranslator"
	"strconv"
	"strings"
)

func (h *TGHandler) processNewWord(ctx context.Context, b *bot.Bot, update *models.Update, user *model.User) {
	newWord := strings.ToLower(strings.TrimSpace(update.Message.Text))

	fromLang, err := language_detector.Detect(newWord, user.GetUserLangs())
	if err != nil {
		h.handleError(ctx, b, update.Message.Chat.ID, "Language detection failed")
		return
	}

	langTo := user.GetOppositeLang(fromLang)

	if len(newWord) >= 100 {
		_, err := h.readerService.AddLongRead(newWord, getFirstWord(newWord), user)
		if err != nil {
			h.handleError(ctx, b, update.Message.Chat.ID, "Long read failed")
			return
		}
		h.tgMessage.SendOrEditMessage(ctx, user.ChatId, 0, "Saved to reader", nil)

		return
	}

	translation, err := wordTranslator.Translate(newWord, fromLang, langTo)
	if err != nil {
		h.handleError(ctx, b, update.Message.Chat.ID, "Translation failed")
		return
	}

	h.processTranslation(ctx, b, newWord, fromLang, update, translation)
}

func (h *TGHandler) HandleWordView(
	ctx context.Context,
	b *bot.Bot,
	mes models.MaybeInaccessibleMessage,
	data []byte,
) {
	wordIdStr := string(data)
	wordId, _ := strconv.Atoi(wordIdStr)

	chatId := mes.Message.Chat.ID
	word, err := h.wordRepository.First(int64(wordId))

	if err != nil {
		return
	}

	wordKeyboard := h.tgKeyboard.GetWordViewKeyboard(
		b,
		wordId,
		h.handleDeleteWord,
	)
	user, _ := h.userRepository.First(chatId)

	msgText := consts.LangFlags[word.ValueLang] + " " +
		word.Value + "\n" +
		consts.LangFlags[word.TranslationLang] + " " +
		word.Translation + "\n" +
		fmt.Sprintf("(%d/%d)", word.Rate, user.TargetRate)
	h.tgMessage.SendOrEditMessage(ctx, chatId, 0, msgText, wordKeyboard)
}

func (h *TGHandler) handleDeleteWord(
	ctx context.Context,
	b *bot.Bot,
	mes models.MaybeInaccessibleMessage,
	data []byte,
) {
	wordIdStr := string(data)
	wordId, _ := strconv.Atoi(wordIdStr)

	chatId := mes.Message.Chat.ID
	word, err := h.wordRepository.First(int64(wordId))

	if err != nil {
		return
	}

	err = h.wordRepository.Delete(int64(wordId))

	if err != nil {
		return
	}

	msgText := word.Value + " Deleted"

	h.tgMessage.SendOrEditMessage(ctx, chatId, mes.Message.ID, msgText, nil)
}

func getFirstWord(s string) string {
	words := strings.Fields(s)
	if len(words) > 0 {
		return words[0]
	}
	return ""
}
