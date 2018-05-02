package splitus

import (
	"net/url"

	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/shared_all"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/shared_group"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/pkg/errors"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/db"
	"github.com/strongo/log"
)

func GetBillMembersCallbackData(billID string) string {
	return billCallbackCommandData(billMembersCommandCode, billID)
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
	if bill, err = facade.GetBillByID(c, bill.ID); err != nil {
		return
	}
	return
}

type billCallbackActionType func(whc bots.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m bots.MessageFromBot, err error)

func billCallbackCommand(code string, txOptions db.RunOptions, f billCallbackActionType) (command bots.Command) {
	command = bots.NewCallbackCommand(code, billCallbackAction(f))
	if txOptions != nil {
		command.CallbackAction = shared_all.TransactionalCallbackAction(txOptions, command.CallbackAction)
	}
	return
}

func billCallbackAction(f billCallbackActionType) func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
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
					if group.ID, err = shared_group.GetUserGroupID(whc); err != nil {
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
