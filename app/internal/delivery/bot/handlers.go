package handlers

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"langs/internal/domain"
	"langs/pkg/nlp/language_detector"
	"langs/pkg/nlp/spellio"
	"langs/pkg/nlp/wordTranslator"
	"strconv"
	"strings"
)

// handleSpellingMistakes checks the given word (in the already-detected
// language) for spelling mistakes and, if any are found, shows replacement
// suggestions. The word and language are passed in from the caller, which has
// already detected them, to avoid a redundant (and expensive) language
// detection pass.
func (h *AppRouter) handleSpellingMistakes(ctx context.Context, b *bot.Bot, update *models.Update, word, lang string) bool {
	checkResults := spellio.Check(word, lang)

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

func (h *AppRouter) handleReplacement(
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

func (h *AppRouter) HandlePaginationListing(
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

	switch sign {
	case ">":
		page++
	case "<":
		page--
	default:
		// Page indicator button, nothing to do.
		return
	}

	h.WordList(ctx, b, mes.Message.Chat.ID, page, mes.Message.ID)
}

func (h *AppRouter) HandleSettingsCallbacks(
	ctx context.Context,
	b *bot.Bot,
	mes models.MaybeInaccessibleMessage,
	data []byte,
) {
	callbackData := string(data)

	if strings.HasPrefix(callbackData, "settings_") {
		h.HandleSettingsCallback(ctx, b, mes, data)
		return
	}

	if strings.HasPrefix(callbackData, setNativeLangPrefix) ||
		strings.HasPrefix(callbackData, setTargetLangPrefix) ||
		strings.HasPrefix(callbackData, setIfaceLangPrefix) {
		h.HandleLangChange(ctx, b, mes, data)
		return
	}

	if strings.HasPrefix(callbackData, "timezone_") {
		timezone := strings.TrimPrefix(callbackData, "timezone_")
		user := h.getUserFromContextMsg(ctx, b, mes)
		user.Timezone = timezone
		user.StateData.Scenario = ""
		h.userRepository.Update(user)

		h.sendSettingsMenu(ctx, b, mes.Message.Chat.ID, user)
		return
	}

	if strings.HasPrefix(callbackData, "interval_") {
		hours, err := strconv.Atoi(strings.TrimPrefix(callbackData, "interval_"))
		if err != nil || hours < model.MinTestIntervalHours || hours > model.MaxTestIntervalHours {
			return
		}
		user := h.getUserFromContextMsg(ctx, b, mes)
		user.TestInterval = uint16(hours)
		h.userRepository.Update(user)

		h.sendSettingsMenu(ctx, b, mes.Message.Chat.ID, user)
		return
	}

	if strings.HasPrefix(callbackData, "day_") {
		day, err := strconv.Atoi(strings.TrimPrefix(callbackData, "day_"))
		if err != nil {
			return
		}
		user := h.getUserFromContextMsg(ctx, b, mes)
		user.QuietHours.ToggleDay(day)
		h.userRepository.Update(user)

		h.sendDaysKeyboard(ctx, mes.Message.Chat.ID, mes.Message.ID, user)
		return
	}

	daysPresets := map[string]string{
		"days_all":      "1,2,3,4,5,6,7",
		"days_weekdays": "1,2,3,4,5",
		"days_weekend":  "6,7",
	}
	if preset, ok := daysPresets[callbackData]; ok {
		user := h.getUserFromContextMsg(ctx, b, mes)
		user.QuietHours.DaysOfWeek = preset
		h.userRepository.Update(user)

		h.sendDaysKeyboard(ctx, mes.Message.Chat.ID, mes.Message.ID, user)
		return
	}
}
