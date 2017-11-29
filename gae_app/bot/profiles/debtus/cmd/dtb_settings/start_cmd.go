package dtb_settings

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/platforms/telegram"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_general"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_transfer"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/db"
	"github.com/strongo/log"
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
	Action: startCommandAction,
}

func startCommandAction(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
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
	case strings.HasPrefix(startParam, "refbytguser-") && startParam != "refbytguser-YOUR_CHANNEL":
		facade.Referer.AddTelegramReferrer(c, whc.AppUserIntID(), strings.TrimPrefix(startParam, "refbytguser-"), whc.GetBotCode())
	default:
		if matched := reInviteCodeFromStart.FindStringSubmatch(startParam); matched != nil {
			return startByLinkCode(whc, matched)
		}
	}

	if chatEntity.GetPreferredLanguage() == "" {
		return onboardingAskLocaleAction(whc, whc.Translate(trans.MESSAGE_TEXT_HI)+"\n\n")
	}

	return dtb_general.MainMenuAction(whc, "", true)
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

func startByLinkCode(whc bots.WebhookContext, matches []string) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	log.Debugf(c, "startByLinkCode() => matches: %v", matches)
	chatEntity := whc.ChatEntity()
	entityType := matches[1]
	entityCode := matches[2]
	operation := matches[4]
	localeCode5 := matches[6]
	//gaClientId := matches[8]
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
		return startReceipt(whc, entityCode, operation, localeCode5)
	case "invite":
		return startInvite(whc, entityCode, operation, localeCode5)
	}
	return
}

func startInvite(whc bots.WebhookContext, inviteCode, operation, localeCode5 string) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	invite, err := dal.Invite.GetInvite(c, inviteCode)
	if err == nil {
		if invite == nil {
			m = whc.NewMessage(fmt.Sprintf("Unknown invite code: %v", inviteCode))
		} else {
			log.Debugf(c, "Invite(%v): ClaimedCount=%v, MaxClaimsCount=%v", inviteCode, invite.ClaimedCount, invite.MaxClaimsCount)
			if invite.MaxClaimsCount == 0 || invite.ClaimedCount < invite.MaxClaimsCount {
				return handleInviteOnStart(whc, inviteCode, invite)
			} else {
				m = whc.NewMessage(fmt.Sprintf("Known & already claimed invite code: %v", inviteCode))
			}
		}
	}
	return m, err
}

func startReceipt(whc bots.WebhookContext, receiptCode, operation, localeCode5 string) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	receiptID, err := strconv.ParseInt(receiptCode, 10, 64)
	if err != nil {
		receiptID, err = common.DecodeID(receiptCode) // TODO: remove obsolete in a while. 2017/11/19
	} else if _, err = dal.Receipt.GetReceiptByID(c, receiptID); err != nil {
		if db.IsNotFound(err) {
			err = nil
			if receiptID, err = common.DecodeID(receiptCode); err != nil {
				return
			}
		} else {
			return
		}
	}
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
}
