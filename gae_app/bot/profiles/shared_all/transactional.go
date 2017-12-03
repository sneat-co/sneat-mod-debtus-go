package shared_all

import (
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/db"
	"net/url"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"golang.org/x/net/context"
)

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

