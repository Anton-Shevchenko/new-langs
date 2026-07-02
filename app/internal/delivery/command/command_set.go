package command

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"langs/internal/consts"
	"langs/internal/delivery/bot"
	"langs/internal/domain"
	"langs/internal/infrastructure/database/repository/word_repository"
	TGbot "langs/internal/infrastructure/platform/telegram/helper"
	"langs/internal/infrastructure/platform/telegram/tg_keyboard"
	"langs/internal/interfaces"
	"langs/pkg/nlp/localizer_lib"
)

type CommandSet struct {
	userService    interfaces.UserService
	userRepo       interfaces.UserRepository
	tgKeyboard     *tg_keyboard.TGKeyboard
	wordRepository *word_repository.WordRepository
	handler        *handlers.AppRouter
}

func NewCommandSet(
	userService interfaces.UserService,
	userRepo interfaces.UserRepository,
	tgKeyboard *tg_keyboard.TGKeyboard,
	wordRepository *word_repository.WordRepository,
	handler *handlers.AppRouter,
) *CommandSet {
	return &CommandSet{
		userService:    userService,
		userRepo:       userRepo,
		tgKeyboard:     tgKeyboard,
		wordRepository: wordRepository,
		handler:        handler,
	}
}

func (s *CommandSet) Start(ctx context.Context, b *bot.Bot, update *models.Update) {
	s.AskNativeLang(ctx, b, update)

	go b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Main Menu - Choose an option:",
		ReplyMarkup: s.tgKeyboard.InitMainMenuKeyboard(
			b,
			s.handler.OnWordList,
			s.handler.OnTestMe,
			s.handler.OnSecondaryMenu,
		),
	})
}

func (s *CommandSet) AskNativeLang(ctx context.Context, b *bot.Bot, update *models.Update) {
	text := "🇬🇧 Please select your native language.\n\n" +
		"🇺🇦 Оберіть вашу рідну мову:\n\n" +
		"🇩🇪 Wählen Sie Ihre Muttersprache aus:\n\n" +
		"🇪🇸 Selecciona tu lengua materna:\n\n" +
		"👇 Tap your language below to continue."

	TGbot.SendMessage(
		ctx,
		b,
		update.Message.Chat.ID,
		text,
		s.tgKeyboard.GetLangsKeyboard(b, s.onNativeLangAndAskLangToLearn, "native"),
	)
}

func (s *CommandSet) onNativeLangAndAskLangToLearn(
	ctx context.Context,
	b *bot.Bot,
	mes models.MaybeInaccessibleMessage,
	data []byte,
) {
	user, ok := ctx.Value("user").(*model.User)

	if !ok || user == nil {
		panic("user not found in context")
	}

	user.NativeLang = string(data)
	user.InterfaceLang = string(data)
	err := s.userRepo.Update(user)

	if err != nil {
		fmt.Println("Error updating user:", err.Error())

		return
	}

	s.userService.InitUser(user.ChatId)
	s.AskLangToStudy(ctx, b, mes)
}

func (s *CommandSet) AskLangToStudy(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage) {
	TGbot.SendPrompt(
		ctx,
		b,
		mes.Message,
		localizer_lib.T("Select_language_you_want_to_learn"),
		s.tgKeyboard.GetLangsKeyboard(b, s.onLangToLearn, "target"),
	)
}

func (s *CommandSet) onLangToLearn(
	ctx context.Context,
	b *bot.Bot,
	mes models.MaybeInaccessibleMessage,
	data []byte,
) {
	user, err := s.userRepo.First(mes.Message.Chat.ID)

	if err != nil {
		return
	}

	user.TargetLang = string(data)
	if err = s.userRepo.Update(user); err != nil {
		fmt.Println("Error updating user:", err)
	}

	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    mes.Message.Chat.ID,
		Text:      consts.LangFlags[user.NativeLang] + consts.LangFlags[user.TargetLang],
		MessageID: mes.Message.ID,
	})

	if err = s.userRepo.Update(user); err != nil {
		fmt.Println("Error updating user:", err)
	}
}
