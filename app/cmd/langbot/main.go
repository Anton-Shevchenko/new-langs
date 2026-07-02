package main

import (
	"context"
	"fmt"
	bot_api "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/robfig/cron/v3"
	handlers "langs/internal/delivery/bot"
	"langs/internal/delivery/command"
	"langs/internal/domain"
	"langs/internal/infrastructure/book_reader"
	"langs/internal/infrastructure/database/connection"
	"langs/internal/infrastructure/database/repository/book_part_repository"
	"langs/internal/infrastructure/database/repository/book_progress_repository"
	"langs/internal/infrastructure/database/repository/book_repository"
	"langs/internal/infrastructure/database/repository/user_repository"
	"langs/internal/infrastructure/database/repository/word_repository"
	"langs/internal/infrastructure/platform/telegram/tg_keyboard"
	"langs/internal/infrastructure/platform/telegram/tg_msg"
	"langs/internal/infrastructure/platform/whatsapp"
	"langs/internal/interfaces"
	service "langs/internal/usecase"
	"langs/internal/usecase/jobs"
	"langs/pkg/nlp/localizer_lib"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
)

var (
	userService interfaces.UserService
	userRepo    interfaces.UserRepository
	once        sync.Once
)

func main() {
	ctx, cancel := setupSignalContext()
	defer cancel()

	b := initDependencies(ctx)

	if b == nil {
		fmt.Println("Failed to initialize dependencies. Exiting.")
		return
	}

	startBot(ctx, b)
}

func setupSignalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
}

func initDependencies(ctx context.Context) *bot_api.Bot {
	var b *bot_api.Bot

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		fmt.Println("Warning: TELEGRAM_BOT_TOKEN not set")
	}

	pgUser := os.Getenv("PG_USER")
	if pgUser == "" {
		pgUser = "postgres"
	}

	pgPassword := os.Getenv("PG_PASSWORD")
	if pgPassword == "" {
		pgPassword = "pass"
	}

	fmt.Printf("Attempting to connect to database with user: %s\n", pgUser)

	postgresConnection := db.NewPostgresConnection(
		fmt.Sprintf(
			"host=db user=%s password=%s dbname=my_database port=5432 sslmode=disable",
			pgUser,
			pgPassword,
		))

	if postgresConnection == nil {
		fmt.Println("Error: Failed to connect to database")
		return nil
	}

	postgresDB := db.NewDB(postgresConnection).DB

	sqlDB, err := postgresDB.DB()
	if err != nil {
		fmt.Printf("Error getting underlying sql.DB: %v\n", err)
		return nil
	}

	if err := sqlDB.Ping(); err != nil {
		fmt.Printf("Error pinging database: %v\n", err)
		return nil
	}

	fmt.Println("Database connection successful!")

	err = postgresDB.AutoMigrate(
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

	appRouter := handlers.NewAppRouter(tgKeyboard, tgMessage, wordService, wordRepo, userRepo, userService, readerService)
	commandSet := command.NewCommandSet(userService, userRepo, tgKeyboard, wordRepo, appRouter)
	appRouter.SetCommandSet(commandSet)

	botOptions := []bot_api.Option{
		bot_api.WithMiddlewares(initUser),
		bot_api.WithDebug(),
		bot_api.WithDefaultHandler(appRouter.DefaultHandler),
		bot_api.WithMessageTextHandler("/start", bot_api.MatchTypeExact, commandSet.Start),
		bot_api.WithHTTPClient(30, &http.Client{Transport: whatsapp.NewProxyClient()}),
	}

	b, err = bot_api.New(botToken, botOptions...)
	if err != nil {
		fmt.Printf("Error creating bot: %v\n", err)
		return nil
	}

	startServer(ctx, b)

	tgMessage.B = b

	chatIds, err := userRepo.GetAllChatIDs()
	if err != nil {
		return nil
	}

	InitCron(appRouter, wordService, userRepo)

	for _, id := range chatIds {
		if user, err := userRepo.First(id); err == nil && user != nil {
			localizer_lib.LoadLang(user.InterfaceLang)
		}
		go b.SendMessage(ctx, &bot_api.SendMessageParams{
			ChatID: id,
			Text:   localizer_lib.T("menu_main_title"),
			ReplyMarkup: tgKeyboard.InitMainMenuKeyboard(
				b,
				appRouter.OnWordList,
				appRouter.OnTestMe,
				appRouter.OnSecondaryMenu,
			),
		})

	}

	return b
}

func initUser(next bot_api.HandlerFunc) bot_api.HandlerFunc {
	return func(ctx context.Context, b *bot_api.Bot, update *models.Update) {
		var chatID int64

		if update.Message != nil {
			chatID = update.Message.Chat.ID
		} else if update.CallbackQuery != nil && update.CallbackQuery.Message.Message != nil {
			chatID = update.CallbackQuery.Message.Message.Chat.ID
		} else {
			next(ctx, b, update)
			return
		}

		user := userService.InitUser(chatID)
		if user == nil {
			user = &model.User{ChatId: chatID}
			userService.Upsert(user)
		}

		ctx = context.WithValue(ctx, "user", user)
		next(ctx, b, update)
	}
}

func startBot(ctx context.Context, b *bot_api.Bot) {
	b.Start(ctx)
}

func startServer(ctx context.Context, b *bot_api.Bot) {
	go func() {
		waMux := http.NewServeMux()
		waMux.Handle("/whatsapp_webhook", whatsapp.NewWebhookHandler(b.WebhookHandler()))
		log.Println("Starting WhatsApp Webhook server on :8081")
		if err := http.ListenAndServe(":8081", waMux); err != nil {
			log.Fatalf("WhatsApp webhook server failed: %v", err)
		}
	}()
}

func InitCron(appRouter *handlers.AppRouter, wordService *service.WordService, userRepo interfaces.UserRepository) {
	job := jobs.NewSendWordJob(wordService, userRepo)

	once.Do(func() {
		c := cron.New()
		// Run hourly; each user receives tests according to their own
		// configured interval (1-8 hours) tracked via LastTestSentAt.
		c.AddFunc("0 * * * *", func() {
			log.Println("Running scheduled word tests")
			job.Execute(appRouter.OnTestAnswer, appRouter.OnWriteTestOption)
		})
		c.Start()
		log.Println("Scheduler started. Word tests are sent based on each user's interval.")
	})
}
