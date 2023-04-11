package splitus

import (
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/dal-go/dalgo/dal"
	"net/url"

	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/shared_group"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"errors"
	"github.com/strongo/log"
)

func GetBillMembersCallbackData(billID string) string {
	return billCallbackCommandData(billMembersCommandCode, billID)
}

func GetBillID(callbackUrl *url.URL) (billID string, err error) {
	if billID = callbackUrl.Query().Get("bill"); billID == "" {
		err = errors.New("required parameter 'bill' is not passed")
	}
	return
}

func getBill(c context.Context, tx dal.ReadSession, callbackUrl *url.URL) (bill models.Bill, err error) {
	if bill.ID, err = GetBillID(callbackUrl); err != nil {
		return
	}
	if bill, err = facade.GetBillByID(c, tx, bill.ID); err != nil {
		return
	}
	return
}

type billCallbackActionType func(whc botsfw.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m botsfw.MessageFromBot, err error)

func billCallbackCommand(code string, f billCallbackActionType) (command botsfw.Command) {
	command = botsfw.NewCallbackCommand(code, billCallbackAction(f))
	//if txOptions != nil {
	//	command.CallbackAction = shared_all.TransactionalCallbackAction(txOptions, command.CallbackAction)
	//}
	return
}

func billCallbackAction(f billCallbackActionType) func(whc botsfw.WebhookContext, callbackUrl *url.URL) (m botsfw.MessageFromBot, err error) {
	return func(whc botsfw.WebhookContext, callbackUrl *url.URL) (m botsfw.MessageFromBot, err error) {
		c := whc.Context()
		var db dal.Database
		if db, err = facade.GetDatabase(c); err != nil {
			return
		}
		if err = db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) (err error) {
			var bill models.Bill
			if bill, err = getBill(c, tx, callbackUrl); err != nil {
				return
			}
			if bill.Data.GetUserGroupID() == "" {
				if whc.IsInGroup() {
					var group models.Group
					if group.ID, err = shared_group.GetUserGroupID(whc); err != nil {
						return
					}
					if bill, group, err = facade.Bill.AssignBillToGroup(c, tx, bill, group.ID, whc.AppUserStrID()); err != nil {
						return
					}
				} else {
					log.Debugf(c, "Not in group")
				}
			}
			m, err = f(whc, callbackUrl, bill)
			return err
		}); err != nil {
			return
		}
		return
	}
}
