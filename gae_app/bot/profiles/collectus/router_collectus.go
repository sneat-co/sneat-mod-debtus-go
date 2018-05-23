package collectus

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/shared_all"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

var botParams = shared_all.BotParams{
	InBotWelcomeMessage: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		var user *models.AppUserEntity
		if user, err = shared_all.GetUser(whc); err != nil {
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
			//[]tgbotapi.InlineKeyboardButton{
			//	shared_all.NewGroupTelegramInlineButton(whc, 0),
			//},
		)
		return
	},
}

var Router = bots.NewWebhookRouter(
	map[bots.WebhookInputType][]bots.Command{},
	func() string { return "Please report any errors to @CollectusGroup" },
)

func init() {
	shared_all.AddSharedRoutes(Router, botParams)
}
