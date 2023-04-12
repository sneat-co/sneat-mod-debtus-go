package shared_all

import "github.com/bots-go-framework/bots-fw/botsfw"

type BotParams struct {
	HelpCommandAction  botsfw.CommandAction
	StartInGroupAction func(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error)
	StartInBotAction   func(whc botsfw.WebhookContext, startParams []string) (m botsfw.MessageFromBot, err error)
	//GetGroupBillCardInlineKeyboard   func(translator strongo.SingleLocaleTranslator, bill models.Bill) *tgbotapi.InlineKeyboardMarkup
	//GetPrivateBillCardInlineKeyboard func(translator strongo.SingleLocaleTranslator, botCode string, bill models.Bill) *tgbotapi.InlineKeyboardMarkup
	//OnAfterBillCurrencySelected      func(translator strongo.SingleLocaleTranslator, billID string) *tgbotapi.InlineKeyboardMarkup
	//DelayUpdateBillCardOnUserJoin    func(c context.Context, billID string, message string) error
	//ShowGroupMembers                 func(whc botsfw.WebhookContext, group models.Group, isEdit bool) (m botsfw.MessageFromBot, err error)
	//InGroupWelcomeMessage func(whc botsfw.WebhookContext, group models.Group) (m botsfw.MessageFromBot, err error)
	InBotWelcomeMessage func(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error)

	// Below we need for sure
	SetMainMenu func(whc botsfw.WebhookContext, m *botsfw.MessageFromBot)
}
