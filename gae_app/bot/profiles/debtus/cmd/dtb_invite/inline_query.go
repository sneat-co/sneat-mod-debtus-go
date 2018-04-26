package dtb_invite

import (
	"fmt"

	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/log"
)

func InlineSendInvite(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
	log.Debugf(whc.Context(), "InlineSendInvite()")
	inlineQuery := whc.Input().(bots.WebhookInlineQuery)
	//callbackData := "call-back1"
	//url := fmt.Sprintf("https://telegram.me/%v?start=invite-%v", whc.GetBotCode(), "some-code")
	botCode := whc.GetBotCode()
	m.BotMessage = telegram.InlineBotMessage(tgbotapi.InlineConfig{
		InlineQueryID: inlineQuery.GetInlineQueryID(),
		//SwitchPmText: "Accept invite",
		//SwitchPmParameter: "invite?code=ABC",
		Results: []interface{}{
			tgbotapi.InlineQueryResultArticle{
				ID:          "invite",
				Type:        "article", // ToDo: Move to constructor
				ThumbURL:    "https://debtstracker-io.appspot.com/img/debtstracker-512x512.png",
				ThumbHeight: 512,
				ThumbWidth:  512,
				Title:       fmt.Sprintf(whc.Translate(trans.INLINE_INVITE_TITLE), botCode),
				Description: whc.Translate(trans.INLINE_INVITE_DESCRIPTION),
				InputMessageContent: tgbotapi.InputTextMessageContent{
					Text: fmt.Sprintf(whc.Translate(trans.INLINE_INVITE_MESSAGE), whc.GetSender().GetFirstName(), botCode) + getMessagePleaseWaitWhileWeGenerateInviteCode(whc),
				},
				ReplyMarkup: &tgbotapi.InlineKeyboardMarkup{
					InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
						[]tgbotapi.InlineKeyboardButton{
							{Text: whc.Translate(trans.COMMAND_TEXT_WAIT_A_SECOND), CallbackData: "invite/inline-query"}, //dtb_inline.ChosenInlineResultCommand()
						},
					},
				},
			},
		},
	})
	return m, err
}

func getMessagePleaseWaitWhileWeGenerateInviteCode(whc bots.WebhookContext) string {
	return "\n\n\u23F3 " + whc.Translate(trans.MESSAGE_TEXT_PLEASE_WAIT_WHILE_WE_GENERATE_INVITE_CODE)
}
