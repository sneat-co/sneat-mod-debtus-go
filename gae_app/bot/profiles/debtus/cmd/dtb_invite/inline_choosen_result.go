package dtb_invite

import (
	"net/url"

	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_transfer"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
)

var ChosenInlineResultCommand = bots.Command{
	Code:       "inline-create-invite",
	InputTypes: []bots.WebhookInputType{bots.WebhookInputChosenInlineResult},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		choosenResult := whc.Input().(bots.WebhookChosenInlineResult)
		query := choosenResult.GetQuery()
		log.Debugf(c, "ChosenInlineResultCommand.Action() => query: %v", query)

		queryUrl, err := url.Parse(query)
		if err != nil {
			return m, err
		}

		switch queryUrl.Path {
		case "receipt":
			return dtb_transfer.OnInlineChoosenCreateReceipt(whc, choosenResult.GetInlineMessageID(), queryUrl)
		default:
			log.Warningf(c, "Unknown choosen inline query: "+query)
		}
		return
	},
}
