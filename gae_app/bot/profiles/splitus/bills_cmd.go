package splitus

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/bot_shared"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"net/url"
)

const BILLS_COMMAND = "bills"

var billsCommand = bots.Command{
	Code:     BILLS_COMMAND,
	Commands: []string{"/bills"},
	Action:   billsAction,
	CallbackAction: func(whc bots.WebhookContext, callbackURL *url.URL) (m bots.MessageFromBot, err error) {
		return billsAction(whc)
	},
}

func billsAction(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
	if !whc.IsInGroup() {
		m.Text = "This command is available just in group chats for now."
		return
	}

	var group models.Group
	if group, err = bot_shared.GetGroup(whc); err != nil {
		return
	}

	m.Format = bots.MessageFormatHTML

	if group.OutstandingBillsCount == 0 {
		mt := "This group has no outstanding bills"
		switch whc.InputType() {
		case bots.WebhookInputCallbackQuery:
			m.BotMessage = telegram_bot.CallbackAnswer(tgbotapi.AnswerCallbackQueryConfig{Text: mt})
		case bots.WebhookInputText:
			m.Text = mt
		default:
			err = errors.New("Unknown input type")
		}
		return
	}

	buf := new(bytes.Buffer)
	buf.WriteString("<b>Outstanding bills</b>\n\n")

	var outstandingBills []models.BillJson
	if outstandingBills, err = group.GetOutstandingBills(); err != nil {
		return
	}
	for i, bill := range outstandingBills {
		fmt.Fprintf(buf, `  %d. <a href="https://t.me/%v?start=bill-%v">%v</a>`+"\n", i+1, whc.GetBotCode(), bill.ID, bill.Name)
	}

	fmt.Fprintf(buf, "\nSend /split@%v to close the bills.\nThe debts records will be available in @DebtsTrackerBot.", whc.GetBotCode())

	m.Text = buf.String()
	return
}
