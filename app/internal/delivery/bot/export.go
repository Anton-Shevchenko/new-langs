package handlers

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"langs/internal/consts"
	"langs/internal/domain"
	TGbot "langs/internal/infrastructure/platform/telegram/helper"
	"langs/pkg/nlp/localizer_lib"
)

const exportCallbackPrefix = "export_"

// OnExport shows the list of language pairs the user has words for so they can
// pick one (or all) to download as a CSV file.
func (h *AppRouter) OnExport(ctx context.Context, b *bot.Bot, update *models.Update) {
	user := h.getUserFromContext(ctx, b, update)
	chatID := TGbot.GetChatIDFromUpdate(update)

	pairs, err := h.wordRepository.GetLangPairsByChatId(user.ChatId)
	if err != nil {
		h.handleError(ctx, b, chatID, "Could not load language pairs")
		return
	}
	if len(pairs) == 0 {
		TGbot.SendMessage(ctx, b, chatID, localizer_lib.T("export_no_words"), nil)
		return
	}

	var rows [][]models.InlineKeyboardButton
	for _, p := range pairs {
		rows = append(rows, []models.InlineKeyboardButton{{
			Text:         exportPairLabel(p),
			CallbackData: exportCallbackPrefix + p.Key(),
		}})
	}
	if len(pairs) > 1 {
		rows = append(rows, []models.InlineKeyboardButton{{
			Text:         localizer_lib.T("export_all_words"),
			CallbackData: exportCallbackPrefix + "all",
		}})
	}

	TGbot.SendMessage(ctx, b, chatID,
		localizer_lib.T("export_choose_pair"),
		&models.InlineKeyboardMarkup{InlineKeyboard: rows})
}

// HandleExportCallback builds and sends the CSV file for the chosen selection.
func (h *AppRouter) HandleExportCallback(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	user := h.getUserFromContextMsg(ctx, b, mes)
	chatID := mes.Message.Chat.ID
	arg := strings.TrimPrefix(string(data), exportCallbackPrefix)

	var (
		words    []*model.Word
		err      error
		fileName string
	)

	if arg == "all" {
		words, err = h.wordRepository.GetAllByChatId(user.ChatId)
		fileName = "words_all.csv"
	} else {
		langs := strings.Split(arg, "_")
		if len(langs) != 2 {
			return
		}
		words, err = h.wordRepository.AllByChatIdAndLangPair(user.ChatId, langs[0], langs[1])
		fileName = fmt.Sprintf("words_%s_%s.csv", langs[0], langs[1])
	}

	if err != nil {
		h.handleError(ctx, b, chatID, "Export failed")
		return
	}
	if len(words) == 0 {
		TGbot.SendMessage(ctx, b, chatID, localizer_lib.T("export_none_selection"), nil)
		return
	}

	csvBytes, err := buildWordsCSV(words)
	if err != nil {
		h.handleError(ctx, b, chatID, "Could not build CSV")
		return
	}

	_, err = b.SendDocument(ctx, &bot.SendDocumentParams{
		ChatID: chatID,
		Document: &models.InputFileUpload{
			Filename: fileName,
			Data:     bytes.NewReader(csvBytes),
		},
		Caption: fmt.Sprintf("%s: %d", localizer_lib.T("export_exported"), len(words)),
	})
	if err != nil {
		h.handleError(ctx, b, chatID, "Could not send file")
	}
}

func exportPairLabel(p model.LangPair) string {
	return fmt.Sprintf("%s %s ↔ %s %s (%d)",
		consts.LangFlags[p.Lang1], strings.ToUpper(p.Lang1),
		consts.LangFlags[p.Lang2], strings.ToUpper(p.Lang2),
		p.Count,
	)
}

func buildWordsCSV(words []*model.Word) ([]byte, error) {
	var buf bytes.Buffer
	// UTF-8 BOM so spreadsheet apps render Cyrillic/umlauts correctly.
	buf.WriteString("\xEF\xBB\xBF")

	w := csv.NewWriter(&buf)
	if err := w.Write([]string{"value", "value_lang", "translation", "translation_lang"}); err != nil {
		return nil, err
	}
	for _, word := range words {
		if err := w.Write([]string{
			word.Value, word.ValueLang, word.Translation, word.TranslationLang,
		}); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
