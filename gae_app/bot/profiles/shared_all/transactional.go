package shared_all

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"context"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/dal-go/dalgo/dal"
	"net/url"
)

//func TransactionalCallbackCommand(c botsfw.Command, o db.RunOptions) botsfw.Command {
//	c.CallbackAction = TransactionalCallbackAction(o, c.CallbackAction)
//	return c
//}

func TransactionalCallbackAction(o dal.TransactionOptions,
	f func(whc botsfw.WebhookContext, callbackUrl *url.URL) (m botsfw.MessageFromBot, err error),
) func(whc botsfw.WebhookContext, callbackUrl *url.URL) (m botsfw.MessageFromBot, err error) {
	return func(whc botsfw.WebhookContext, callbackUrl *url.URL) (m botsfw.MessageFromBot, err error) {
		c := whc.Context()
		var db dal.Database
		if db, err = facade.GetDatabase(c); err != nil {
			return
		}
		err = db.RunReadwriteTransaction(c, func(tc context.Context, tx dal.ReadwriteTransaction) error {
			whc.SetContext(tc)
			m, err = f(whc, callbackUrl)
			whc.SetContext(c)
			return err
		})
		return
	}
}
