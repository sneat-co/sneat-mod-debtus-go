package bot_shared

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/pkg/errors"
	"github.com/strongo/app/db"
	"github.com/strongo/bots-framework/core"
	"golang.org/x/net/context"
	"net/url"
)

func GetBillMembersCallbackData(billID string) string {
	return BillCallbackCommandData(BILL_MEMBERS_COMMAND, billID)
}

func GetBillID(callbackUrl *url.URL) (billID string, err error) {
	q := callbackUrl.Query()
	sBillID := q.Get("bill")
	if sBillID == "" {
		err = errors.New("Required parameter 'bill' is not passed")
	}
	return
}

func getBill(c context.Context, callbackUrl *url.URL) (bill models.Bill, err error) {
	if bill.ID, err = GetBillID(callbackUrl); err != nil {
		return
	}
	if bill, err = dal.Bill.GetBillByID(c, bill.ID); err != nil {
		return
	}
	return
}

func BillCallbackCommand(code string, f func(whc bots.WebhookContext, callbackURL *url.URL, bill models.Bill) (m bots.MessageFromBot, err error)) bots.Command {
	return bots.NewCallbackCommand(code,
		func(whc bots.WebhookContext, callbackURL *url.URL) (m bots.MessageFromBot, err error) {
			var bill models.Bill
			if bill, err = getBill(whc.Context(), callbackURL); err != nil {
				return
			}
			return f(whc, callbackURL, bill)
		},
	)
}

func transactionalCallbackCommand(c bots.Command, o db.RunOptions) bots.Command {
	a := c.CallbackAction
	c.CallbackAction = func(whc bots.WebhookContext, callbackURL *url.URL) (m bots.MessageFromBot, err error) {
		err = dal.DB.RunInTransaction(whc.Context(), func(c context.Context) error {
			m, err = a(whc, callbackURL)
			return err
		}, o)
		return
	}
	return c
}
