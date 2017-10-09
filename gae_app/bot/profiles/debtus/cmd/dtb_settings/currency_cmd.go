package dtb_settings

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app/log"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
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
	Action: func(whc bots.WebhookContext) (bots.MessageFromBot, error) {

		whc.ChatEntity().SetAwaitingReplyTo("")
		userEntity, err := whc.GetAppUser()
		if err == nil {
			user, _ := userEntity.(*models.AppUserEntity)
			user.PrimaryCurrency = whc.Input().(bots.WebhookTextMessage).Text()
			err = whc.SaveAppUser(whc.AppUserIntID(), userEntity)
			if err != nil {
				log.Errorf(whc.Context(), "Failed to update user: %v", err)
			}
		} else {
			log.Errorf(whc.Context(), "Failed to get user: %v", err)
		}
		return whc.NewMessageByCode(trans.MESSAGE_TEXT_PRIMARY_CURRENCY_IS_SET_TO, whc.Input().(bots.WebhookTextMessage).Text()), nil
	},
}
