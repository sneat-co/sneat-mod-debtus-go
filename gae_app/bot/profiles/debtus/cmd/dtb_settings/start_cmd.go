package dtb_settings

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_general"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_transfer"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/app/log"
	"github.com/strongo/bots-framework/core"
	"regexp"
	"strings"
	"time"
	"github.com/strongo/bots-api-telegram"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/platforms/telegram"
)

/*
Examples:
 receipt-{ID}-view_{LANG_CODE5}_[GA_CLIENT_ID]
*/
var reInviteCodeFromStart = regexp.MustCompile(`^(invite|receipt)-(\w+)(-(view|accept|decline))?(_(\w{2}(-\w{2})?))(_(.+))?$`)

var StartCommand = bots.Command{
	Code:     "start",
	Commands: trans.Commands(trans.COMMAND_START),
	Title:    "/start",
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		c := whc.Context()

		log.Debugf(c, "StartCommand.Action()")

		chatEntity := whc.ChatEntity()
		chatEntity.SetAwaitingReplyTo("")

		startParam, _ := telegram.ParseStartCommand(whc)
		switch {
		case startParam == "help_inline":
			return startInlineHelp(whc)
		case strings.HasPrefix(startParam, "login-"):
			loginID, err := common.DecodeID(startParam[len("login-"):])
			if err != nil {
				return m, err
			}
			return startLoginGac(whc, loginID)
			//case strings.HasPrefix(textToMatchNoStart, JOIN_BILL_COMMAND):
			//	return JoinBillCommand.Action(whc)
		default:
			if matched := reInviteCodeFromStart.FindStringSubmatch(startParam); matched != nil {
				return startInviteCode(whc, matched)
			}
		}

		if chatEntity.GetPreferredLanguage() == "" {
			return onboardingAskLocaleAction(whc, whc.Translate(trans.MESSAGE_TEXT_HI) + "\n\n")
		}

		return dtb_general.MainMenuAction(whc, "", true)
	},
}

func startInlineHelp(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
	m = whc.NewMessage("<b>Help: How to use this bot in chats</b>\n\nExplain here how to use bot's inline mode.")
	m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{{Text: "Button 1", URL: "https://debtstracker.io/#btn=1"}},
		[]tgbotapi.InlineKeyboardButton{{Text: "Button 2", URL: "https://debtstracker.io/#btn=2"}},
		//[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonSwitch("Back to chat 1", "1")},
		//[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonSwitch("Back to chat 2", "2")},
		[]tgbotapi.InlineKeyboardButton{{Text: "Button 3", CallbackData: "help-3"}},
		[]tgbotapi.InlineKeyboardButton{{Text: "Button 4", CallbackData: "help-4"}},
		[]tgbotapi.InlineKeyboardButton{{Text: "Button 5", CallbackData: "help-5"}},
	)
	return m, err
}

func startLoginGac(whc bots.WebhookContext, loginID int64) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	var loginPin models.LoginPin
	if loginPin, err = facade.AuthFacade.AssignPinCode(c, loginID, whc.AppUserIntID()); err != nil {
		return
	}
	return whc.NewMessageByCode(trans.MESSAGE_TEXT_LOGIN_CODE, models.LoginCodeToString(loginPin.Code)), nil
}

func startInviteCode(whc bots.WebhookContext, matched []string) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	log.Debugf(c, "operation: %v, invite: %v", matched[1], matched[2])
	chatEntity := whc.ChatEntity()
	entityType := matched[1]
	entityCode := matched[2]
	operation := matched[4]
	localeCode5 := matched[6]
	gaClientId := matched[8]
	log.Debugf(c, "gaClientId: [%v]", gaClientId)
	if localeCode5 != "" {
		if len(localeCode5) == 2 {
			localeCode5 = common.Locale2to5(localeCode5)
		}
		whc.SetLocale(localeCode5)
		chatEntity.SetPreferredLanguage(localeCode5)
		if err = dal.User.DelaySetUserPreferredLocale(c, time.Second, whc.AppUserIntID(), localeCode5); err != nil {
			return
		}

	}
	switch entityType {
	case "receipt":
		receiptID, err := common.DecodeID(entityCode)
		if err != nil {
			return m, errors.Wrap(err, "Failed to decode receipt ID")
		}
		switch operation {
		case "view":
			whc.SetLocale(localeCode5)
			return dtb_transfer.ShowReceipt(whc, receiptID)
		default:
			return dtb_transfer.AcknowledgeReceipt(whc, receiptID, operation)
		}
	case "invite":
		invite, err := dal.Invite.GetInvite(c, entityCode)
		if err == nil {
			if invite == nil {
				m = whc.NewMessage(fmt.Sprintf("Unknown invite code: %v", entityCode))
			} else {
				log.Debugf(c, "Invite(%v): ClaimedCount=%v, MaxClaimsCount=%v", entityCode, invite.ClaimedCount, invite.MaxClaimsCount)
				if invite.MaxClaimsCount == 0 || invite.ClaimedCount < invite.MaxClaimsCount {
					return handleInviteOnStart(whc, entityCode, invite)
				} else {
					m = whc.NewMessage(fmt.Sprintf("Known & already claimed invite code: %v", entityCode))
				}
			}
		}
		return m, err
	}
	return
}