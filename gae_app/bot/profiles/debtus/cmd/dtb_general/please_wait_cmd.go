package dtb_general

import (
	"net/url"
)

const PLEASE_WAIT_COMMAND = "please-wait"

var PleaseWaitCommand = botsfw.Command{
	Code: PLEASE_WAIT_COMMAND,
	CallbackAction: func(whc botsfw.WebhookContext, _ *url.URL) (botsfw.MessageFromBot, error) {
		return whc.NewMessageByCode(trans.MESSAGE_TEXT_PLEASE_WAIT), nil
	},
}
