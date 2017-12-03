package dtb_settings

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_general"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_transfer"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"github.com/pkg/errors"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/db"
	"github.com/strongo/log"
)

/*
Examples:
 receipt-{ID}-view_{LANG_CODE5}_[GA_CLIENT_ID]
*/
var reInviteCodeFromStart = regexp.MustCompile(`^(invite|receipt)-(\w+)(-(view|accept|decline))?(_(\w{2}(-\w{2})?))(_(.+))?$`)


func startCommandAction(whc bots.WebhookContext, startParam string) (m bots.MessageFromBot, err error) {
	//c := whc.Context()
	//if matched := reInviteCodeFromStart.FindStringSubmatch(startParam); matched != nil {
	//	return startByLinkCode(whc, matched)
	//}
	//
	//if chatEntity.GetPreferredLanguage() == "" {
	//	return shared_all.OnboardingAskLocaleAction(whc, whc.Translate(trans.MESSAGE_TEXT_HI)+"\n\n")
	//}
	//
	return dtb_general.MainMenuAction(whc, "", true)
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
