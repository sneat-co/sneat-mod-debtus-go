package dtb_inline

import (
	"github.com/strongo/log"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
)

func InlineEmptyQuery(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
	log.Debugf(whc.Context(), "InlineEmptyQuery()")
	inlineQuery := whc.Input().(bots.WebhookInlineQuery)
	m.BotMessage = telegram_bot.InlineBotMessage(tgbotapi.InlineConfig{
		InlineQueryID:     inlineQuery.GetInlineQueryID(),
		CacheTime:         60,
		SwitchPMText:      "Help: How to use this bot?",
		SwitchPMParameter: "help_inline",
	})
	return m, err
}
