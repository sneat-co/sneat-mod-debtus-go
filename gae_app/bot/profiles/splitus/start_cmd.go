package splitus

import (
	"github.com/strongo/bots-framework/core"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"github.com/strongo/bots-api-telegram"
	"fmt"
	"github.com/DebtsTracker/translations/trans"
	"strings"
	"github.com/strongo/log"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/shared_all"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/shared_group"
	"github.com/strongo/bots-framework/platforms/telegram"
)

func startInGroupAction(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	var group models.Group
	if group, err = shared_group.GetGroup(whc, nil); err != nil {
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
	m.Keyboard = currenciesInlineKeyboard(
		GroupSettingsSetCurrencyCommandCode+"?start=y&group="+group.ID,
		[]tgbotapi.InlineKeyboardButton{
			{
				Text: whc.Translate(trans.BT_OTHER_CURRENCY),
				URL:  fmt.Sprintf("https://t.me/%v?start=%v__group=%v", whc.GetBotCode(), GroupSettingsChooseCurrencyCommandCode, group.ID),
			},
		},
	)
	return
}

func startInBotAction(whc bots.WebhookContext, startParams []string) (m bots.MessageFromBot, err error) {
	log.Debugf(whc.Context(), "splitus.startInBotAction() => startParams: %v", startParams)
	if len(startParams) > 0 {
		switch {
		case strings.HasPrefix(startParams[0], "bill-"):
			return startBillAction(whc, startParams[0])
		case startParams[0] == SettleGroupAskForCounterpartyCommandCode:
			return settleGroupStartAction(whc, startParams[1:])
		}
	}
	err = shared_all.ErrUnknownStartParam
	return
}


func onStartCallbackInGroup(whc bots.WebhookContext, group models.Group) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	log.Debugf(c, "onStartCallbackInGroup()")

	if twhc, ok := whc.(*telegram_bot.TelegramWebhookContext); ok {
		if err = twhc.CreateOrUpdateTgChatInstance(); err != nil {
			return
		}
	}

	m, err = GroupSettingsAction(whc, group, false)
	if err != nil {
		return
	}
	if _, err = whc.Responder().SendMessage(whc.Context(), m, bots.BotApiSendMessageOverHTTPS); err != nil {
		return
	}

	return whc.NewEditMessage(whc.Translate(trans.MESSAGE_TEXT_HI)+
		"\n\n"+ whc.Translate(trans.SPLITUS_TEXT_HI_IN_GROUP)+
		"\n\n"+ whc.Translate(trans.SPLITUS_TEXT_ABOUT_ME_AND_CO),
		bots.MessageFormatHTML)

	if _, err = whc.Responder().SendMessage(whc.Context(), m, bots.BotApiSendMessageOverHTTPS); err != nil {
		return
	}

	return showGroupMembers(whc, group, false)
}
