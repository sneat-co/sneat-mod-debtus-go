package splitus

import (
	"github.com/strongo/bots-framework/core"
	"net/url"
	"github.com/strongo/app/log"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
	"golang.org/x/net/context"
	"bitbucket.com/debtstracker/gae_app/debtstracker/dal"
	"bitbucket.com/debtstracker/gae_app/bot/bot_shared"
)

var billChangeSplitModeCommand = bots.Command{
	Code: "split-mode",
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "billChangeSplitModeCommand.CallbackAction()")
		var bill models.Bill
		if bill.ID, err = bot_shared.GetBillID(callbackUrl); err != nil {
			return
		}
		if err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
			if bill, err = dal.Bill.GetBillByID(c, bill.ID); err != nil {
				return
			}
			splitMode := models.SplitMode(callbackUrl.Query().Get("mode"))
			if bill.SplitMode != splitMode {
				bill.SplitMode = splitMode
				if err = dal.Bill.UpdateBill(c, bill); err != nil {
					return
				}
			}
			return
		}, dal.SingleGroupTransaction); err != nil {
			return
		}
		return bot_shared.ShowBillCard(whc, botParams, true, bill, "")
	},
}
