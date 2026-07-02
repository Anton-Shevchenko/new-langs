package tg_msg

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/keyboard/inline"
	"langs/internal/consts"
	"langs/internal/domain"
	"langs/internal/infrastructure/platform/telegram/tg_keyboard"
	"langs/internal/presentation/formatter"
	"langs/pkg/nlp/wordTranslator"
)

type TGMessageService struct {
	keyboard *tg_keyboard.TGKeyboard
	B        *bot.Bot
}

func NewTGMessageService(b *bot.Bot, keyboard *tg_keyboard.TGKeyboard) *TGMessageService {
	return &TGMessageService{
		B:        b,
		keyboard: keyboard,
	}
}

func (tgm *TGMessageService) sendMessage(
	ctx context.Context,
	chatID int64,
	text string,
	replyMarkup interface{},
) error {
	_, err := tgm.B.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: replyMarkup,
	})
	return err
}

func (tgm *TGMessageService) editMessage(
	ctx context.Context,
	chatID int64,
	msgID int,
	text string,
	replyMarkup interface{},
) error {
	_, err := tgm.B.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   msgID,
		Text:        text,
		ReplyMarkup: replyMarkup,
	})
	return err
}

func (tgm *TGMessageService) handleError(err error) {
	if err != nil {
		fmt.Println("Error sending message:", err)
	}
}

func (tgm *TGMessageService) SendWordTest(
	chatId int64,
	handle inline.OnSelect,
	word *model.WordOption,
	translations []string,
) {
	err := tgm.sendMessage(
		context.Background(),
		chatId,
		testQuestionText(word),
		tgm.keyboard.GetAnswerOptionsKeyboard(tgm.B, word.WordID, handle, translations),
	)
	tgm.handleError(err)
}

// testQuestionText renders the quiz prompt: the headword in its language and a
// placeholder for the translation the user has to recall.
func testQuestionText(word *model.WordOption) string {
	return fmt.Sprintf(
		"%s <b>%s</b>\n%s ❓",
		consts.LangFlags[word.WordLang], word.Word,
		consts.LangFlags[word.TranslationLang],
	)
}

func (tgm *TGMessageService) SendExam(
	chatId int64,
	handle inline.OnSelect,
	word *model.WordOption,
) {
	err := tgm.sendMessage(
		context.Background(),
		chatId,
		testQuestionText(word),
		tgm.keyboard.GetWriteTestKeyboard(tgm.B, word.WordID, handle),
	)
	tgm.handleError(err)
}

func (tgm *TGMessageService) SendWordView(
	ctx context.Context,
	sourceWord string,
	sourceWordLang string,
	msg *models.Message,
	wordMsg *wordTranslator.TranslateResult,
	handlerFunc inline.OnSelect,
	onCustomTranslation inline.OnSelect,
) {
	user, ok := ctx.Value("user").(*model.User)
	if !ok {
		panic("user not found in context")
	}

	var msgText string

	if user.NativeLang == sourceWordLang {
		msgText = wordMsg.ToNativeWordString()
		if wordMsg.TranslationLang == "de" && wordMsg.Article != "" && len(wordMsg.Translations) > 0 {
			wordWithArticle := wordMsg.Article + " " + wordMsg.Translations[0]
			wordMsg.Translations = append([]string{wordWithArticle}, wordMsg.Translations...)
		}
	} else {
		msgText = wordMsg.ToString(user.NativeLang)
	}

	err := tgm.sendMessage(
		ctx,
		msg.Chat.ID,
		msgText,
		tgm.keyboard.GetTranslateOptionsKeyboard(
			tgm.B,
			sourceWord,
			handlerFunc,
			onCustomTranslation,
			wordMsg.Translations,
		),
	)
	tgm.handleError(err)
}

func (tgm *TGMessageService) SendFormattedWordMessage(
	ctx context.Context,
	chatID int64,
	sourceWord,
	customTranslation string,
) {
	msgText := formatter.FormatWordMessage(sourceWord, customTranslation)
	err := tgm.sendMessage(ctx, chatID, msgText, nil)
	tgm.handleError(err)
}

func (tgm *TGMessageService) SendOrEditMessage(
	ctx context.Context,
	chatId int64,
	msgId int,
	text string,
	replyMarkup interface{},
) {
	var err error
	if msgId == 0 {
		err = tgm.sendMessage(ctx, chatId, text, replyMarkup)
	} else {
		err = tgm.editMessage(ctx, chatId, msgId, text, replyMarkup)
	}
	tgm.handleError(err)
}
