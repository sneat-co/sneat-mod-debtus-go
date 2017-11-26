package bot_shared

import (
	"net/url"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/pkg/errors"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/db"
	"github.com/strongo/log"
	"golang.org/x/net/context"
)

func GetBillMembersCallbackData(billID string) string {
	return BillCallbackCommandData(BILL_MEMBERS_COMMAND, billID)
}

func GetBillID(callbackUrl *url.URL) (billID string, err error) {
	if billID = callbackUrl.Query().Get("bill"); billID == "" {
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

type BillCallbackAction func(whc bots.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m bots.MessageFromBot, err error)

func BillCallbackCommand(code string, f BillCallbackAction) bots.Command {
	return bots.NewCallbackCommand(code, billCallbackAction(f))
}

func billCallbackAction(f BillCallbackAction) func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
	return func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		var bill models.Bill
		if bill, err = getBill(c, callbackUrl); err != nil {
			return
		}
		if bill.UserGroupID() == "" {
			if whc.IsInGroup() {
				if dal.DB.IsInTransaction(c) {
					var group models.Group
					if group.ID, err = GetUserGroupID(whc); err != nil {
						return
					}
					if bill, group, err = facade.Bill.AssignBillToGroup(c, bill, group.ID, whc.AppUserStrID()); err != nil {
						return
					}
				} else {
					log.Debugf(c, "Will not update bill.UserGroupID as not in transaction")
				}
			} else {
				log.Debugf(c, "Not in group")
			}
		}
		return f(whc, callbackUrl, bill)
	}
}

func TransactionalCallbackCommand(c bots.Command, o db.RunOptions) bots.Command {
	c.CallbackAction = TransactionalCallbackAction(o, c.CallbackAction)
	return c
}

func TransactionalCallbackAction(o db.RunOptions,
	f func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error),
) func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
	return func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		err = dal.DB.RunInTransaction(c, func(tc context.Context) error {
			whc.SetContext(tc)
			m, err = f(whc, callbackUrl)
			whc.SetContext(c)
			return err
		}, o)
		return
	}
}
