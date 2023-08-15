package splitus

import (
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/dtdal"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/facade"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/models"
	"github.com/strongo/log"
	"net/url"
)

var billChangeSplitModeCommand = botsfw.Command{
	Code: "split-mode",
	CallbackAction: func(whc botsfw.WebhookContext, callbackUrl *url.URL) (m botsfw.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "billChangeSplitModeCommand.CallbackAction()")
		var bill models.Bill
		if bill.ID, err = GetBillID(callbackUrl); err != nil {
			return
		}
		tx := whc.Tx()
		if bill, err = facade.GetBillByID(c, tx, bill.ID); err != nil {
			return
		}
		splitMode := models.SplitMode(callbackUrl.Query().Get("mode"))
		if bill.Data.SplitMode != splitMode {
			bill.Data.SplitMode = splitMode
			if err = dtdal.Bill.SaveBill(c, tx, bill); err != nil {
				return
			}
		}
		return ShowBillCard(whc, true, bill, "")
	},
}
