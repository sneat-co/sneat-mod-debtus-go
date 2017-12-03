package splitus

import (
	"bytes"
	"fmt"
	"net/url"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-framework/core"
	"golang.org/x/net/context"
)

const BILL_SHARES_COMMAND = "bill_shares"

var billSharesCommand = billCallbackCommand(BILL_SHARES_COMMAND,
	func(whc bots.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m bots.MessageFromBot, err error) {
		whc.LogRequest()
		c := whc.Context()
		members := bill.GetBillMembers()
		return editSplitCallbackAction(
			whc, callbackUrl,
			billCallbackCommandData(BILL_SHARES_COMMAND, bill.ID),
			billCardCallbackCommandData(bill.ID),
			trans.MESSAGE_TEXT_ASK_HOW_TO_SPLIT_IN_GROP,
			members,
			bill.TotalAmount(),
			func(buffer *bytes.Buffer) error {
				return writeBillCardTitle(c, bill, whc.GetBotCode(), buffer, whc)
			},
			func(memberID string, addValue int) (member models.BillMemberJson, err error) {
				err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
					if bill, err = dal.Bill.GetBillByID(c, bill.ID); err != nil {
						return
					}
					for i, m := range members {
						if m.ID == memberID {
							m.Shares += addValue
							if m.Shares < 0 {
								m.Shares = 0
							}
							members[i] = m
							bill.SplitMode = models.SplitModeShare
							if err = bill.SetBillMembers(members); err != nil {
								return
							}
							if err = dal.Bill.SaveBill(c, bill); err != nil {
								return
							}
							member = m
							return err
						}
					}
					return fmt.Errorf("member not found by ID: %v", member.ID)
				}, dal.CrossGroupTransaction)
				return
			},
		)
	},
)
