package tg_keyboard

import (
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/ui/keyboard/inline"
	"github.com/go-telegram/ui/keyboard/reply"
	"strconv"
)

type TGKeyboard struct{}

func NewTGKeyboard() *TGKeyboard {
	return &TGKeyboard{}
}

type ListItem struct {
	Id   int64
	Data string
}

func (r *TGKeyboard) GetLangsKeyboard(b *bot.Bot, onSelect inline.OnSelect, prefix string) *inline.Keyboard {
	return inline.New(
		b,
		inline.WithPrefix(prefix),
		inline.NoDeleteAfterClick(),
	).
		Row().
		Button("English", []byte("en"), onSelect).
		Row().
		Button("Українська", []byte("uk"), onSelect).
		Row().
		Button("Deutsch", []byte("de"), onSelect).
		Row().
		Button("Espana", []byte("es"), onSelect)
}

func (r *TGKeyboard) GetTranslateOptionsKeyboard(
	b *bot.Bot,
	sourceWord string,
	onSelect inline.OnSelect,
	onCustomTranslation inline.OnSelect,
	options []string,
) *inline.Keyboard {
	keyboard := inline.New(
		b,
		inline.NoDeleteAfterClick(),
	)

	for _, option := range options {
		keyboard.
			Row().
			Button(option, []byte(option), onSelect)
	}

	return keyboard.
		Row().
		Button("My option", []byte(sourceWord), onCustomTranslation)
}

func (r *TGKeyboard) GetAnswerOptionsKeyboard(
	b *bot.Bot,
	sourceWordId int64,
	onSelect inline.OnSelect,
	options []string,
) *inline.Keyboard {
	keyboard := inline.New(
		b,
		inline.NoDeleteAfterClick(),
	)

	for _, option := range options {
		keyboard.
			Row().
			Button(option, []byte(option), onSelect)
	}

	return keyboard
}

func (r *TGKeyboard) InitMenuKeyboard(
	b *bot.Bot,
	onWordList bot.HandlerFunc,
	onBack bot.HandlerFunc,
	onBook bot.HandlerFunc,
) *reply.ReplyKeyboard {
	return reply.New(reply.IsSelective()).
		Button("List", b, bot.MatchTypeExact, onWordList).
		Row().
		Button("Back", b, bot.MatchTypeExact, onBack).
		Row().
		Button("Book", b, bot.MatchTypeExact, onBook)
}

func (r *TGKeyboard) InitReaderKeyboard(
	b *bot.Bot,
	onBookNext bot.HandlerFunc,
	onBookPrev bot.HandlerFunc,
	onBookResend bot.HandlerFunc,
	onBack bot.HandlerFunc,
) *reply.ReplyKeyboard {
	return reply.New(reply.IsSelective(), reply.WithPrefix("reader")).
		Button("⬅️\u200B", b, bot.MatchTypeExact, onBookPrev).
		Button("🔄\u200B", b, bot.MatchTypeExact, onBookResend).
		Button("➡️\u200B", b, bot.MatchTypeExact, onBookNext).
		Row().
		Button("Back", b, bot.MatchTypeExact, onBack)
}

func (r *TGKeyboard) GetWordViewKeyboard(
	b *bot.Bot,
	id int,
	handlerFunc inline.OnSelect,
) *inline.Keyboard {
	keyboard := inline.New(
		b,
		inline.NoDeleteAfterClick(),
	)

	return keyboard.
		Row().
		Button("delete 🚮", []byte(strconv.Itoa(id)), handlerFunc)
}

func (r *TGKeyboard) BuildReplacementKeyboard(
	b *bot.Bot,
	prefix string,
	replacements []string,
	handlerFunc inline.OnSelect,
) *inline.Keyboard {
	replacementKeyboard := inline.New(b, inline.WithPrefix("r_"+prefix), inline.NoDeleteAfterClick())
	for _, replacement := range replacements {
		replacementKeyboard.Row().Button(replacement, []byte(replacement), handlerFunc)
	}
	return replacementKeyboard
}

func (r *TGKeyboard) GetListingKeyboard(
	b *bot.Bot,
	listItems []*ListItem,
	handlerFunc inline.OnSelect,
	handlerPaginationFunc inline.OnSelect,
	page int,
) *inline.Keyboard {
	keyboard := inline.New(
		b,
		inline.NoDeleteAfterClick(),
	)

	for _, item := range listItems {
		keyboard.
			Row().
			Button(item.Data, []byte(fmt.Sprintf("%d", item.Id)), handlerFunc)
	}

	if len(listItems) > 10 {
		keyboard.
			Row().
			Button("<", []byte("<,"+strconv.Itoa(page)), handlerPaginationFunc).
			Button(">", []byte(">,"+strconv.Itoa(page)), handlerPaginationFunc)
	}

	return keyboard
}
