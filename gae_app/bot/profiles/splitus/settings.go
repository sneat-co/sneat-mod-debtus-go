package splitus

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/shared_all"
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/shared_group"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bytes"
	"github.com/crediterra/money"
	"github.com/sneat-co/debtstracker-translations/emoji"
	"net/url"
)

func GroupSettingsAction(whc botsfw.WebhookContext, group models.Group, isEdit bool) (m botsfw.MessageFromBot, err error) {
	var buf bytes.Buffer
	buf.WriteString(whc.Translate(trans.MT_GROUP_LABEL, group.Name))
	buf.WriteString("\n")
	buf.WriteString(whc.Translate(trans.MT_TEXT_MEMBERS_COUNT, group.MembersCount))
	m.Format = botsfw.MessageFormatHTML
	m.Text = buf.String()
	defaultCurrency := group.DefaultCurrency
	if defaultCurrency == "" {
		defaultCurrency = money.Currency(whc.Translate(trans.NOT_SET))
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

var settingsCommand = func() (settingsCommand botsfw.Command) {
	settingsCommand = shared_all.SettingsCommandTemplate
	settingsCommand.Action = settingsAction
	settingsCommand.CallbackAction = func(whc botsfw.WebhookContext, callbackUrl *url.URL) (m botsfw.MessageFromBot, err error) {
		m, err = settingsAction(whc)
		m.IsEdit = true
		return
	}
	return
}()

func settingsAction(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error) {
	if whc.IsInGroup() {
		groupAction := shared_group.NewGroupAction(func(whc botsfw.WebhookContext, group models.Group) (m botsfw.MessageFromBot, err error) {
			return GroupSettingsAction(whc, group, false)
		})
		return groupAction(whc)
	} else {
		m, _, err = shared_all.SettingsMainTelegram(whc)
		return
	}
}
