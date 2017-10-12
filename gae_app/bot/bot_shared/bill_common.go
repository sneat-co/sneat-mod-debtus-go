package bot_shared

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/pkg/errors"
	"github.com/strongo/app/db"
	"github.com/strongo/bots-framework/core"
	"golang.org/x/net/context"
	"net/url"
	"github.com/strongo/app/log"
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

func BillCallbackCommand(code string, f func(whc bots.WebhookContext, callbackURL *url.URL, bill models.Bill) (m bots.MessageFromBot, err error)) bots.Command {
	return bots.NewCallbackCommand(code, billCallbackAction(f))
}

func billCallbackAction(f func(whc bots.WebhookContext, callbackURL *url.URL, bill models.Bill) (m bots.MessageFromBot, err error)) func(whc bots.WebhookContext, callbackURL *url.URL) (m bots.MessageFromBot, err error) {
	return func(whc bots.WebhookContext, callbackURL *url.URL) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		var bill models.Bill
		if bill, err = getBill(c, callbackURL); err != nil {
			return
		}
		if bill.UserGroupID() == "" {
			if whc.IsInGroup() {
				if dal.DB.IsInTransaction(c) {
					var groupID string
					if groupID, err = GetUserGroupID(whc); err != nil {
						return
					}
					if err = bill.AssignToGroup(groupID); err != nil {
						return
					}
				} else {
					log.Debugf(c, "Will not update bill.UserGroupID as not in transaction")
				}
			} else {
				log.Debugf(c, "Not in group")
			}
		}
		return f(whc, callbackURL, bill)
	}
}

func transactionalCallbackCommand(c bots.Command, o db.RunOptions) bots.Command {
	c.CallbackAction = transactionalCallbackAction(o, c.CallbackAction)
	return c
}

func transactionalCallbackAction(o db.RunOptions,
	f func(whc bots.WebhookContext, callbackURL *url.URL) (m bots.MessageFromBot, err error),
) func(whc bots.WebhookContext, callbackURL *url.URL) (m bots.MessageFromBot, err error) {
	return func(whc bots.WebhookContext, callbackURL *url.URL) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		err = dal.DB.RunInTransaction(c, func(tc context.Context) error {
			whc.SetContext(tc)
			m, err = f(whc, callbackURL)
			whc.SetContext(c)
			return err
		}, o)
		return
	}
}
