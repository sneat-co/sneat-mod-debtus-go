package dtb_transfer

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/db"
	"github.com/strongo/decimal"
	"github.com/strongo/log"
)

var ParseTransferCommand = bots.Command{
	Code: "parse-transfer",
	Matcher: func(c bots.Command, whc bots.WebhookContext) bool {
		input := whc.Input()
		switch input.(type) {
		case bots.WebhookTextMessage:
			return transferRegex.MatchString(input.(bots.WebhookTextMessage).Text())
		default:
			return false
		}
	},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		match := transferRegex.FindStringSubmatch(whc.Input().(bots.WebhookTextMessage).Text())
		var verb, valueS, counterpartyName, when string
		var direction models.TransferDirection
		var currency models.Currency

		for i, name := range transferRegex.SubexpNames() {
			if i != 0 && len(name) > 0 {
				v := strings.TrimSpace(match[i])
				if len(v) > 0 {
					switch name {
					case "verb":
						verb = v
					case "value":
						valueS = v
					case "currency":
						if string(v) == "" {
							currency = models.CURRENCY_USD //TODO: Replace with user's default currency
						} else {
							currency = models.Currency(strings.ToUpper(v))
						}
					case "direction":
						direction = models.TransferDirection(v)
					case "contact":
						counterpartyName = v
					case "when":
						when = v
					}
				}
			}
		}
		if verb == "" {
			switch direction {
			case models.TransferDirectionUser2Counterparty:
				verb = "got"
			case models.TransferDirectionCounterparty2User:
				verb = "gave"
			}
		} else {
			verb = strings.ToLower(verb)
			switch verb {
			case "send":
				verb = "sent"
			case "return":
				verb = "returned"
			}
		}

		m = whc.NewMessage("")

		value, _ := decimal.ParseDecimal64p2(valueS)

		isReturn := false

		creatorInfo := models.TransferCounterpartyInfo{
			UserID:      whc.AppUserIntID(),
			ContactName: counterpartyName,
		}
		c := whc.Context()

		from, to := facade.TransferCounterparties(direction, creatorInfo)

		var botUserEntity bots.BotAppUser
		botUserEntity, err = whc.GetAppUser()
		creatorUser := models.AppUser{
			IntegerID:     db.NewIntID(whc.AppUserIntID()),
			AppUserEntity: botUserEntity.(*models.AppUserEntity),
		}

		newTransfer := facade.NewTransferInput(whc.Environment(),
			GetTransferSource(whc),
			creatorUser,
			"",
			isReturn,
			0,
			from, to,
			models.Amount{Currency: currency, Value: value},
			time.Time{},
			models.NoInterest(),
		)

		output, err := facade.Transfers.CreateTransfer(c, newTransfer)

		//transferKey, err = nds.Put(ctx, transferKey, transfer)

		if err != nil {
			log.Errorf(c, "Failed to save transfer & counterparty to datastore: %v", err)
			return m, err
		}

		whc.ChatEntity().SetAwaitingReplyTo(fmt.Sprintf("ask-for-deadline:transferID=%v", output.Transfer.ID))

		m.Keyboard = tgbotapi.NewReplyKeyboardUsingStrings([][]string{
			{whc.Translate(trans.COMMAND_TEXT_YES_IT_HAS_RETURN_DEADLINE) + " " + emoji.ALARM_CLOCK_ICON},
			{whc.Translate(trans.COMMAND_TEXT_NO_IT_CAN_BE_RETURNED_ANYTIME)},
		})

		var buffer bytes.Buffer
		buffer.WriteString(fmt.Sprintf("You've %v %v %v %v %v", verb, valueS, currency, direction, counterpartyName))
		if when != "" {
			//TODO: Convert to time.Time
			buffer.WriteString(" " + when)
		}
		var counterparty models.Contact
		switch direction {
		case models.TransferDirectionUser2Counterparty:
			counterparty = output.To.Contact
		case models.TransferDirectionCounterparty2User:
			counterparty = output.From.Contact
		}
		counterpartyBalance := counterparty.Balance()
		buffer.WriteString(fmt.Sprintf(".\nTotal balance: %v", counterpartyBalance))
		//switch {
		//case counterparty.BalanceJson > 0: buffer.WriteString(fmt.Sprintf(".\nTotal balance: %v ows to you %v %v", contact, counterparty.BalanceJson, currency))
		//case counterparty.BalanceJson < 0: buffer.WriteString(fmt.Sprintf(".\nTotal balance: You owe to %v %v %v", contact, counterparty.BalanceJson, currency))
		//default:
		//}

		switch direction {
		case models.TransferDirectionCounterparty2User:
			buffer.WriteString("\n\nDo you need to return it on a specific date?")
		case models.TransferDirectionUser2Counterparty:
			buffer.WriteString(fmt.Sprintf("\n\nDoes %v have to return it on a specific date?", counterpartyName))
		}
		m.Text = buffer.String()

		return m, nil
	},
}
