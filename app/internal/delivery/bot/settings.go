package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"langs/internal/consts"
	"langs/internal/domain"
	TGbot "langs/internal/infrastructure/platform/telegram/helper"
	"langs/pkg/nlp/localizer_lib"
)

const (
	TimezoneScenario        = "timezone"
	QuietHoursStartScenario = "quiet_hours_start"
	QuietHoursEndScenario   = "quiet_hours_end"
)

func (h *AppRouter) OnSettings(ctx context.Context, b *bot.Bot, update *models.Update) {
	user := h.getUserFromContext(ctx, b, update)
	h.sendSettingsMenu(ctx, b, TGbot.GetChatIDFromUpdate(update), user)
}

func (h *AppRouter) sendSettingsMenu(ctx context.Context, b *bot.Bot, chatID int64, user *model.User) {
	text := fmt.Sprintf("⚙️ Settings\n\n"+
		"🌐 Native Language: %s\n"+
		"🎯 Learning: %s\n"+
		"🗣 Interface: %s\n"+
		"🕐 Timezone: %s\n"+
		"⏱ Test Frequency: %s\n"+
		"🔇 Quiet Hours: %s\n"+
		"⏰ Start Time: %s\n"+
		"⏰ End Time: %s\n"+
		"📅 Days: %s\n\n"+
		"During quiet hours the bot will not send you word tests.\n\n"+
		"Choose an option:",
		formatLang(user.NativeLang),
		formatLang(user.TargetLang),
		formatLang(user.InterfaceLang),
		user.Timezone,
		formatTestInterval(user.GetTestIntervalHours()),
		formatEnabled(user.QuietHours.Enabled),
		user.QuietHours.StartTime,
		user.QuietHours.EndTime,
		formatDaysOfWeek(user.QuietHours.DaysOfWeek))

	keyboard := [][]models.InlineKeyboardButton{
		{{Text: localizer_lib.T("settings_btn_native"), CallbackData: "settings_native"}},
		{{Text: localizer_lib.T("settings_btn_target"), CallbackData: "settings_target"}},
		{{Text: localizer_lib.T("settings_btn_interface"), CallbackData: "settings_interface"}},
		{{Text: localizer_lib.T("settings_btn_timezone"), CallbackData: "settings_timezone"}},
		{{Text: localizer_lib.T("settings_btn_frequency"), CallbackData: "settings_test_interval"}},
		{{Text: localizer_lib.T("settings_btn_quiet"), CallbackData: "settings_quiet_hours"}},
		{{Text: localizer_lib.T("settings_btn_start"), CallbackData: "settings_start_time"}},
		{{Text: localizer_lib.T("settings_btn_end"), CallbackData: "settings_end_time"}},
		{{Text: localizer_lib.T("settings_btn_days"), CallbackData: "settings_days"}},
	}

	TGbot.SendMessage(ctx, b, chatID, text, &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	})
}

func (h *AppRouter) HandleSettingsCallback(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	user := h.getUserFromContextMsg(ctx, b, mes)
	callbackData := string(data)

	switch callbackData {
	case "settings_native":
		h.handleNativeLangSelection(ctx, b, mes, user)
	case "settings_target":
		h.handleTargetLangSelection(ctx, b, mes, user)
	case "settings_interface":
		h.handleInterfaceLangSelection(ctx, b, mes, user)
	case "settings_timezone":
		h.handleTimezoneSelection(ctx, b, mes, user)
	case "settings_test_interval":
		h.handleTestIntervalSelection(ctx, b, mes, user)
	case "settings_quiet_hours":
		h.handleQuietHoursToggle(ctx, b, mes, user)
	case "settings_start_time":
		h.handleStartTimeInput(ctx, b, mes, user)
	case "settings_end_time":
		h.handleEndTimeInput(ctx, b, mes, user)
	case "settings_days":
		h.handleDaysSelection(ctx, b, mes, user)
	case "settings_back":
		h.sendSettingsMenu(ctx, b, mes.Message.Chat.ID, user)
	}
}

const (
	setNativeLangPrefix = "setnative_"
	setTargetLangPrefix = "settarget_"
	setIfaceLangPrefix  = "setiface_"
)

func (h *AppRouter) handleNativeLangSelection(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, user *model.User) {
	TGbot.SendMessage(ctx, b, mes.Message.Chat.ID,
		localizer_lib.T("settings_select_native"),
		langSelectionKeyboard(setNativeLangPrefix, user.NativeLang, user.TargetLang))
}

func (h *AppRouter) handleTargetLangSelection(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, user *model.User) {
	TGbot.SendMessage(ctx, b, mes.Message.Chat.ID,
		localizer_lib.T("settings_select_target"),
		langSelectionKeyboard(setTargetLangPrefix, user.TargetLang, user.NativeLang))
}

func (h *AppRouter) handleInterfaceLangSelection(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, user *model.User) {
	TGbot.SendMessage(ctx, b, mes.Message.Chat.ID,
		localizer_lib.T("settings_select_interface"),
		langSelectionKeyboard(setIfaceLangPrefix, user.InterfaceLang, ""))
}

// HandleLangChange applies a native/target/interface language change chosen from
// the settings menu. Native and target must stay different; the interface
// language is independent (it only controls the language of buttons/templates).
func (h *AppRouter) HandleLangChange(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	user := h.getUserFromContextMsg(ctx, b, mes)
	callbackData := string(data)

	switch {
	case strings.HasPrefix(callbackData, setNativeLangPrefix):
		lang := strings.TrimPrefix(callbackData, setNativeLangPrefix)
		if consts.LangFullName[lang] == "" {
			return
		}
		if lang == user.TargetLang {
			TGbot.SendMessage(ctx, b, mes.Message.Chat.ID,
				localizer_lib.T("settings_langs_differ"),
				langSelectionKeyboard(setNativeLangPrefix, user.NativeLang, user.TargetLang))
			return
		}
		user.NativeLang = lang

	case strings.HasPrefix(callbackData, setTargetLangPrefix):
		lang := strings.TrimPrefix(callbackData, setTargetLangPrefix)
		if consts.LangFullName[lang] == "" {
			return
		}
		if lang == user.NativeLang {
			TGbot.SendMessage(ctx, b, mes.Message.Chat.ID,
				localizer_lib.T("settings_langs_differ"),
				langSelectionKeyboard(setTargetLangPrefix, user.TargetLang, user.NativeLang))
			return
		}
		user.TargetLang = lang

	case strings.HasPrefix(callbackData, setIfaceLangPrefix):
		lang := strings.TrimPrefix(callbackData, setIfaceLangPrefix)
		if consts.LangFullName[lang] == "" {
			return
		}
		user.InterfaceLang = lang
		// Apply immediately so the settings menu re-renders in the new language.
		localizer_lib.LoadLang(lang)

	default:
		return
	}

	if err := h.userRepository.Update(user); err != nil {
		h.handleError(ctx, b, mes.Message.Chat.ID, "Failed to update language.")
		return
	}

	h.sendSettingsMenu(ctx, b, mes.Message.Chat.ID, user)
}

// langSelectionKeyboard renders the four supported languages, marking the
// current selection and disabling the one already used for the opposite role.
func langSelectionKeyboard(prefix, current, other string) *models.InlineKeyboardMarkup {
	langs := []string{"uk", "en", "de", "es", "nl", "ru"}
	var rows [][]models.InlineKeyboardButton
	for _, lang := range langs {
		label := formatLang(lang)
		switch lang {
		case current:
			label = "✅ " + label
		case other:
			label = "🚫 " + label
		}
		rows = append(rows, []models.InlineKeyboardButton{{
			Text:         label,
			CallbackData: prefix + lang,
		}})
	}
	return &models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func formatLang(lang string) string {
	name := consts.LangFullName[lang]
	if name == "" {
		return "—"
	}
	return consts.LangFlags[lang] + " " + name
}

func (h *AppRouter) handleTimezoneSelection(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, user *model.User) {
	user.StateData.Scenario = TimezoneScenario
	h.userRepository.Update(user)

	keyboard := [][]models.InlineKeyboardButton{
		{{Text: "UTC", CallbackData: "timezone_UTC"}},
		{{Text: "Europe/London", CallbackData: "timezone_Europe/London"}},
		{{Text: "Europe/Berlin", CallbackData: "timezone_Europe/Berlin"}},
		{{Text: "Europe/Kyiv", CallbackData: "timezone_Europe/Kyiv"}},
		{{Text: "America/New_York", CallbackData: "timezone_America/New_York"}},
		{{Text: "Asia/Tokyo", CallbackData: "timezone_Asia/Tokyo"}},
	}

	TGbot.SendMessage(ctx, b, mes.Message.Chat.ID,
		localizer_lib.T("settings_select_timezone"),
		&models.InlineKeyboardMarkup{InlineKeyboard: keyboard})
}

func (h *AppRouter) handleTestIntervalSelection(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, user *model.User) {
	current := user.GetTestIntervalHours()

	var row []models.InlineKeyboardButton
	var keyboard [][]models.InlineKeyboardButton
	for hours := model.MinTestIntervalHours; hours <= model.MaxTestIntervalHours; hours++ {
		label := fmt.Sprintf("%dh", hours)
		if hours == current {
			label = "✅ " + label
		}
		row = append(row, models.InlineKeyboardButton{
			Text:         label,
			CallbackData: fmt.Sprintf("interval_%d", hours),
		})
		if len(row) == 4 {
			keyboard = append(keyboard, row)
			row = nil
		}
	}
	if len(row) > 0 {
		keyboard = append(keyboard, row)
	}

	TGbot.SendMessage(ctx, b, mes.Message.Chat.ID,
		localizer_lib.T("settings_frequency_q"),
		&models.InlineKeyboardMarkup{InlineKeyboard: keyboard})
}

func (h *AppRouter) handleQuietHoursToggle(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, user *model.User) {
	user.QuietHours.Enabled = !user.QuietHours.Enabled
	h.userRepository.Update(user)

	h.sendSettingsMenu(ctx, b, mes.Message.Chat.ID, user)
}

func (h *AppRouter) handleStartTimeInput(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, user *model.User) {
	user.StateData.Scenario = QuietHoursStartScenario
	h.userRepository.Update(user)

	TGbot.SendMessage(ctx, b, mes.Message.Chat.ID, localizer_lib.T("settings_enter_start"), nil)
}

func (h *AppRouter) handleEndTimeInput(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, user *model.User) {
	user.StateData.Scenario = QuietHoursEndScenario
	h.userRepository.Update(user)

	TGbot.SendMessage(ctx, b, mes.Message.Chat.ID, localizer_lib.T("settings_enter_end"), nil)
}

func (h *AppRouter) handleDaysSelection(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, user *model.User) {
	h.sendDaysKeyboard(ctx, mes.Message.Chat.ID, 0, user)
}

func (h *AppRouter) sendDaysKeyboard(ctx context.Context, chatID int64, msgID int, user *model.User) {
	dayNames := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	dayButton := func(day int) models.InlineKeyboardButton {
		label := dayNames[day-1]
		if user.QuietHours.HasDay(day) {
			label = "✅ " + label
		}
		return models.InlineKeyboardButton{
			Text:         label,
			CallbackData: fmt.Sprintf("day_%d", day),
		}
	}

	keyboard := [][]models.InlineKeyboardButton{
		{dayButton(1), dayButton(2), dayButton(3)},
		{dayButton(4), dayButton(5), dayButton(6)},
		{dayButton(7)},
		{{Text: localizer_lib.T("settings_days_all"), CallbackData: "days_all"}},
		{{Text: localizer_lib.T("settings_days_weekdays"), CallbackData: "days_weekdays"}},
		{{Text: localizer_lib.T("settings_days_weekend"), CallbackData: "days_weekend"}},
	}

	h.tgMessage.SendOrEditMessage(ctx, chatID, msgID, localizer_lib.T("settings_select_days"), &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	})
}

func (h *AppRouter) processTimezoneInput(ctx context.Context, b *bot.Bot, update *models.Update, user *model.User) {
	timezone := strings.TrimSpace(update.Message.Text)

	if _, err := time.LoadLocation(timezone); err != nil {
		TGbot.SendMessage(ctx, b, update.Message.Chat.ID, "❌ Invalid timezone. Please enter a valid timezone (e.g., UTC, Europe/London):", nil)
		return
	}

	user.Timezone = timezone
	user.StateData.Scenario = ""
	h.userRepository.Update(user)

	h.sendSettingsMenu(ctx, b, update.Message.Chat.ID, user)
}

func (h *AppRouter) processStartTimeInput(ctx context.Context, b *bot.Bot, update *models.Update, user *model.User) {
	timeStr, ok := normalizeTime(update.Message.Text)
	if !ok {
		TGbot.SendMessage(ctx, b, update.Message.Chat.ID, "❌ Invalid time format. Please use HH:MM format (e.g., 22:00):", nil)
		return
	}

	user.QuietHours.StartTime = timeStr
	user.StateData.Scenario = ""
	h.userRepository.Update(user)

	h.sendSettingsMenu(ctx, b, update.Message.Chat.ID, user)
}

func (h *AppRouter) processEndTimeInput(ctx context.Context, b *bot.Bot, update *models.Update, user *model.User) {
	timeStr, ok := normalizeTime(update.Message.Text)
	if !ok {
		TGbot.SendMessage(ctx, b, update.Message.Chat.ID, "❌ Invalid time format. Please use HH:MM format (e.g., 08:00):", nil)
		return
	}

	user.QuietHours.EndTime = timeStr
	user.StateData.Scenario = ""
	h.userRepository.Update(user)

	h.sendSettingsMenu(ctx, b, update.Message.Chat.ID, user)
}

// normalizeTime validates the input and returns it in canonical HH:MM form
// so stored values are always comparable (e.g. "9:30" -> "09:30").
func normalizeTime(input string) (string, bool) {
	t, err := time.Parse("15:04", strings.TrimSpace(input))
	if err != nil {
		return "", false
	}
	return t.Format("15:04"), true
}

func formatTestInterval(hours int) string {
	if hours == 1 {
		return "every hour"
	}
	return fmt.Sprintf("every %d hours", hours)
}

func formatEnabled(enabled bool) string {
	if enabled {
		return "✅ Enabled"
	}
	return "❌ Disabled"
}

func formatDaysOfWeek(daysStr string) string {
	days := strings.Split(daysStr, ",")
	dayNames := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}

	var result []string
	for _, day := range days {
		if dayNum, err := strconv.Atoi(strings.TrimSpace(day)); err == nil && dayNum >= 1 && dayNum <= 7 {
			result = append(result, dayNames[dayNum-1])
		}
	}

	if len(result) == 0 {
		return "None"
	}
	if len(result) == 7 {
		return "All days"
	}
	return strings.Join(result, ", ")
}
