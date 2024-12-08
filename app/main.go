package main

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"langs/internal/command"
	"langs/internal/handlers"
	"langs/internal/interfaces"
	"langs/internal/jobs"
	"langs/internal/model"
	"langs/internal/repository/book_part_repository"
	"langs/internal/repository/book_progress_repository"
	"langs/internal/repository/book_repository"
	"langs/internal/repository/user_repository"
	"langs/internal/repository/word_repository"
	"langs/internal/service"
	"langs/internal/tg_bot/tg_keyboard"
	"langs/internal/tg_bot/tg_msg"
	"langs/pkg/TGbot"
	"langs/pkg/book_reader"
	"langs/pkg/db"
	"log"
	"net/http"
	"os"
	"os/signal"
)

var (
	userService interfaces.UserService
	user        *model.User
	userRepo    interfaces.UserRepository
)

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")

		return
	}
	ctx, cancel := setupSignalContext()
	defer cancel()

	b := initDependencies(ctx)

	startBot(ctx, b)
}

func setupSignalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
}

func initDependencies(ctx context.Context) *bot.Bot {
	var b *bot.Bot

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")

	if botToken == "" {
		botToken = "6631879525:AAFM7m_0W7IlW1d2II5rtE4-mCH16Pl-sY8"
		//botToken = "6207869114:AAEFF8TKV24lJaUj6__m5kt4lT9dGj3BETY"
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

	wordService := service.NewWordService(wordRepo, tgMessage)

	tgHandler := handlers.NewTGHandler(tgKeyboard, tgMessage, wordService, wordRepo, userRepo, userService, readerService)
	commandSet := command.NewCommandSet(userService, userRepo, tgKeyboard, wordRepo, tgHandler)
	tgHandler.SetCommandSet(commandSet)

	botOptions := []bot.Option{
		bot.WithMiddlewares(initUser),
		bot.WithDebug(),
		bot.WithDefaultHandler(tgHandler.DefaultHandler),
		bot.WithMessageTextHandler("/start", bot.MatchTypeExact, commandSet.Start),
	}

	b, err = bot.New(botToken, botOptions...)

	startServer(ctx, b)

	tgMessage.B = b

	chatIds, err := userRepo.GetAllChatIDs()
	if err != nil {
		return nil
	}

	for _, id := range chatIds {
		go b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: id,
			Text:   "Choose an option:",
			ReplyMarkup: tgKeyboard.InitMenuKeyboard(
				b,
				tgHandler.OnWordList,
				tgHandler.HandleBack,
				tgHandler.HandleBooks,
				tgHandler.OnTestMe,
			),
		})

	}

	c := cron.New()
	c.AddFunc("*/59 * * * *", func() {
		job := jobs.NewSendWordJob(wordService, userRepo)
		job.Execute(tgHandler.OnTestAnswer)
		log.Println("Scheduler started. The task will run every 59 minutes.")
	})
	c.Start()
	log.Println("Scheduler started. The task will run every 59 minutes.")

	return b
}

func initUser(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message != nil || update.CallbackQuery.Message.Message.Chat.ID != 0 {
			chatID := TGbot.GetChatIDFromUpdate(update)
			user = userService.InitUser(chatID)

			if user == nil {
				user = &model.User{
					ChatId: chatID,
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

func startServerWithSSL(ctx context.Context, b *bot.Bot) {
	b.SetWebhook(ctx, &bot.SetWebhookParams{
		URL: "https://anton-shevchenko.com/webhook",
	})

	go func() {
		err := http.ListenAndServeTLS(":443", "fullchain.crt", "server.key", b.WebhookHandler())
		http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Test GET endpoint is working"))
		})
		if err != nil {
			log.Fatalf("Failed to start server with SSL: %v", err)
		}
	}()
}

func startBot(ctx context.Context, b *bot.Bot) {
	env := os.Getenv("ENV")
	if env == "prod" {
		b.StartWebhook(ctx)
	} else {
		b.DeleteWebhook(ctx, &bot.DeleteWebhookParams{
			DropPendingUpdates: true,
		})
		b.Start(ctx)
	}
}

func startServer(ctx context.Context, b *bot.Bot) {
	env := os.Getenv("ENV")

	fmt.Println("ENV", env)
	if env == "prod" {
		startServerWithSSL(ctx, b)
	} else {

	}
}
