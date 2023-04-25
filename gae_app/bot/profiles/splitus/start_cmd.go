package splitus

import (
	"fmt"
	"github.com/bots-go-framework/bots-api-telegram/tgbotapi"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/sneat-co/debtstracker-go/gae_app/bot/profiles/shared_all"
	"github.com/sneat-co/debtstracker-go/gae_app/bot/profiles/shared_group"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/facade"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/models"
	"github.com/sneat-co/debtstracker-translations/trans"
	"github.com/strongo/log"
	"strings"
)

func startInGroupAction(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error) {
	c := whc.Context()
	log.Debugf(c, "splitus.startInGroupAction()")
	var group models.Group
	if group, err = shared_group.GetGroup(whc, nil); err != nil {
		return
	}
	var user botsfw.BotAppUser
	if user, err = whc.GetAppUser(); err != nil {
		return
	}

	appUser := user.(*models.AppUserData)

	var botUser botsfw.BotUser

	if botUser, err = whc.GetBotUserByID(c, whc.Input().GetSender().GetID()); err != nil {
		return
	}

	if group, _, err = facade.Group.AddUsersToTheGroupAndOutstandingBills(c, group.ID, []facade.NewUser{
		{
			Name:       appUser.FullName(),
			BotUser:    botUser,
			ChatMember: whc.Input().GetSender(),
		},
	}); err != nil {
		err = fmt.Errorf("%w: failed to add user to the group", err)
		return
	}
	m.Text = whc.Translate(trans.MESSAGE_TEXT_HI) +
		"\n\n" + whc.Translate(trans.SPLITUS_TEXT_HI_IN_GROUP) +
		"\n\n<b>" + whc.Translate(trans.MESSAGE_TEXT_ASK_PRIMARY_CURRENCY_FOR_GROUP) + "</b>"

	m.Format = botsfw.MessageFormatHTML
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

func startInBotAction(whc botsfw.WebhookContext, startParams []string) (m botsfw.MessageFromBot, err error) {
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
