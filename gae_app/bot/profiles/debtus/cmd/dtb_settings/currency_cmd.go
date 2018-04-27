package dtb_settings

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
)

const ASK_CURRENCY_SETTING_COMMAND = "ask-currency-settings"

var AskCurrencySettingsCommand = bots.Command{
	Code:     ASK_CURRENCY_SETTING_COMMAND,
	Replies:  []bots.Command{SetPrimaryCurrency},
	Commands: []string{"\xF0\x9F\x92\xB1"},
	Icon:     emoji.CURRENCY_EXCAHNGE_ICON,
	Title:    trans.COMMAND_TEXT_SETTINGS_PRIMARY_CURRENCY,
	Action: func(whc bots.WebhookContext) (bots.MessageFromBot, error) {
		m := whc.NewMessageByCode(trans.MESSAGE_TEXT_ASK_PRIMARY_CURRENCY)
		m.Keyboard = tgbotapi.NewReplyKeyboardUsingStrings([][]string{
			{
				"€ - Euro ",
				"$ - USD",
				"₽ - RUB",
			},
			{
				"Other",
			},
		})
		whc.ChatEntity().SetAwaitingReplyTo(ASK_CURRENCY_SETTING_COMMAND)
		return m, nil
	},
}

const SET_PRIMARY_CURRENCY_COMMAND = "set-primary-currency"

var SetPrimaryCurrency = bots.Command{
	Code: SET_PRIMARY_CURRENCY_COMMAND,
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "SetPrimaryCurrency.Action()")
		whc.ChatEntity().SetAwaitingReplyTo("")
		primaryCurrency := whc.Input().(bots.WebhookTextMessage).Text()
		if err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
			var user models.AppUser
			if user, err = dal.User.GetUserByID(c, whc.AppUserIntID()); err != nil {
				return
			}
			user.PrimaryCurrency = primaryCurrency
			return dal.User.SaveUser(c, user)
		}, nil); err != nil {
			return
		}
		return whc.NewMessageByCode(trans.MESSAGE_TEXT_PRIMARY_CURRENCY_IS_SET_TO, whc.Input().(bots.WebhookTextMessage).Text()), nil
	},
}
