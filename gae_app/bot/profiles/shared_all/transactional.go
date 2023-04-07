package shared_all

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"context"
	"github.com/strongo/bots-framework/core"
	"net/url"
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
		err = dtdal.DB.RunInTransaction(c, func(tc context.Context) error {
			whc.SetContext(tc)
			m, err = f(whc, callbackUrl)
			whc.SetContext(c)
			return err
		}, o)
		return
	}
}
