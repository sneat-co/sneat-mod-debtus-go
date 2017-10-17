package splitus

import (
	"github.com/strongo/bots-framework/core"
	"net/url"
	"bytes"
	"github.com/DebtsTracker/translations/trans"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/bot_shared"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/app/db"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"fmt"
)

const SETTINGS_COMMAND = "settings"

var settingsCommand = bots.Command{
	Code:     SETTINGS_COMMAND,
	Commands: []string{"/" + SETTINGS_COMMAND},
	Action:   bot_shared.NewGroupAction(func(whc bots.WebhookContext, group models.Group) (m bots.MessageFromBot, err error) {
		return groupSettingsAction(whc, group, false)
	}),
	CallbackAction: bot_shared.NewGroupCallbackAction(func(whc bots.WebhookContext, callbackUrl *url.URL, group models.Group) (m bots.MessageFromBot, err error) {
		return groupSettingsAction(whc, group, true)
	}),
}

func groupSettingsAction(whc bots.WebhookContext, group models.Group, isEdit bool) (m bots.MessageFromBot, err error) {
	var buf bytes.Buffer
	buf.WriteString(whc.Translate(trans.MT_GROUP_LABEL, group.Name))
	buf.WriteString("\n")
	buf.WriteString(whc.Translate(trans.MT_TEXT_MEMBERS_COUNT, group.MembersCount))
	m.Format = bots.MessageFormatHTML
	m.Text = buf.String()
	defaultCurrency := group.DefaultCurrency
	if defaultCurrency == "" {
		defaultCurrency = models.Currency(whc.Translate(trans.NOT_SET))
	}
	m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			{
				Text:         whc.Translate(trans.BUTTON_TEXT_MANAGE_MEMBERS),
				CallbackData: GROUP_MEMBERS_COMMAND + "?group=" + group.ID,
			},
		},
		[]tgbotapi.InlineKeyboardButton{
			{
				Text:         whc.Translate(trans.BT_DEFAULT_CURRENCY, defaultCurrency),
				CallbackData: GROUP_SETTINGS_CHOOSE_CURRENCY_COMMAND,
			},
		},
	)
	m.IsEdit = isEdit
	return
}

const (
	GROUP_SETTINGS_CHOOSE_CURRENCY_COMMAND = "group-settings-choose-currency"
	GROUP_SETTINGS_SET_CURRENCY_COMMAD     = "group-settings-set-currency"
)

var groupSettingsChooseCurrencyCommand = bot_shared.GroupCallbackCommand(GROUP_SETTINGS_CHOOSE_CURRENCY_COMMAND,
	func(whc bots.WebhookContext, callbackUrl *url.URL, group models.Group) (m bots.MessageFromBot, err error) {
		m.Text = whc.Translate(trans.MESSAGE_TEXT_ASK_PRIMARY_CURRENCY)
		m.Keyboard = bot_shared.CurrenciesInlineKeyboard(
			GROUP_SETTINGS_SET_CURRENCY_COMMAD + "?group=" + group.ID,
				[]tgbotapi.InlineKeyboardButton{
					{
						Text: whc.Translate(trans.BT_OTHER_CURRENCY),
						URL: fmt.Sprintf("https://t.me/%v?start=", whc.GetBotCode()) + GROUP_SETTINGS_CHOOSE_CURRENCY_COMMAND,
					},
				},
		)
		return
	},
)

var groupSettingsSetCurrencyCommand = bots.Command{
	Code: GROUP_SETTINGS_SET_CURRENCY_COMMAD,
	CallbackAction: bot_shared.TransactionalCallbackAction(db.CrossGroupTransaction, // TODO: Should be single group transaction, but a chat entity is loaded as well
		bot_shared.NewGroupCallbackAction(func(whc bots.WebhookContext, callbackUrl *url.URL, group models.Group) (m bots.MessageFromBot, err error) {
			currency := models.Currency(callbackUrl.Query().Get(bot_shared.CURRENCY_PARAM_NAME))
			if group.DefaultCurrency != currency {
				group.DefaultCurrency = currency
				if err = dal.Group.SaveGroup(whc.Context(), group); err != nil {
					return
				}
			}
			return groupSettingsAction(whc, group, true)
		})),
}
