package handlers

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"langs/internal/consts"
	"langs/internal/tg_bot/helper"
	"strconv"
)

func (h *TGHandler) OnTestMe(ctx context.Context, b *bot.Bot, update *models.Update) {
	user := h.getUserFromContext(ctx, b, update)

	h.wordService.SendTest(user, h.OnTestAnswer)
}

func (h *TGHandler) OnTestAnswer(
	ctx context.Context,
	b *bot.Bot,
	mes models.MaybeInaccessibleMessage,
	answer []byte,
) {
	answerData := helper.StringToData(string(answer))
	valueId, ok := answerData[0].(int)
	if !ok {
		panic("answerData[0] is not an int")
	}

	wordId := int64(valueId)
	word, err := h.wordRepository.First(wordId)

	if err != nil {

	}

	valueString, ok := answerData[1].(string)

	if !ok {
		panic("answerData[0] is not an int")
	}

	//user := h.getUserFromContextMsg(ctx, b, mes)

	fmt.Println("VVVVVVVVV", word, valueString)

	if word.Value == valueString || word.Translation == valueString {
		h.tgMessage.SendOrEditMessage(
			ctx,
			mes.Message.Chat.ID,
			mes.Message.ID,
			consts.LangFlags[word.ValueLang]+" "+
				""+word.Value+" - "+consts.LangFlags[word.TranslationLang]+" "+word.Translation+" ✅ ("+strconv.Itoa(int(word.Rate))+"/"+"10"+")",
			nil,
		)

		return
	}

	h.tgMessage.SendOrEditMessage(
		ctx,
		mes.Message.Chat.ID,
		mes.Message.ID,
		consts.LangFlags[word.ValueLang]+" "+
			""+word.Value+" - "+consts.LangFlags[word.TranslationLang]+" "+word.Translation+" ❌ ("+strconv.Itoa(int(word.Rate))+"/"+"10"+")",
		nil,
	)

	//h.tgMessage.SendOrEditMessage(ctx, mes.Message.Chat.ID, mes.Message.ID, string(answer), nil)
	//user, err := h.userService.GetUserFromContext(ctx)
	//if err != nil {
	//	h.handleError(ctx, b, mes.Message.Chat.ID, "User retrieval failed")
	//	return
	//}

}
