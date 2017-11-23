package bot_shared

import (
	"github.com/strongo/bots-framework/core"
	"net/url"
	"bytes"
	"github.com/DebtsTracker/translations/trans"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/db"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"fmt"
	"github.com/DebtsTracker/translations/emoji"
)

const SETTINGS_COMMAND = "settings"

var settingsCommand = bots.Command{
	Code:     SETTINGS_COMMAND,
	Commands: []string{"/" + SETTINGS_COMMAND},
	Action:   NewGroupAction(func(whc bots.WebhookContext, group models.Group) (m bots.MessageFromBot, err error) {
		return GroupSettingsAction(whc, group, false)
	}),
	CallbackAction: NewGroupCallbackAction(func(whc bots.WebhookContext, callbackUrl *url.URL, group models.Group) (m bots.MessageFromBot, err error) {
		return GroupSettingsAction(whc, group, true)
	}),
}

func GroupSettingsAction(whc bots.WebhookContext, group models.Group, isEdit bool) (m bots.MessageFromBot, err error) {
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
		[]tgbotapi.InlineKeyboardButton{
			{
				Text:         whc.Translate(trans.BUTTON_TEXT_SPLIT_MODE, whc.Translate(string(group.GetSplitMode()))),
				CallbackData: GroupCallbackCommandData(GROUP_SPLIT_COMMAND, group.ID),
			},
		},
		[]tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonSwitchInlineQueryCurrentChat(
				emoji.CLIPBOARD_ICON+whc.Translate(trans.COMMAND_TEXT_NEW_BILL),
				"",
			),
		},
	)
	m.IsEdit = isEdit
	return
}

const (
	GROUP_SETTINGS_CHOOSE_CURRENCY_COMMAND = "grp-stngs-chs-ccy"
	GROUP_SETTINGS_SET_CURRENCY_COMMAD     = "grp-stngs-set-ccy"
)

var groupSettingsChooseCurrencyCommand = GroupCallbackCommand(GROUP_SETTINGS_CHOOSE_CURRENCY_COMMAND,
	func(whc bots.WebhookContext, callbackUrl *url.URL, group models.Group) (m bots.MessageFromBot, err error) {
		m.IsEdit = true
		m.Text = whc.Translate(trans.MESSAGE_TEXT_ASK_PRIMARY_CURRENCY)
		m.Keyboard = CurrenciesInlineKeyboard(
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

func groupSettingsSetCurrencyCommand(params BotParams) bots.Command {
	return bots.Command{
		Code: GROUP_SETTINGS_SET_CURRENCY_COMMAD,
		CallbackAction: TransactionalCallbackAction(db.CrossGroupTransaction, // TODO: Should be single group transaction, but a chat entity is loaded as well
			NewGroupCallbackAction(func(whc bots.WebhookContext, callbackUrl *url.URL, group models.Group) (m bots.MessageFromBot, err error) {
				currency := models.Currency(callbackUrl.Query().Get(CURRENCY_PARAM_NAME))
				if group.DefaultCurrency != currency {
					group.DefaultCurrency = currency
					if err = dal.Group.SaveGroup(whc.Context(), group); err != nil {
						return
					}
				}
				if callbackUrl.Query().Get("start") == "y" {
					return params.InGroupWelcomeMessage(whc, group)
				} else {
					return GroupSettingsAction(whc, group, true)
				}
			})),
	}
}
