package splitus

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bytes"
	"fmt"
	"github.com/bots-go-framework/bots-fw-telegram"
	"net/url"
)

const billSharesCommandCode = "bill_shares"

var billSharesCommand = billCallbackCommand(billSharesCommandCode, db.CrossGroupTransaction,
	func(whc botsfw.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m botsfw.MessageFromBot, err error) {
		whc.LogRequest()
		c := whc.Context()
		members := bill.GetBillMembers()
		if bill.Currency == "" {
			m.BotMessage = telegram.CallbackAnswer(tgbotapi.NewCallback("", whc.Translate(trans.MESSAGE_TEXT_ASK_BILL_CURRENCY)))
			return
		}
		var billID string
		if bill.MembersCount <= 1 {
			billID = bill.ID
		}
		return editSplitCallbackAction(
			whc, callbackUrl,
			billID,
			billCallbackCommandData(billSharesCommandCode, bill.ID),
			billCardCallbackCommandData(bill.ID),
			trans.MESSAGE_TEXT_ASK_HOW_TO_SPLIT_IN_GROP,
			members,
			bill.TotalAmount(),
			func(buffer *bytes.Buffer) error {
				return writeBillCardTitle(c, bill, whc.GetBotCode(), buffer, whc)
			},
			func(memberID string, addValue int) (member models.BillMemberJson, err error) {
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
						if err = dtdal.Bill.SaveBill(c, bill); err != nil {
							return
						}
						member = m
						return
					}
				}
				err = fmt.Errorf("member not found by ID: %v", member.ID)
				return
			},
		)
	},
)
