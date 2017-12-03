package splitus

import (
	"net/url"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/shared_all"
)

const setBillCurrencyCommandCode = "set-bill-currency"

var setBillCurrencyCommand = shared_all.TransactionalCallbackCommand(billCallbackCommand(setBillCurrencyCommandCode,
	func(whc bots.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "setBillCurrencyCommand.CallbackAction()")
		query := callbackUrl.Query()
		currencyCode := models.Currency(query.Get("currency"))
		if bill.Currency != currencyCode {
			previousCurrency := bill.Currency
			bill.Currency = currencyCode
			if err = dal.Bill.SaveBill(c, bill); err != nil {
				return
			}

			if bill.UserGroupID() != "" {
				var group models.Group
				if group, err = dal.Group.GetGroupByID(c, bill.UserGroupID()); err != nil {
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
				if dal.Group.SaveGroup(c, group); err != nil {
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
), dal.CrossGroupTransaction)
