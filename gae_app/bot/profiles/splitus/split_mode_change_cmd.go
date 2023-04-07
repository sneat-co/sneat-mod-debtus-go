package splitus

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
	"net/url"
)

var billChangeSplitModeCommand = bots.Command{
	Code: "split-mode",
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "billChangeSplitModeCommand.CallbackAction()")
		var bill models.Bill
		if bill.ID, err = GetBillID(callbackUrl); err != nil {
			return
		}
		if err = dtdal.DB.RunInTransaction(c, func(c context.Context) (err error) {
			if bill, err = facade.GetBillByID(c, bill.ID); err != nil {
				return
			}
			splitMode := models.SplitMode(callbackUrl.Query().Get("mode"))
			if bill.SplitMode != splitMode {
				bill.SplitMode = splitMode
				if err = dtdal.Bill.SaveBill(c, bill); err != nil {
					return
				}
			}
			return
		}, dtdal.SingleGroupTransaction); err != nil {
			return
		}
		return ShowBillCard(whc, true, bill, "")
	},
}
