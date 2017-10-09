package splitus

import (
	"github.com/strongo/bots-framework/core"
	"net/url"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
	"fmt"
	"github.com/DebtsTracker/translations/trans"
	"golang.org/x/net/context"
	"bitbucket.com/debtstracker/gae_app/debtstracker/dal"
	"bytes"
	"bitbucket.com/debtstracker/gae_app/bot/bot_shared"
)

const BILL_SPLIT_COMMAND = "bill-split"

var billSplitCommand = bot_shared.BillCallbackCommand(BILL_SPLIT_COMMAND,
	func(whc bots.WebhookContext, callbackURL *url.URL, bill models.Bill) (m bots.MessageFromBot, err error) {
		whc.LogRequest()
		c := whc.Context()
		return editSplitCallbackAction(
			whc, callbackURL,
			bot_shared.BillCallbackCommandData(BILL_SPLIT_COMMAND, bill.ID),
			bot_shared.BillCardCallbackCommandData(bill.ID),
			trans.MESSAGE_TEXT_ASK_HOW_TO_SPLIT_IN_GROP,
			bill.GetMembers(),
			bill.TotalAmount(),
			func(buffer *bytes.Buffer) error {
				return bot_shared.WriteBillCardTitle(c, bill, whc.GetBotCode(), buffer, whc)
			},
			func(memberID string, addValue int) (member models.MemberJson, err error) {
				err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
					if bill, err = dal.Bill.GetBillByID(c, bill.ID); err != nil {
						return
					}
					members := bill.GetBillMembers()
					for i, m := range members {
						if m.ID == memberID {
							m.Shares += addValue
							if m.Shares < 0 {
								m.Shares = 0
							}
							members[i] = m
							bill.SetBillMembers(members)
							if err = dal.Bill.UpdateBill(c, bill); err != nil {
								return
							}
							member = m.MemberJson
							return err
						}
					}
					return fmt.Errorf("member not found by ID: %d", member.ID)
				}, dal.CrossGroupTransaction)
				return
			},
		)
	},
)
