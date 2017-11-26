package collectus

import (
	"bytes"

	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/bot_shared"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

var botParams = bot_shared.BotParams{
	GetGroupBillCardInlineKeyboard:   nil,
	GetPrivateBillCardInlineKeyboard: nil,
	DelayUpdateBillCardOnUserJoin:    nil,
	OnAfterBillCurrencySelected:      nil,
	//ShowGroupMembers:                 nil,
	WelcomeText: func(translator strongo.SingleLocaleTranslator, buf *bytes.Buffer) {
		buf.WriteString(translator.Translate(trans.COLLECTUS_TEXT_HI))
		buf.WriteString("\n\n")
		buf.WriteString(translator.Translate(trans.COLLECTUS_TEXT_ABOUT_ME_AND_CO))
	},
	InGroupWelcomeMessage: func(whc bots.WebhookContext, _ models.Group) (m bots.MessageFromBot, err error) {
		return whc.NewEditMessage(whc.Translate(trans.MESSAGE_TEXT_HI)+
			"\n\n"+whc.Translate(trans.COLLECTUS_TEXT_HI_IN_GROUP)+
			"\n\n"+whc.Translate(trans.COLLECTUS_TEXT_ABOUT_ME_AND_CO),
			bots.MessageFormatHTML)
	},
	InBotWelcomeMessage: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		var user *models.AppUserEntity
		if user, err = bot_shared.GetUser(whc); err != nil {
			return
		}
		m.Text = whc.Translate(
			trans.MESSAGE_TEXT_HI_USERNAME, user.FirstName) + " " + whc.Translate(trans.COLLECTUS_TEXT_HI) +
			"\n\n" + whc.Translate(trans.COLLECTUS_TEXT_ABOUT_ME_AND_CO) +
			"\n\n" + whc.Translate(trans.COLLECTUS_TG_COMMANDS)
		m.Format = bots.MessageFormatHTML
		m.IsEdit = true

		m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
			[]tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonSwitchInlineQuery(
					whc.CommandText(trans.COMMAND_TEXT_NEW_FUNDRAISING, emoji.MEMO_ICON),
					"",
				),
			},
			[]tgbotapi.InlineKeyboardButton{
				bot_shared.NewGroupTelegramInlineButton(whc, 0),
			},
		)
		return
	},
}

var Router bots.WebhooksRouter = bots.NewWebhookRouter(
	map[bots.WebhookInputType][]bots.Command{},
	func() string { return "Please report any errors to @CollectusGroup" },
)

func init() {
	bot_shared.AddSharedRoutes(Router, botParams)
}
