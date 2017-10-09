package bot_shared

import (
	"github.com/strongo/bots-framework/core"
	"net/url"
	"github.com/strongo/app/log"
	"github.com/pkg/errors"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/decimal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"fmt"
)

const (
	NEW_BILL_PARAM_I = "i"
	NEW_BILL_PARAM_V = "v"
	NEW_BILL_PARAM_C = "c"
	NEW_BILL_PARAM_I_OWE = "owe"
	NEW_BILL_PARAM_I_PAID = "paid"
)

const NEW_BILL_COMMAND = "new-bill"
func NewBillCommand(params BotParams) bots.Command {
	return bots.Command{
		Code: NEW_BILL_COMMAND,
		CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
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

			billEntity := models.NewBillEntity(
				models.BillCommon{
					Status:        models.BillStatusDraft,
					SplitMode:     models.SplitModeEqually,
					CreatorUserID: whc.AppUserIntID(),
					AmountTotal:   amountValue,
					Currency:      query.Get("c"),
					UserIDs:       []int64{whc.AppUserIntID()},
				},
			)
			//tgMessage := whc.Input().(telegram_bot.TelegramWebhookInput).
			//callbackQuery :=
			tgChatMessageID := fmt.Sprintf("%v@%v@%v", whc.Input().(bots.WebhookCallbackQuery).GetInlineMessageID(), whc.GetBotCode(), whc.Locale().Code5)
			billEntity.TgChatMessageIDs = []string{tgChatMessageID}

			var appUser bots.BotAppUser
			if appUser, err = whc.GetAppUser(); err != nil {
				return
			}
			user := appUser.(*models.AppUserEntity)
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
			if bill, err = dal.Bill.InsertBillEntity(c, billEntity); err != nil {
				return
			}
			return ShowBillCard(whc, params, true, bill, "")
		},
	}
}
