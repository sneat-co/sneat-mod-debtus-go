package dtb_invite

import (
	"net/url"

	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_transfer"
	"github.com/strongo/log"
)

var ChosenInlineResultCommand = botsfw.Command{
	Code:       "inline-create-invite",
	InputTypes: []botsfw.WebhookInputType{bots.WebhookInputChosenInlineResult},
	Action: func(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error) {
		c := whc.Context()
		chosenResult := whc.Input().(bots.WebhookChosenInlineResult)
		query := chosenResult.GetQuery()
		log.Debugf(c, "ChosenInlineResultCommand.Action() => query: %v", query)

		queryUrl, err := url.Parse(query)
		if err != nil {
			return m, err
		}

		switch queryUrl.Path {
		case "receipt":
			return dtb_transfer.OnInlineChosenCreateReceipt(whc, chosenResult.GetInlineMessageID(), queryUrl)
		default:
			log.Warningf(c, "Unknown chosen inline query: "+query)
		}
		return
	},
}
