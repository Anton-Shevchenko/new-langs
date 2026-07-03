package handlers

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"langs/internal/domain"
	TGbot "langs/internal/infrastructure/platform/telegram/helper"
	"langs/internal/presentation/formatter"
	"langs/pkg/nlp/wordTranslator"
	"regexp"
	"strings"
)

func (h *AppRouter) processCustomTranslation(
	ctx context.Context,
	b *bot.Bot,
	update *models.Update,
	user *model.User,
) {
	sourceWord := user.StateData.Value
	user.StateData.Scenario = ""
	user.StateData.Value = ""

	if err := h.userRepository.Update(user); err != nil {
		h.handleError(ctx, b, update.Message.Chat.ID, "User update error")
		return
	}

	customTranslation := strings.TrimSpace(update.Message.Text)

	collisionWord, err := h.wordService.AddWord(sourceWord, customTranslation, user)
	if err != nil || (collisionWord != nil && collisionWord.ID != 0) {
		h.handleError(ctx, b, update.Message.Chat.ID, "You already have this word")
		return
	}

	h.tgMessage.SendFormattedWordMessage(ctx, update.Message.Chat.ID, sourceWord, customTranslation)
}

func (h *AppRouter) processTranslation(
	ctx context.Context,
	b *bot.Bot,
	sourceWord string,
	sourceWordLang string,
	update *models.Update,
	translation *wordTranslator.TranslateResult,
) {
	if len(strings.Fields(update.Message.Text)) == 1 {
		// Google Translate silently "corrects" misspellings and still returns a
		// translation with a high confidence score, so neither IsValid nor the
		// confidence can be trusted to catch typos. The one reliable Google
		// signal is RecognizedWord (a real dictionary entry): when it is set
		// the word is definitely not a typo, so we skip the spell check and
		// avoid an extra external request. Otherwise we ask LanguageTool, which
		// is the source of truth across all languages; if it finds a mistake it
		// shows suggestions and we stop here.
		if !translation.RecognizedWord && h.handleSpellingMistakes(ctx, b, update) {
			return
		}

		if !translation.IsValid {
			ts := translation.SimpleString()
			TGbot.SendMessage(ctx, b, update.Message.Chat.ID, ts, nil)
			return
		}
	}

	h.tgMessage.SendWordView(
		ctx,
		sourceWord,
		sourceWordLang,
		update.Message,
		translation,
		h.OnSelectTranslateOption,
		h.handleCustomTranslation,
	)
}

func (h *AppRouter) handleCustomTranslation(
	ctx context.Context,
	b *bot.Bot,
	mes models.MaybeInaccessibleMessage,
	data []byte,
) {
	user, err := h.userService.GetUserFromContext(ctx)
	if err != nil {
		h.handleError(ctx, b, mes.Message.Chat.ID, "User retrieval failed")
		return
	}

	user.StateData.Value = string(data)
	user.StateData.Scenario = model.CustomTranslationScenario
	err = h.userRepository.Update(user)
	if err != nil {
		h.handleError(ctx, b, mes.Message.Chat.ID, "Failed to update user data")
		return
	}

	h.tgMessage.SendOrEditMessage(ctx, user.ChatId, 0, "Write your translation", nil)
}

func (h *AppRouter) OnSelectTranslateOption(
	ctx context.Context,
	b *bot.Bot,
	mes models.MaybeInaccessibleMessage,
	translation []byte,
) {
	user, err := h.userService.GetUserFromContext(ctx)
	if err != nil {
		h.handleError(ctx, b, mes.Message.Chat.ID, "User retrieval failed")
		return
	}

	sourceWord := wordTranslator.ParseSourceWordsFromTranslateMsg(mes.Message.Text)

	if !hasGermanArticle(sourceWord) {
		if article := wordTranslator.ParseArticleFromTranslateMsg(mes.Message.Text); article != "" {
			sourceWord = article + " " + sourceWord
		}
	}

	msgText := formatter.FormatWordMessage(sourceWord, string(translation))
	chatId := mes.Message.Chat.ID

	collisionWord, err := h.wordService.AddWord(sourceWord, string(translation), user)

	if err != nil {
		h.handleError(ctx, b, chatId, "Error adding word")
		return
	}

	if collisionWord != nil {
		h.tgMessage.SendOrEditMessage(ctx, chatId, 0, "You have this word", nil)
		return
	}

	h.tgMessage.SendOrEditMessage(ctx, chatId, 0, msgText, nil)

}

var articleRe = regexp.MustCompile(`(?i)^(der|die|das)\s+(.+)$`)

func splitGermanArticle(s string) (article, rest string, ok bool) {
	s = strings.TrimSpace(s)
	if m := articleRe.FindStringSubmatch(s); len(m) == 3 {
		return strings.ToLower(m[1]), m[2], true
	}
	return "", s, false
}

func hasGermanArticle(s string) bool {
	_, _, ok := splitGermanArticle(s)
	return ok
}
