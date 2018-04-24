package shared_all

import "github.com/strongo/bots-framework/core"

const howToCommandCode = "how-to"

var howToCommand = bots.Command{
	Code: howToCommandCode,
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		m.Text = "<b>How To</b> - not implemented yet"
		return
	},
}
