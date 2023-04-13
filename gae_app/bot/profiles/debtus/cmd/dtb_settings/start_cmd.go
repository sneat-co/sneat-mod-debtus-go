package dtb_settings

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_transfer"
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/shared_all"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"fmt"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/dal-go/dalgo/dal"
	"github.com/strongo/log"
	"regexp"
	"strconv"
	"time"
)

/*
Examples:

	receipt-{ID}-view_{LANG_CODE5}_[GA_CLIENT_ID]
*/
var reInviteOrReceiptCodeFromStart = regexp.MustCompile(`^(invite|receipt)-(\w+)(-(view|accept|decline))?(_(\w{2}(-\w{2})?))(_(.+))?$`)

func StartInBotAction(whc botsfw.WebhookContext, startParams []string) (m botsfw.MessageFromBot, err error) {
	if len(startParams) == 1 {
		if matched := reInviteOrReceiptCodeFromStart.FindStringSubmatch(startParams[0]); matched != nil {
			return startByLinkCode(whc, matched)
		}
	}
	err = shared_all.ErrUnknownStartParam
	return
}

func startByLinkCode(whc botsfw.WebhookContext, matches []string) (m botsfw.MessageFromBot, err error) {
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
		if err = dtdal.User.DelaySetUserPreferredLocale(c, time.Second, whc.AppUserIntID(), localeCode5); err != nil {
			return
		}

	}
	switch entityType {
	case "receipt":
		return startReceipt(whc, entityCode, operation, localeCode5)
	case "invite":
		return startInvite(whc, entityCode, operation, localeCode5)
	default:
		err = shared_all.ErrUnknownStartParam
	}
	return
}

func startInvite(whc botsfw.WebhookContext, inviteCode, operation, localeCode5 string) (m botsfw.MessageFromBot, err error) {
	c := whc.Context()
	invite, err := dtdal.Invite.GetInvite(c, inviteCode)
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

func startReceipt(whc botsfw.WebhookContext, receiptCode, operation, localeCode5 string) (m botsfw.MessageFromBot, err error) {
	c := whc.Context()
	receiptID, err := strconv.Atoi(receiptCode)
	if err != nil {
		receiptID, err = common.DecodeIntID(receiptCode) // TODO: remove obsolete in a while. 2017/11/19
	} else if _, err = dtdal.Receipt.GetReceiptByID(c, nil, receiptID); err != nil {
		if dal.IsNotFound(err) {
			err = nil
			if receiptID, err = common.DecodeIntID(receiptCode); err != nil {
				err = fmt.Errorf("failed to decode receipt ID: %w", err)
				return
			}
		} else {
			return
		}
	}
	switch operation {
	case "view":
		if err = whc.SetLocale(localeCode5); err != nil {
			return
		}
		return dtb_transfer.ShowReceipt(whc, receiptID)
	default:
		return dtb_transfer.AcknowledgeReceipt(whc, receiptID, operation)
	}
}
