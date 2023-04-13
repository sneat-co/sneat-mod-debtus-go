package splitus

import (
	"fmt"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/crediterra/money"
	"net/url"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"errors"
	"github.com/strongo/decimal"
	"github.com/strongo/log"
)

const (
	NEW_BILL_PARAM_I      = "i"
	NEW_BILL_PARAM_V      = "v"
	NEW_BILL_PARAM_C      = "c"
	NEW_BILL_PARAM_I_OWE  = "owe"
	NEW_BILL_PARAM_I_PAID = "paid"
)

const newBillCommandCode = "new-bill"

var newBillCommand = botsfw.Command{
	Code: newBillCommandCode,
	CallbackAction: func(whc botsfw.WebhookContext, callbackUrl *url.URL) (m botsfw.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "newBillCommand.CallbackAction(callbackUrl=%v)", callbackUrl)
		query := callbackUrl.Query()
		paramI := query.Get(NEW_BILL_PARAM_I)
		if paramI != NEW_BILL_PARAM_I_OWE && paramI != NEW_BILL_PARAM_I_PAID {
			err = errors.New("paramI != NEW_BILL_PARAM_I_OWE && paramI != NEW_BILL_PARAM_I_PAID")
			return
		}
		var amountValue, paidAmount decimal.Decimal64p2
		if amountValue, err = decimal.ParseDecimal64p2(query.Get(NEW_BILL_PARAM_V)); err != nil {
			return
		}
		if paramI == NEW_BILL_PARAM_I_PAID {
			paidAmount = amountValue
		}

		strUserID := whc.AppUserStrID()

		billEntity := models.NewBillEntity(
			models.BillCommon{
				Status:        models.BillStatusDraft,
				SplitMode:     models.SplitModeEqually,
				CreatorUserID: strUserID,
				AmountTotal:   amountValue,
				Currency:      money.Currency(query.Get("c")),
				UserIDs:       []string{strUserID},
			},
		)
		//tgMessage := whc.Input().(telegram.TelegramWebhookInput).
		//callbackQuery :=
		tgChatMessageID := fmt.Sprintf("%v@%v@%v", whc.Input().(botsfw.WebhookCallbackQuery).GetInlineMessageID(), whc.GetBotCode(), whc.Locale().Code5)
		billEntity.TgChatMessageIDs = []string{tgChatMessageID}

		var appUser botsfw.BotAppUser
		if appUser, err = whc.GetAppUser(); err != nil {
			return
		}
		user := appUser.(*models.AppUserData)
		userName := user.FullName()
		if userName == "" {
			err = errors.New("User has no name")
			return
		}

		billMember := models.BillMemberJson{
			Paid: paidAmount,
		}

		//appUserID := whc.AppUserIntID()

		if err = billEntity.SetBillMembers([]models.BillMemberJson{billMember}); err != nil {
			return
		}
		var bill models.Bill
		if bill, err = facade.Bill.CreateBill(c, c, nil, billEntity); err != nil {
			return
		}
		return ShowBillCard(whc, true, bill, "")
	},
}
