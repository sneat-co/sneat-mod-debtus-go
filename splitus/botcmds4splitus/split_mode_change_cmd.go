package botcmds4splitus

import (
	"github.com/bots-go-framework/bots-fw/botsfw"
	facade4splitus2 "github.com/sneat-co/sneat-mod-debtus-go/splitus/facade4splitus"
	models4splitus2 "github.com/sneat-co/sneat-mod-debtus-go/splitus/models4splitus"
	"github.com/strongo/logus"
	"net/url"
)

var billChangeSplitModeCommand = botsfw.Command{
	Code: "split-mode",
	CallbackAction: func(whc botsfw.WebhookContext, callbackUrl *url.URL) (m botsfw.MessageFromBot, err error) {
		ctx := whc.Context()
		logus.Debugf(ctx, "billChangeSplitModeCommand.CallbackAction()")
		var bill models4splitus2.BillEntry
		if bill.ID, err = GetBillID(callbackUrl); err != nil {
			return
		}
		tx := whc.Tx()
		if bill, err = facade4splitus2.GetBillByID(ctx, tx, bill.ID); err != nil {
			return
		}
		splitMode := models4splitus2.SplitMode(callbackUrl.Query().Get("mode"))
		if bill.Data.SplitMode != splitMode {
			bill.Data.SplitMode = splitMode
			if err = facade4splitus2.SaveBill(ctx, tx, bill); err != nil {
				return
			}
		}
		return ShowBillCard(whc, true, bill, "")
	},
}
