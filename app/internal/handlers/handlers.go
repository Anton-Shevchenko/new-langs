package handlers

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"langs/pkg/language_detector"
	"langs/pkg/spellio"
	"langs/pkg/wordTranslator"
	"strconv"
	"strings"
)

func (h *TGHandler) handleSpellingMistakes(ctx context.Context, b *bot.Bot, update *models.Update) bool {
	user, err := h.userService.GetUserFromContext(ctx)
	if err != nil {
		h.handleError(ctx, b, update.Message.Chat.ID, "User retrieval failed")
		return false
	}

	newWord := strings.ToLower(strings.TrimSpace(update.Message.Text))
	detectedLang, err := language_detector.Detect(newWord, user.GetUserLangs())

	if err != nil {
		return false
	}

	checkResults := spellio.Check(newWord, detectedLang)

	if checkResults != nil && checkResults.Message == spellio.SpellingMistake {
		replacementKeyboard := h.tgKeyboard.BuildReplacementKeyboard(
			b,
			update.Message.Text,
			checkResults.Replacements,
			h.handleReplacement,
		)
		h.tgMessage.SendOrEditMessage(
			ctx,
			update.Message.Chat.ID,
			0,
			"Spelling suggestions:",
			replacementKeyboard,
		)
		return true
	}

	return false
}

func (h *TGHandler) handleReplacement(
	ctx context.Context,
	b *bot.Bot,
	mes models.MaybeInaccessibleMessage,
	data []byte,
) {
	user, err := h.userService.GetUserFromContext(ctx)
	if err != nil {
		return
	}
	fromLang, err := language_detector.Detect(string(data), user.GetUserLangs())
	if err != nil {
		h.handleError(ctx, b, mes.Message.Chat.ID, "Language detection failed")
		return
	}

	langTo := user.GetOppositeLang(fromLang)
	translation, err := wordTranslator.Translate(string(data), fromLang, langTo)

	if err != nil {
		h.handleError(ctx, b, mes.Message.Chat.ID, "Error translating replacement")
		return
	}

	h.tgMessage.SendWordView(
		ctx,
		string(data),
		fromLang,
		mes.Message,
		translation,
		h.OnSelectTranslateOption,
		h.handleCustomTranslation,
	)
}

func (h *TGHandler) HandlePaginationListing(
	ctx context.Context,
	b *bot.Bot,
	mes models.MaybeInaccessibleMessage,
	data []byte,
) {
	pageData := strings.Split(string(data), ",")
	sign := pageData[0]
	pageStr := pageData[1]

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if sign == ">" {
		page++
	} else {
		if page > 0 {
			page--
		}
	}

	h.WordList(ctx, b, mes.Message.Chat.ID, page, mes.Message.ID)
}
