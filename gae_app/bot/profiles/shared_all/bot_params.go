package shared_all

import (
	"github.com/strongo/bots-framework/core"
)

type BotParams struct {
	HelpCommandAction  bots.CommandAction
	StartInGroupAction func(whc bots.WebhookContext) (m bots.MessageFromBot, err error)
	StartInBotAction   func(whc bots.WebhookContext, startParams []string) (m bots.MessageFromBot, err error)
	//GetGroupBillCardInlineKeyboard   func(translator strongo.SingleLocaleTranslator, bill models.Bill) *tgbotapi.InlineKeyboardMarkup
	//GetPrivateBillCardInlineKeyboard func(translator strongo.SingleLocaleTranslator, botCode string, bill models.Bill) *tgbotapi.InlineKeyboardMarkup
	//OnAfterBillCurrencySelected      func(translator strongo.SingleLocaleTranslator, billID string) *tgbotapi.InlineKeyboardMarkup
	//DelayUpdateBillCardOnUserJoin    func(c context.Context, billID string, message string) error
	//ShowGroupMembers                 func(whc bots.WebhookContext, group models.Group, isEdit bool) (m bots.MessageFromBot, err error)
	//InGroupWelcomeMessage func(whc bots.WebhookContext, group models.Group) (m bots.MessageFromBot, err error)
	InBotWelcomeMessage func(whc bots.WebhookContext) (m bots.MessageFromBot, err error)

	// Below we need for sure
	SetMainMenu func(whc bots.WebhookContext, m *bots.MessageFromBot)
}
