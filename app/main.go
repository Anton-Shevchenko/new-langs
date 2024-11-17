package main

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"langs/internal/command"
	"langs/internal/handlers"
	"langs/internal/repository/book_part_repository"
	"langs/internal/repository/book_progress_repository"
	"langs/internal/repository/book_repository"
	"langs/internal/repository/word_repository"
	"langs/internal/tg_bot/tg_keyboard"
	"langs/internal/tg_bot/tg_msg"
	"langs/pkg/TGbot"
	"langs/pkg/book_reader"
	"log"
	"net/http"
	"os"
	"os/signal"

	"langs/internal/interfaces"
	"langs/internal/model"
	"langs/internal/repository/user_repository"
	"langs/internal/service"
	"langs/pkg/db"
)

var (
	userService interfaces.UserService
	user        *model.User
	userRepo    interfaces.UserRepository
)

func main() {
	ctx, cancel := setupSignalContext()
	defer cancel()

	b := initDependencies(ctx)

	b.StartWebhook(ctx)
}

func setupSignalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
}

func initDependencies(ctx context.Context) *bot.Bot {
	var b *bot.Bot

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")

	if botToken == "" {
		botToken = "6631879525:AAFM7m_0W7IlW1d2II5rtE4-mCH16Pl-sY8"
	}

	postgresConnection := db.NewPostgresConnection("")
	postgresDB := db.NewDB(postgresConnection).DB

	err := postgresDB.AutoMigrate(
		&model.User{},
		&model.Word{},
		&model.Book{},
		&model.BookPart{},
		&model.BookProgress{},
	)
	if err != nil {
		fmt.Println("Database migration error:", err)
	}

	reader := book_reader.NewBookReader(postgresConnection)

	userRepo = user_repository.NewUserRepository(postgresConnection)
	wordRepo := word_repository.NewWordRepository(postgresConnection)
	bookRepo := book_repository.NewBookRepository(postgresConnection)
	bookPartRepo := book_part_repository.NewBookPartRepository(postgresConnection)
	bookProgressRepo := book_progress_repository.NewBookProgressRepository(postgresConnection)

	userService = service.NewUserService(userRepo)
	readerService := service.NewReaderService(bookRepo, bookPartRepo, reader, bookProgressRepo)

	tgKeyboard := tg_keyboard.NewTGKeyboard()
	tgMessage := tg_msg.NewTGMessageService(b, tgKeyboard)

	wordService := service.NewWordService(wordRepo)

	tgHandler := handlers.NewTGHandler(tgKeyboard, tgMessage, wordService, wordRepo, userRepo, userService, readerService)
	commandSet := command.NewCommandSet(userService, userRepo, tgKeyboard, wordRepo, tgHandler)

	botOptions := []bot.Option{
		bot.WithMiddlewares(initUser),
		bot.WithDebug(),
		bot.WithDefaultHandler(tgHandler.DefaultHandler),
		bot.WithMessageTextHandler("/start", bot.MatchTypeExact, commandSet.Start),
	}

	b, err = bot.New(botToken, botOptions...)

	b.SetWebhook(ctx, &bot.SetWebhookParams{
		URL: "https://anton-shevchenko.com/webhook",
	})

	go func() {
		log.Println("Starting server on :443")
		err = http.ListenAndServeTLS(":443", "fullchain.crt", "server.key", b.WebhookHandler())
		if err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	tgMessage.B = b

	//job := jobs.NewSendWordJob(wordRepo, tgMessage, userRepo)
	//job.Execute(tgHandler.Te)

	//chatIds, err := userRepo.GetAllChatIDs()
	//if err != nil {
	//	return nil
	//}
	//
	//for _, id := range chatIds {
	//	go b.SendMessage(ctx, &bot.SendMessageParams{
	//		ChatID: id,
	//		Text:   "Choose an option:",
	//		ReplyMarkup: tgKeyboard.InitMenuKeyboard(
	//			b,
	//			tgHandler.OnWordList,
	//			tgHandler.HandleBack,
	//			tgHandler.HandleBooks,
	//		),
	//	})
	//
	//}

	return b
}

func initUser(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message != nil || update.CallbackQuery.Message.Message.Chat.ID != 0 {
			chatID := TGbot.GetChatIDFromUpdate(update)
			user = userService.InitUser(chatID)

			if user == nil {
				user = &model.User{
					ChatID: chatID,
				}

				fmt.Println("USER", user)
				userService.Upsert(user)

			}
		}

		ctx = context.Background()
		ctx = context.WithValue(ctx, "user", user)

		next(ctx, b, update)
	}
}
