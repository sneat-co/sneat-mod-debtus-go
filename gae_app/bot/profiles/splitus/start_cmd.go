package splitus

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/shared_all"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/shared_group"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
	"strings"
)

func startInGroupAction(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	log.Debugf(c, "splitus.startInGroupAction()")
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
		err = errors.WithMessage(err, "failed to add user to the group")
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
