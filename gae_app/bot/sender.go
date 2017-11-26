package bot

import (
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/log"
)

func SendRefreshOrNothingChanged(whc bots.WebhookContext, m bots.MessageFromBot) (m2 bots.MessageFromBot, err error) {
	c := whc.Context()
	if _, err = whc.Responder().SendMessage(c, m, bots.BotApiSendMessageOverHTTPS); err != nil {
		log.Debugf(c, "error type: %T", err)
		if apiResponse, ok := err.(tgbotapi.APIResponse); ok && apiResponse.ErrorCode == 400 {
			m.BotMessage = telegram_bot.CallbackAnswer(tgbotapi.NewCallback("", whc.Translate(trans.ALERT_TEXT_NOTHING_CHANGED)))
			err = nil
		}
	}
	return m, err
}
