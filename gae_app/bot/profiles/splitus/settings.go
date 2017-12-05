package splitus

import (
	"bytes"
	"net/url"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/shared_all"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/shared_group"
)

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
				CallbackData: GroupMembersCommandCode + "?group=" + group.ID,
			},
		},
		[]tgbotapi.InlineKeyboardButton{
			{
				Text:         whc.Translate(trans.BT_DEFAULT_CURRENCY, defaultCurrency),
				CallbackData: GroupSettingsChooseCurrencyCommandCode,
			},
		},
		[]tgbotapi.InlineKeyboardButton{
			{
				Text:         whc.Translate(trans.BUTTON_TEXT_SPLIT_MODE, whc.Translate(string(group.GetSplitMode()))),
				CallbackData: shared_group.GroupCallbackCommandData(groupSplitCommandCode, group.ID),
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

var settingsCommand = func() (settingsCommand bots.Command) {
	settingsCommand = shared_all.SettingsCommandTemplate
	settingsCommand = shared_all.SettingsCommandTemplate
	settingsCommand.Action = settingsAction
	settingsCommand.CallbackAction = func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		return settingsAction(whc)
	}
	return
}()


func settingsAction(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
	if whc.IsInGroup() {
		groupAction := shared_group.NewGroupAction(func(whc bots.WebhookContext, group models.Group) (m bots.MessageFromBot, err error) {
			return GroupSettingsAction(whc, group, false)
		})
		return groupAction(whc)
	} else {
		m, _, err = shared_all.SettingsMainTelegram(whc)
		return
	}
}