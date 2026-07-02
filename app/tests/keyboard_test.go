package tests

import (
	"context"
	"testing"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/stretchr/testify/assert"

	"langs/internal/infrastructure/platform/telegram/tg_keyboard"
)

func TestKeyboardCreation(t *testing.T) {
	t.Run("InitMenuKeyboard", func(t *testing.T) {
		mockBot := &bot.Bot{}

		keyboard := tg_keyboard.NewTGKeyboard()
		menuKeyboard := keyboard.InitMenuKeyboard(
			mockBot,
			func(ctx context.Context, b *bot.Bot, update *models.Update) {},
			func(ctx context.Context, b *bot.Bot, update *models.Update) {},
			func(ctx context.Context, b *bot.Bot, update *models.Update) {},
		)

		assert.NotNil(t, menuKeyboard)
	})

	t.Run("InitMainMenuKeyboard", func(t *testing.T) {
		mockBot := &bot.Bot{}

		keyboard := tg_keyboard.NewTGKeyboard()
		mainMenuKeyboard := keyboard.InitMainMenuKeyboard(
			mockBot,
			func(ctx context.Context, b *bot.Bot, update *models.Update) {},
			func(ctx context.Context, b *bot.Bot, update *models.Update) {},
			func(ctx context.Context, b *bot.Bot, update *models.Update) {},
		)

		assert.NotNil(t, mainMenuKeyboard)
	})

	t.Run("InitSecondaryMenuKeyboard", func(t *testing.T) {
		mockBot := &bot.Bot{}

		keyboard := tg_keyboard.NewTGKeyboard()
		secondaryMenuKeyboard := keyboard.InitSecondaryMenuKeyboard(
			mockBot,
			func(ctx context.Context, b *bot.Bot, update *models.Update) {},
			func(ctx context.Context, b *bot.Bot, update *models.Update) {},
			func(ctx context.Context, b *bot.Bot, update *models.Update) {},
			func(ctx context.Context, b *bot.Bot, update *models.Update) {},
		)

		assert.NotNil(t, secondaryMenuKeyboard)
	})
}

func TestKeyboardButtons(t *testing.T) {
	t.Run("MenuKeyboardButtons", func(t *testing.T) {
		mockBot := &bot.Bot{}

		keyboard := tg_keyboard.NewTGKeyboard()
		menuKeyboard := keyboard.InitMenuKeyboard(
			mockBot,
			func(ctx context.Context, b *bot.Bot, update *models.Update) {},
			func(ctx context.Context, b *bot.Bot, update *models.Update) {},
			func(ctx context.Context, b *bot.Bot, update *models.Update) {},
		)

		assert.NotNil(t, menuKeyboard)

	})

	t.Run("MainMenuKeyboardButtons", func(t *testing.T) {
		mockBot := &bot.Bot{}

		keyboard := tg_keyboard.NewTGKeyboard()
		mainMenuKeyboard := keyboard.InitMainMenuKeyboard(
			mockBot,
			func(ctx context.Context, b *bot.Bot, update *models.Update) {},
			func(ctx context.Context, b *bot.Bot, update *models.Update) {},
			func(ctx context.Context, b *bot.Bot, update *models.Update) {},
		)

		assert.NotNil(t, mainMenuKeyboard)

	})
}
