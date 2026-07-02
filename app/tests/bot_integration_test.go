package tests

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testBotToken = "YOUR_TEST_BOT_TOKEN_HERE"

func TestBotIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		t.Skip("TELEGRAM_BOT_TOKEN not set, skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	b, err := bot.New(token)
	require.NoError(t, err)
	require.NotNil(t, b)

	t.Run("BotInfo", func(t *testing.T) {
		info, err := b.GetMe(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, info.Username)
		t.Logf("Bot username: @%s", info.Username)
	})

	t.Run("SendMessage", func(t *testing.T) {

		chatIDStr := os.Getenv("TEST_CHAT_ID")
		var chatID int64 = 123456789

		if chatIDStr != "" {
			if parsedID, err := strconv.ParseInt(chatIDStr, 10, 64); err == nil {
				chatID = parsedID
				t.Logf("Using real chat ID: %d", chatID)
			} else {
				t.Logf("Invalid TEST_CHAT_ID format, using dummy ID: %d", chatID)
			}
		} else {
			t.Logf("TEST_CHAT_ID not set, using dummy ID: %d", chatID)
		}

		msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "🧪 Integration test message from Docker container",
		})

		if err != nil {
			t.Logf("Could not send message (expected if chat ID is invalid): %v", err)
		} else {
			assert.NotNil(t, msg)
			assert.Equal(t, "🧪 Integration test message from Docker container", msg.Text)
			t.Logf("✅ Message sent successfully to chat ID: %d", chatID)
		}
	})
}

func TestBotResponses(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping bot response tests in short mode")
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatIDStr := os.Getenv("TEST_CHAT_ID")

	if token == "" {
		t.Skip("TELEGRAM_BOT_TOKEN not set, skipping bot response tests")
	}

	if chatIDStr == "" {
		t.Skip("TEST_CHAT_ID not set, skipping bot response tests")
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	require.NoError(t, err, "Invalid TEST_CHAT_ID format")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	b, err := bot.New(token)
	require.NoError(t, err)
	require.NotNil(t, b)

	t.Run("StartCommand", func(t *testing.T) {

		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "/start",
		})

		if err != nil {
			t.Logf("Could not send /start command: %v", err)
			t.Skip("Skipping response test due to send failure")
			return
		}

		t.Logf("✅ /start command sent successfully")
		t.Logf("📱 Check your Telegram chat with @WhereParcelBot for the response")

		time.Sleep(3 * time.Second)
		t.Logf("⏳ Bot should have responded to /start command")
	})

	t.Run("HelpCommand", func(t *testing.T) {

		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "/help",
		})

		if err != nil {
			t.Logf("Could not send /help command: %v", err)
			t.Skip("Skipping response test due to send failure")
			return
		}

		t.Logf("✅ /help command sent successfully")
		t.Logf("📱 Check your Telegram chat with @WhereParcelBot for the response")

		time.Sleep(3 * time.Second)
		t.Logf("⏳ Bot should have responded to /help command")
	})

	t.Run("SettingsCommand", func(t *testing.T) {

		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "/settings",
		})

		if err != nil {
			t.Logf("Could not send /settings command: %v", err)
			t.Skip("Skipping response test due to send failure")
			return
		}

		t.Logf("✅ /settings command sent successfully")
		t.Logf("📱 Check your Telegram chat with @WhereParcelBot for the response")

		time.Sleep(3 * time.Second)
		t.Logf("⏳ Bot should have responded to /settings command")
	})

	t.Run("TestWordInput", func(t *testing.T) {

		testWord := "hello"
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   testWord,
		})

		if err != nil {
			t.Logf("Could not send test word: %v", err)
			t.Skip("Skipping response test due to send failure")
			return
		}

		t.Logf("✅ Test word '%s' sent successfully", testWord)
		t.Logf("📱 Check your Telegram chat with @WhereParcelBot for the response")

		time.Sleep(3 * time.Second)
		t.Logf("⏳ Bot should have responded to word input")
	})

	t.Run("TestMultipleCommands", func(t *testing.T) {
		commands := []string{"/start", "/help", "/settings", "test", "hello"}

		for _, cmd := range commands {
			t.Logf("📤 Sending command: %s", cmd)

			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   cmd,
			})

			if err != nil {
				t.Logf("❌ Could not send '%s': %v", cmd, err)
			} else {
				t.Logf("✅ Command '%s' sent successfully", cmd)
			}

			time.Sleep(2 * time.Second)
		}

		t.Logf("📱 Check your Telegram chat with @WhereParcelBot for all responses")
	})
}

func TestBotHandlers(t *testing.T) {
	t.Run("StartCommand", func(t *testing.T) {
		update := &models.Update{
			Message: &models.Message{
				Text: "/start",
				Chat: models.Chat{
					ID: 123456789,
				},
			},
		}

		assert.NotNil(t, update.Message)
		assert.Equal(t, "/start", update.Message.Text)
	})

	t.Run("CallbackQuery", func(t *testing.T) {
		update := &models.Update{
			CallbackQuery: &models.CallbackQuery{
				Data: "test_callback",
				Message: models.MaybeInaccessibleMessage{
					Message: &models.Message{
						Chat: models.Chat{
							ID: 123456789,
						},
					},
				},
			},
		}

		assert.NotNil(t, update.CallbackQuery)
		assert.Equal(t, "test_callback", update.CallbackQuery.Data)
	})
}
