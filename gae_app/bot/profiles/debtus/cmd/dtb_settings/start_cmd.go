package dtb_settings

import (
	"fmt"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/debtstracker-go/gae_app/bot/profiles/debtus/cmd/dtb_transfer"
	"github.com/sneat-co/debtstracker-go/gae_app/bot/profiles/shared_all"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/common"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/dtdal"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/models"
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
		if err = whc.SetLocale(localeCode5); err != nil {
			return
		}
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
	var invite models.Invite
	if invite, err = dtdal.Invite.GetInvite(c, nil, inviteCode); err != nil {
		if dal.IsNotFound(err) {
			return whc.NewMessage(fmt.Sprintf("Unknown invite code: %v", inviteCode)), nil
		}
		return
	}
	log.Debugf(c, "Invite(%v): ClaimedCount=%v, MaxClaimsCount=%v", inviteCode, invite.Data.ClaimedCount, invite.Data.MaxClaimsCount)
	if invite.Data.MaxClaimsCount == 0 || invite.Data.ClaimedCount < invite.Data.MaxClaimsCount {
		return handleInviteOnStart(whc, inviteCode, invite)
	} else {
		m = whc.NewMessage(fmt.Sprintf("Known & already claimed invite code: %v", inviteCode))
	}
	return m, err
}

func startReceipt(whc botsfw.WebhookContext, receiptCode, operation, localeCode5 string) (m botsfw.MessageFromBot, err error) {
	c := whc.Context()
	var receiptID int
	if receiptID, err = strconv.Atoi(receiptCode); err != nil {
		if receiptID, err = common.DecodeIntID(receiptCode); err != nil { // TODO: remove obsolete in a while. 2017/11/19
			return
		}
	} else if _, err = dtdal.Receipt.GetReceiptByID(c, nil, receiptID); err != nil {
		if dal.IsNotFound(err) {
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
