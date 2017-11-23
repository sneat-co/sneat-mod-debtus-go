package bot_shared

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/log"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/bots-api-telegram"
	"fmt"
)

func startInGroupAction(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	log.Debugf(c, "startInGroupAction()")

	var group models.Group
	if group, err = GetGroup(whc, nil); err != nil {
		return
	}
	var user bots.BotAppUser
	if user, err = whc.GetAppUser(); err != nil {
		return
	}

	appUser := user.(*models.AppUserEntity)

	var botUser bots.BotUser

	if botUser, err = whc.GetBotUserById(c, whc.Input().GetSender().GetID()); err != nil {
		return
	}

	if group, _, err = facade.Group.AddUsersToTheGroupAndOutstandingBills(c, group.ID, []facade.NewUser{
		{
			Name:       appUser.FullName(),
			BotUser:    botUser,
			ChatMember: whc.Input().GetSender(),
		},
	}); err != nil {
		return
	}
	m.Text = whc.Translate(trans.MESSAGE_TEXT_HI) +
		"\n\n" + whc.Translate(trans.SPLITUS_TEXT_HI_IN_GROUP) +
		"\n\n<b>" + whc.Translate(trans.MESSAGE_TEXT_ASK_PRIMARY_CURRENCY_FOR_GROUP) + "</b>"

	m.Format = bots.MessageFormatHTML
	m.Keyboard = CurrenciesInlineKeyboard(
		GROUP_SETTINGS_SET_CURRENCY_COMMAD + "?start=y&group=" + group.ID,
		[]tgbotapi.InlineKeyboardButton{
			{
				Text: whc.Translate(trans.BT_OTHER_CURRENCY),
				URL: fmt.Sprintf("https://t.me/%v?start=%v__group=%v", whc.GetBotCode(), GROUP_SETTINGS_CHOOSE_CURRENCY_COMMAND, group.ID),
			},
		},
	)
	return
}

func onStartCallbackInGroup(whc bots.WebhookContext, group models.Group, params BotParams) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	log.Debugf(c, "onStartCallbackInGroup()")

	if twhc, ok := whc.(*telegram_bot.TelegramWebhookContext); ok {
		if err = twhc.CreateOrUpdateTgChatInstance(); err != nil {
			return
		}
	}

	if m, err = params.InGroupWelcomeMessage(whc, group); err != nil {
		return
	}

	if _, err = whc.Responder().SendMessage(whc.Context(), m, bots.BotApiSendMessageOverHTTPS); err != nil {
		return
	}

	return showGroupMembers(whc, group, false)
}
