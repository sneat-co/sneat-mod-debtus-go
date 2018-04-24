package splitus

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bytes"
	"fmt"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/db"
	"net/url"
)

const billSharesCommandCode = "bill_shares"

var billSharesCommand = billCallbackCommand(billSharesCommandCode, db.CrossGroupTransaction,
	func(whc bots.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m bots.MessageFromBot, err error) {
		whc.LogRequest()
		c := whc.Context()
		members := bill.GetBillMembers()
		if bill.Currency == "" {
			m.BotMessage = telegram_bot.CallbackAnswer(tgbotapi.NewCallback("", whc.Translate(trans.MESSAGE_TEXT_ASK_BILL_CURRENCY)))
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
						if err = dal.Bill.SaveBill(c, bill); err != nil {
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
