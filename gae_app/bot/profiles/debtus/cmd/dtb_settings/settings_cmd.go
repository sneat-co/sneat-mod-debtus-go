package dtb_settings

import (
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"net/url"
	"bitbucket.com/debtstracker/gae_app/debtstracker/dal"
	"golang.org/x/net/context"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
)

const SETTINGS_CALLBACK_PATH = "settings"

var SettingsCommand = bots.Command{
	Code:     "settings",
	Title:    trans.COMMAND_TEXT_SETTING,
	Icon:     emoji.SETTINGS_ICON,
	Commands: trans.Commands(trans.COMMAND_SETTINGS, emoji.SETTINGS_ICON),
	Action:   SettingsAction,
}

var BackToSettingsCallbackCommand = bots.NewCallbackCommand(SETTINGS_CALLBACK_PATH, backToSettingsCallbackAction)

func backToSettingsCallbackAction(whc bots.WebhookContext, _ *url.URL) (bots.MessageFromBot, error) {
	return BackToSettingsAction(whc, "")
}

func SettingsAction(whc bots.WebhookContext) (bots.MessageFromBot, error) {
	return BackToSettingsAction(whc, "")
}

func BackToSettingsAction(whc bots.WebhookContext, messageText string) (m bots.MessageFromBot, err error) {
	if messageText == "" {
		messageText = whc.Translate(trans.MESSAGE_TEXT_SETTINGS)
	} else {
		messageText += "\n\n" + whc.Translate(trans.MESSAGE_TEXT_SETTINGS)
	}
	m = whc.NewMessage(messageText)
	m.IsEdit = whc.InputType() == bots.WebhookInputCallbackQuery
	m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			//{AskCurrencySettingsCommand.DefaultTitle(whc)},
			{
				Text:         whc.CommandText(trans.COMMAND_TEXT_LANGUAGE, emoji.EARTH_ICON),
				CallbackData: SETTINGS_LOCALE_LIST_CALLBACK_PATH,
			},
			{
				Text:         whc.CommandText(trans.COMMAND_TEXT_CONTACTS, ""),
				CallbackData: CONTACTS_LIST_COMMAND,
			},
			//{
			//	emoji.NO_ENTRY_SIGN_ICON + " Мои данные",
			//	emoji.NO_ENTRY_SIGN_ICON + " Мой аккаунт",
			//},
			//{Text: dtb_general.MainMenuCommand.DefaultTitle(whc), CallbackData: "main-menu"},
		},
	)
	return m, err
}


var FixBalanceCommand = bots.Command{
	Code: "fixbalance",
	Commands: []string{"/fixbalance"},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		if err = dal.DB.RunInTransaction(whc.Context(), func(c context.Context) error {
			user, err := dal.User.GetUserByID(c, whc.AppUserIntID())
			if err != nil {
				return err
			}
			contacts := user.Contacts()
			balance := make(models.Balance, user.BalanceCount)
			for _, contact := range contacts {
				b, err := contact.Balance()
				if err != nil {
					return err
				}
				for k, v := range b {
					balance[k] += v
				}
			}
			user.SetBalance(balance)
			return dal.User.SaveUser(c, user)
		}, dal.CrossGroupTransaction); err != nil {
			return
		}
		m = whc.NewMessage("Balance fixed")
		return
	},
}