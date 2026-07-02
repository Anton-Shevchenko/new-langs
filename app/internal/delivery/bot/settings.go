package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"langs/internal/domain"
	TGbot "langs/internal/infrastructure/platform/telegram/helper"
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
		"🕐 Timezone: %s\n"+
		"🔇 Quiet Hours: %s\n"+
		"⏰ Start Time: %s\n"+
		"⏰ End Time: %s\n"+
		"📅 Days: %s\n\n"+
		"During quiet hours the bot will not send you word tests.\n\n"+
		"Choose an option:",
		user.Timezone,
		formatEnabled(user.QuietHours.Enabled),
		user.QuietHours.StartTime,
		user.QuietHours.EndTime,
		formatDaysOfWeek(user.QuietHours.DaysOfWeek))

	keyboard := [][]models.InlineKeyboardButton{
		{{Text: "🕐 Set Timezone", CallbackData: "settings_timezone"}},
		{{Text: "🔇 Toggle Quiet Hours", CallbackData: "settings_quiet_hours"}},
		{{Text: "⏰ Start Time", CallbackData: "settings_start_time"}},
		{{Text: "⏰ End Time", CallbackData: "settings_end_time"}},
		{{Text: "📅 Days of Week", CallbackData: "settings_days"}},
		{{Text: "🔙 Back", CallbackData: "settings_back"}},
	}

	TGbot.SendMessage(ctx, b, chatID, text, &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	})
}

func (h *AppRouter) HandleSettingsCallback(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	user := h.getUserFromContextMsg(ctx, b, mes)
	callbackData := string(data)

	switch callbackData {
	case "settings_timezone":
		h.handleTimezoneSelection(ctx, b, mes, user)
	case "settings_quiet_hours":
		h.handleQuietHoursToggle(ctx, b, mes, user)
	case "settings_start_time":
		h.handleStartTimeInput(ctx, b, mes, user)
	case "settings_end_time":
		h.handleEndTimeInput(ctx, b, mes, user)
	case "settings_days":
		h.handleDaysSelection(ctx, b, mes, user)
	case "settings_back":
		h.OnBack(ctx, b, &models.Update{
			CallbackQuery: &models.CallbackQuery{
				Message: mes,
			},
		})
	}
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
		{{Text: "🔙 Back", CallbackData: "settings_back"}},
	}

	TGbot.SendMessage(ctx, b, mes.Message.Chat.ID,
		"🕐 Select your timezone or type it (e.g., Europe/Berlin):",
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

	TGbot.SendMessage(ctx, b, mes.Message.Chat.ID, "⏰ Enter start time (HH:MM format, e.g., 22:00):", nil)
}

func (h *AppRouter) handleEndTimeInput(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, user *model.User) {
	user.StateData.Scenario = QuietHoursEndScenario
	h.userRepository.Update(user)

	TGbot.SendMessage(ctx, b, mes.Message.Chat.ID, "⏰ Enter end time (HH:MM format, e.g., 08:00):", nil)
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
		{{Text: "All Days", CallbackData: "days_all"}},
		{{Text: "Weekdays", CallbackData: "days_weekdays"}},
		{{Text: "Weekend", CallbackData: "days_weekend"}},
		{{Text: "🔙 Back", CallbackData: "settings_back"}},
	}

	h.tgMessage.SendOrEditMessage(ctx, chatID, msgID, "📅 Select days for quiet hours:", &models.InlineKeyboardMarkup{
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
