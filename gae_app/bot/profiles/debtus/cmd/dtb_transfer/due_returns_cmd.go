package dtb_transfer

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bytes"
	"fmt"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/app"
	"github.com/strongo/log"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"html"
	"net/url"
	"strings"
	"time"
)

const DUE_RETURNS_COMMAND = "due-returns"

var DueReturnsCallbackCommand = bots.NewCallbackCommand(DUE_RETURNS_COMMAND, dueReturnsCallbackAction)

func dueReturnsCallbackAction(whc bots.WebhookContext, _ *url.URL) (m bots.MessageFromBot, err error) {
	c := whc.Context()

	userID := whc.AppUserIntID()
	var (
		overdueTransfers, dueTransfers []models.Transfer
	)

	er := make(chan error, 2)
	go func(er chan<- error) {
		if overdueTransfers, err = dal.Transfer.LoadOverdueTransfers(c, userID, 5); err != nil {
			er <- errors.Wrap(err, "Failed to get overdue transfers")
		} else {
			log.Debugf(c, "Loaded %v overdue transfer", len(overdueTransfers))
			er <- nil
		}
	}(er)
	go func(er chan<- error) {
		if dueTransfers, err = dal.Transfer.LoadDueTransfers(c, userID, 5); err != nil {
			er <- errors.Wrap(err, "Failed to get due transfers")
		} else {
			log.Debugf(c, "Loaded %v due transfer", len(dueTransfers))
			er <- nil
		}
	}(er)

	for i := 0; i < 2; i++ {
		if err = <-er; err != nil {
			return
		}
	}

	if len(overdueTransfers) == 0 || len(dueTransfers) == 0 {
		if m, err = whc.NewEditMessage(whc.Translate(trans.MESSAGE_TEXT_DUE_RETURNS_EMPTY), bots.MessageFormatHTML); err != nil {
			return
		}
	} else {
		var buffer bytes.Buffer

		now := time.Now()
		listTransfers := func(header string, transfers []models.Transfer) {
			if len(transfers) == 0 {
				return
			}
			buffer.WriteString(whc.Translate(header))
			buffer.WriteString("\n\n")
			for i, transfer := range transfers {
				switch transfer.Direction() {
				case models.TransferDirectionCounterparty2User:
					buffer.WriteString(whc.Translate(trans.MESSAGE_TEXT_DUE_RETURNS_ROW_BY_USER, html.EscapeString(transfer.Counterparty().ContactName), transfer.GetAmount(), DurationToString(transfer.DtDueOn.Sub(now), whc)))
				case models.TransferDirectionUser2Counterparty:
					buffer.WriteString(whc.Translate(trans.MESSAGE_TEXT_DUE_RETURNS_ROW_BY_COUNTERPARTY, html.EscapeString(transfer.Counterparty().ContactName), transfer.GetAmount(), DurationToString(transfer.DtDueOn.Sub(now), whc)))
				default:
					panic(fmt.Sprintf("Unknown direction for transfer id=%v: %v", transfers[i].ID, transfer))
				}
				buffer.WriteString("\n")
			}
			buffer.WriteString("\n")
		}
		listTransfers(trans.MESSAGE_TEXT_OVERDUE_RETURNS_HEADER, overdueTransfers)
		listTransfers(trans.MESSAGE_TEXT_DUE_RETURNS_HEADER, dueTransfers)
		if m, err = whc.NewEditMessage(strings.TrimSuffix(buffer.String(), "\n"), bots.MessageFormatHTML); err != nil {
			return
		}
	}
	m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			{
				Text:         whc.CommandText(trans.COMMAND_TEXT_BALANCE, emoji.BALANCE_ICON),
				CallbackData: BALANCE_COMMAND,
			},
		},
	)

	return m, err
}

func DurationToString(d time.Duration, translator strongo.SingleLocaleTranslator) string {
	hours := d.Hours()
	switch hours {
	case 0:
		switch d.Minutes() {
		case 0:
			return translator.Translate(trans.DUE_IN_NOW)
		case 1:
			return translator.Translate(trans.DUE_IN_A_MINUTE)
		default:
			return fmt.Sprintf(translator.Translate(trans.DUE_IN_X_MINUTES), d.Minutes())
		}
	case 1:
		return translator.Translate(trans.DUE_IN_AN_HOUR)
	default:
		if hours < 24 {
			return fmt.Sprintf(translator.Translate(trans.DUE_IN_X_HOURS), int(hours))
		}
		return fmt.Sprintf(translator.Translate(trans.DUE_IN_X_DAYS), int(hours/24))
	}
}
