package splitus

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/crediterra/money"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/debtstracker-translations/trans"
	"github.com/strongo/log"
	"net/url"
)

const setBillCurrencyCommandCode = "set-bill-currency"

var setBillCurrencyCommand = billCallbackCommand(setBillCurrencyCommandCode,
	func(whc botsfw.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m botsfw.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "setBillCurrencyCommand.CallbackAction()")
		query := callbackUrl.Query()
		currencyCode := money.Currency(query.Get("currency"))
		if bill.Data.Currency != currencyCode {
			previousCurrency := bill.Data.Currency
			bill.Data.Currency = currencyCode
			if err = dtdal.Bill.SaveBill(c, bill); err != nil {
				return
			}

			if bill.Data.GetUserGroupID() != "" {
				err = whc.RunReadwriteTransaction(c, func(ctx context.Context, tx dal.ReadwriteTransaction) (err error) {
					var group models.Group
					if group, err = dtdal.Group.GetGroupByID(c, tx, bill.Data.GetUserGroupID()); err != nil {
						return
					}
					diff := bill.Data.GetBalance().BillBalanceDifference(make(models.BillBalanceByMember, 0))
					if _, err = group.Data.ApplyBillBalanceDifference(bill.Data.Currency, diff); err != nil {
						return
					}
					if previousCurrency != "" {
						if _, err = group.Data.ApplyBillBalanceDifference(previousCurrency, diff.Reverse()); err != nil {
							return
						}
					}
					if dtdal.Group.SaveGroup(c, tx, group); err != nil {
						return
					}
					return nil
				})
				if err != nil {
					return
				}
			}
		}
		if m.Text, err = getBillCardMessageText(c, whc.GetBotCode(), whc, bill, true, whc.Translate(trans.MESSAGE_TEXT_BILL_ASK_WHO_PAID)); err != nil {
			return
		}
		m.Format = botsfw.MessageFormatHTML
		m.Keyboard = getWhoPaidInlineKeyboard(whc, bill.ID)
		m.IsEdit = true

		return
	},
)
