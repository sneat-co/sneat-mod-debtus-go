package shared_all

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"context"
	"net/url"
)

func TransactionalCallbackCommand(c botsfw.Command, o db.RunOptions) botsfw.Command {
	c.CallbackAction = TransactionalCallbackAction(o, c.CallbackAction)
	return c
}

func TransactionalCallbackAction(o db.RunOptions,
	f func(whc botsfw.WebhookContext, callbackUrl *url.URL) (m botsfw.MessageFromBot, err error),
) func(whc botsfw.WebhookContext, callbackUrl *url.URL) (m botsfw.MessageFromBot, err error) {
	return func(whc botsfw.WebhookContext, callbackUrl *url.URL) (m botsfw.MessageFromBot, err error) {
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
