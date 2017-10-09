package dtb_general

import (
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-framework/core"
	"net/url"
)

const PLEASE_WAIT_COMMAND = "please-wait"

var PleaseWaitCommand = bots.Command{
	Code: PLEASE_WAIT_COMMAND,
	CallbackAction: func(whc bots.WebhookContext, _ *url.URL) (bots.MessageFromBot, error) {
		return whc.NewMessageByCode(trans.MESSAGE_TEXT_PLEASE_WAIT), nil
	},
}
