package command

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"langs/internal/consts"
	"langs/internal/handlers"
	"langs/internal/interfaces"
	"langs/internal/model"
	"langs/internal/repository/word_repository"
	"langs/internal/tg_bot/tg_keyboard"
	"langs/pkg/TGbot"
	"langs/pkg/localizer_lib"
)

type CommandSet struct {
	userService    interfaces.UserService
	userRepo       interfaces.UserRepository
	tgKeyboard     *tg_keyboard.TGKeyboard
	wordRepository *word_repository.WordRepository
	handler        *handlers.TGHandler
}

func NewCommandSet(
	userService interfaces.UserService,
	userRepo interfaces.UserRepository,
	tgKeyboard *tg_keyboard.TGKeyboard,
	wordRepository *word_repository.WordRepository,
	handler *handlers.TGHandler,
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
	s.askNativeLang(ctx, b, update)
}

func (s *CommandSet) askNativeLang(ctx context.Context, b *bot.Bot, update *models.Update) {
	text := "🇬🇧 Please select your native language.\n\n" +
		"🇺🇦 Оберіть вашу рідну мову:\n\n" +
		"🇩🇪 Wählen Sie Ihre Muttersprache aus:\n\n" +
		"🇪🇸 Selecciona tu lengua materna:\n\n" +
		"👇 Tap your language below to continue."

	TGbot.SendPrompt(
		ctx,
		b,
		update.Message,
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

	s.userService.InitUser(user.ChatID)
	s.askLangToStudy(ctx, b, mes)
}

func (s *CommandSet) askLangToStudy(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage) {
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
