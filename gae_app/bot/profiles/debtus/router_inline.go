package debtus

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_inline"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_invite"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_transfer"
	"github.com/strongo/log"
	"github.com/strongo/bots-framework/core"
	"strings"
)

var InlineQueryCommand = bots.Command{
	Code:       "inline-query",
	//InputTypes: []bots.WebhookInputType{bots.WebhookInputInlineQuery},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		c := whc.Context()

		inlineQuery := whc.Input().(bots.WebhookInlineQuery)
		query := inlineQuery.GetQuery()
		log.Debugf(c, "InlineQueryCommand.Action(query=%v)", query)
		switch {
		case query == "":
			m, err = dtb_inline.InlineEmptyQuery(whc)
		case query == "/invite":
			m, err = dtb_invite.InlineSendInvite(whc)
		case strings.HasPrefix(query, "receipt?id="):
			m, err = dtb_transfer.InlineSendReceipt(whc)
		//case strings.HasPrefix(query, "accept?transfer="):
		//	m, err = dtb_transfer.InlineAcceptTransfer(whc)
		default:
			amountMatches := dtb_inline.ReInlineQueryAmount.FindStringSubmatch(query)
			if amountMatches != nil {
				return dtb_inline.InlineNewRecord(whc, amountMatches)
			}
			log.Debugf(c, "Inline query not matched to any action: [%v]", query)
		}
		return
	},
}
