package splitus

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/trans"
	"github.com/crediterra/money"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
	"net/url"
)

const setBillCurrencyCommandCode = "set-bill-currency"

var setBillCurrencyCommand = billCallbackCommand(setBillCurrencyCommandCode, db.CrossGroupTransaction,
	func(whc bots.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "setBillCurrencyCommand.CallbackAction()")
		query := callbackUrl.Query()
		currencyCode := money.Currency(query.Get("currency"))
		if bill.Currency != currencyCode {
			previousCurrency := bill.Currency
			bill.Currency = currencyCode
			if err = dtdal.Bill.SaveBill(c, bill); err != nil {
				return
			}

			if bill.GetUserGroupID() != "" {
				var group models.Group
				if group, err = dtdal.Group.GetGroupByID(c, bill.GetUserGroupID()); err != nil {
					return
				}
				diff := bill.GetBalance().BillBalanceDifference(make(models.BillBalanceByMember, 0))
				if _, err = group.ApplyBillBalanceDifference(bill.Currency, diff); err != nil {
					return
				}
				if previousCurrency != "" {
					if _, err = group.ApplyBillBalanceDifference(previousCurrency, diff.Reverse()); err != nil {
						return
					}
				}
				if dtdal.Group.SaveGroup(c, group); err != nil {
					return
				}
			}
		}
		if m.Text, err = getBillCardMessageText(c, whc.GetBotCode(), whc, bill, true, whc.Translate(trans.MESSAGE_TEXT_BILL_ASK_WHO_PAID)); err != nil {
			return
		}
		m.Format = bots.MessageFormatHTML
		m.Keyboard = getWhoPaidInlineKeyboard(whc, bill.ID)
		m.IsEdit = true

		return
	},
)
