package dtb_transfer

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/decimal"
	"github.com/strongo/log"
	"golang.org/x/net/html"
)

//var StartReturnWizardCommand = bots.Command{
//	Code: "start-return-wizard",
//	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
//	},
//}

const RETURN_WIZARD_COMMAND = "return-wizard"

var StartReturnWizardCommand = bots.Command{
	Code:     RETURN_WIZARD_COMMAND,
	Commands: trans.Commands(trans.COMMAND_RETURNED),
	Replies:  []bots.Command{AskReturnCounterpartyCommand, AskToChooseDebtToReturnCommand},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		log.Debugf(whc.Context(), "StartReturnWizardCommand.Action()")
		whc.ChatEntity().SetAwaitingReplyTo(RETURN_WIZARD_COMMAND)
		return AskReturnCounterpartyCommand.Action(whc)
	},
}

func askIfReturnedInFull(whc bots.WebhookContext, counterparty models.Contact, currency models.Currency, value decimal.Decimal64p2) (m bots.MessageFromBot, err error) {
	amount := models.Amount{Currency: models.Currency(currency), Value: value}
	var mt string
	switch {
	case value < 0:
		mt = trans.MESSAGE_TEXT_YOU_OWE_TO_COUNTERPARTY_SINGLE_DEBT
	case value > 0:
		mt = trans.MESSAGE_TEXT_COUNTERPARTY_OWES_YOU_SINGLE_DEBT
	case value == 0:
		errorMessage := fmt.Sprintf("ERROR: Balance for currency [%v] is: %v", currency, value)
		log.Warningf(whc.Context(), errorMessage)
		m = whc.NewMessage(errorMessage)
		return
	}
	chatEntity := whc.ChatEntity()
	chatEntity.PushStepToAwaitingReplyTo(ASK_IF_RETURNED_IN_FULL_COMMAND)
	chatEntity.AddWizardParam("currency", string(currency))
	amount.Value = amount.Value.Abs()
	m = whc.NewMessage(fmt.Sprintf(
		whc.Translate(mt), html.EscapeString(counterparty.FullName()), amount) +
		"\n\n" + whc.Translate(trans.MESSAGE_TEXT_IS_IT_RETURNED_IN_FULL))
	m.Format = bots.MessageFormatHTML
	m.Keyboard = tgbotapi.NewReplyKeyboardUsingStrings(
		[][]string{
			{whc.Translate(trans.BUTTON_TEXT_DEBT_RETURNED_FULLY)},
			{whc.Translate(trans.BUTTON_TEXT_DEBT_RETURNED_PARTIALLY)},
			{whc.Translate(trans.COMMAND_TEXT_CANCEL)},
		},
	)
	return
}

const ASK_RETURN_COUNTERPARTY_COMMAND = "ask-return-counterparty"

var AskReturnCounterpartyCommand = CreateAskTransferCounterpartyCommand(
	true,
	ASK_RETURN_COUNTERPARTY_COMMAND,
	trans.COMMAND_TEXT_RETURN,
	emoji.RETURN_BACK_ICON,
	trans.MESSAGE_TEXT_RETURN_ASK_TO_CHOOSE_COUNTERPARTY,
	[]bots.Command{
		AskToChooseDebtToReturnCommand,
		AskIfReturnedInFullCommand,
	},
	bots.Command{}, //newContactCommand - We do not allow to create a new contact on return
	func(whc bots.WebhookContext, counterparty models.Contact) (m bots.MessageFromBot, err error) {
		c := whc.Context()

		log.Debugf(c, "StartReturnWizardCommand.onCounterpartySelectedAction(counterparty.ID=%v)", counterparty.ID)
		var balanceWithInterest models.Balance
		balanceWithInterest, err = counterparty.BalanceWithInterest(c, time.Now())
		if err != nil {
			err = errors.WithMessage(err, fmt.Sprintf("Failed to get counterparty balance with interest: %v", err))
			return
		}
		//TODO: Display MESSAGE_TEXT_COUNTERPARTY_OWES_YOU_SINGLE_DEBT or MESSAGE_TEXT_YOU_OWE_TO_COUNTERPARTY_SINGLE_DEBT
		switch len(balanceWithInterest) {
		case 1:
			for currency, value := range balanceWithInterest {
				return askIfReturnedInFull(whc, counterparty, currency, value)
			}
		case 0:
			errorMessage := whc.Translate(trans.MESSAGE_TEXT_COUNTERPARTY_HAS_EMPTY_BALANCE, counterparty.FullName())
			log.Debugf(c, "Balance is empty: "+errorMessage)
			m = whc.NewMessage(errorMessage)
		default:
			buttons := make([][]string, len(balanceWithInterest)+1)
			var i int
			buttons[0] = []string{whc.Translate(trans.COMMAND_TEXT_CANCEL)}
			for currency, value := range balanceWithInterest {
				i += 1
				buttons[i] = []string{_debtAmountButtonText(whc, currency, value, counterparty)}
			}
			m = askToChooseDebt(whc, buttons)
		}
		return
	},
)

func askToChooseDebt(whc bots.WebhookContext, buttons [][]string) (m bots.MessageFromBot) {
	if len(buttons) > 0 {
		whc.ChatEntity().PushStepToAwaitingReplyTo(ASK_TO_CHOOSE_DEBT_TO_RETURN_COMMAND)
		m = whc.NewMessage(whc.Translate(trans.MESSAGE_TEXT_CHOOSE_DEBT_THAT_HAS_BEEN_RETURNED))
		m.Keyboard = tgbotapi.NewReplyKeyboardUsingStrings(buttons)
	} else {
		m = whc.NewMessage(whc.Translate(trans.MESSAGE_TEXT_NO_DEBTS_TO_RETURN))
	}
	return m
}

func _debtAmountButtonText(whc bots.WebhookContext, currency models.Currency, value decimal.Decimal64p2, counterparty models.Contact) string {
	amount := models.Amount{Currency: currency, Value: value.Abs()}
	var mt string
	switch {
	case value > 0:
		mt = trans.BUTTON_TEXT_SOMEONE_OWES_TO_YOU_AMOUNT
	case value < 0:
		mt = trans.BUTTON_TEXT_YOU_OWE_AMOUNT_TO_SOMEONE
	default:
		mt = "ERROR (%v) - zero value: %v"
	}
	return fmt.Sprintf(whc.Translate(mt), counterparty.FullName(), amount)
}

const ASK_IF_RETURNED_IN_FULL_COMMAND = "ask-if-return-in-full"

var AskIfReturnedInFullCommand = bots.Command{
	Code:    ASK_IF_RETURNED_IN_FULL_COMMAND,
	Replies: []bots.Command{AskHowMuchHaveBeenReturnedCommand},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		chatEntity := whc.ChatEntity()
		if chatEntity.IsAwaitingReplyTo(ASK_IF_RETURNED_IN_FULL_COMMAND) {
			switch whc.Input().(bots.WebhookTextMessage).Text() {
			case whc.Translate(trans.BUTTON_TEXT_DEBT_RETURNED_FULLY):
				m, err = processReturnCommand(whc, 0)
				//common.CreateTransfer(whc.Context(), whc.AppUserIntID(), )
			case whc.Translate(trans.BUTTON_TEXT_DEBT_RETURNED_PARTIALLY):
				m, err = AskHowMuchHaveBeenReturnedCommand.Action(whc)
			default:
				return TryToProcessHowMuchHasBeenReturned(whc)
			}
			return m, err

		} else {
			err = errors.New("AskIfReturnedInFullCommand: Not implemented yet")
			return m, err
		}
	},
}

func processReturnCommand(whc bots.WebhookContext, returnValue decimal.Decimal64p2) (m bots.MessageFromBot, err error) {
	if returnValue < 0 {
		panic(fmt.Sprintf("returnValue < 0: %v", returnValue))
	}
	c := whc.Context()
	chatEntity := whc.ChatEntity()
	var (
		counterpartyID int64
		transferID     int64
	)
	if counterpartyID, transferID, err = getReturnWizardParams(whc); err != nil {
		return m, err
	}
	counterparty, err := getCounterparty(whc, counterpartyID)
	if err != nil {
		return m, err
	}
	counterpartyBalanceWithInterest, err := counterparty.BalanceWithInterest(c, time.Now())
	if err != nil {
		err = errors.WithMessage(err, fmt.Sprintf("Failed to get balance with interest for contact %v: %v", counterparty.ID, err))
		return
	}
	awaitingUrl, err := url.Parse(chatEntity.GetAwaitingReplyTo())
	if err != nil {
		return m, err
	}
	currency := models.Currency(awaitingUrl.Query().Get("currency"))

	if transferID != 0 && returnValue > 0 {
		var transfer models.Transfer
		if transfer, err = facade.GetTransferByID(whc.Context(), transferID); err != nil {
			return
		}

		returnAmount := models.NewAmount(currency, returnValue)
		if outstandingAmount := transfer.GetOutstandingAmount(time.Now()); outstandingAmount.Value < returnValue {
			m.Text = whc.Translate(trans.MESSAGE_TEXT_RETURN_IS_TOO_BIG, returnAmount, outstandingAmount, outstandingAmount.Value)
			return
		}
	}

	if previousBalance, ok := counterpartyBalanceWithInterest[currency]; ok {
		if returnValue == 0 {
			returnValue = previousBalance.Abs()
		}
		previousBalance := models.Amount{Currency: currency, Value: previousBalance}
		direction, err := getReturnDirectionFromDebtValue(previousBalance)
		if err != nil {
			return m, err
		}
		return CreateReturnAndShowReceipt(whc, transferID, counterpartyID, direction, models.NewAmount(currency, returnValue))
	} else {
		return m, fmt.Errorf("Contact has no currency in balance. counterpartyID=%v,  currency='%v'", counterpartyID, currency)
	}
}

const ASK_HOW_MUCH_HAVE_BEEN_RETURNED = "ask-how-much-have-been-returned"

var AskHowMuchHaveBeenReturnedCommand = bots.Command{
	Code: ASK_HOW_MUCH_HAVE_BEEN_RETURNED,
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "AskHowMuchHaveBeenReturnedCommand.Action()")
		chatEntity := whc.ChatEntity()
		if chatEntity.IsAwaitingReplyTo(ASK_HOW_MUCH_HAVE_BEEN_RETURNED) {
			return TryToProcessHowMuchHasBeenReturned(whc)
		} else {
			m = whc.NewMessage(whc.Translate(trans.MESSAGE_TEXT_ASK_HOW_MUCH_HAS_BEEN_RETURNED))
			m.Keyboard = tgbotapi.NewHideKeyboard(true)
			chatEntity.PushStepToAwaitingReplyTo(ASK_HOW_MUCH_HAVE_BEEN_RETURNED)
			return m, err
		}
	},
}

func TryToProcessHowMuchHasBeenReturned(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
	if amountValue, err := decimal.ParseDecimal64p2(whc.Input().(bots.WebhookTextMessage).Text()); err != nil {
		m = whc.NewMessage(whc.Translate(trans.MESSAGE_TEXT_INCORRECT_VALUE_NOT_A_NUMBER))
		return m, nil
	} else {
		if amountValue > 0 {
			return processReturnCommand(whc, amountValue)
		} else {
			m = whc.NewMessage(whc.Translate(trans.MESSAGE_TEXT_INCORRECT_VALUE_IS_NEGATIVE))
			return m, nil
		}
	}
}

const ASK_TO_CHOOSE_DEBT_TO_RETURN_COMMAND = "ask-to-choose-debt-to-return"

var AskToChooseDebtToReturnCommand = bots.Command{
	Code: ASK_TO_CHOOSE_DEBT_TO_RETURN_COMMAND,
	Replies: []bots.Command{
		AskIfReturnedInFullCommand,
	},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		counterpartyID, _, _ := getReturnWizardParams(whc)
		var (
			theCounterparty models.Contact
			balance         models.Balance
		)
		if counterpartyID == 0 {
			// Let's try to get counterpartyEntity from message text
			mt := whc.Input().(bots.WebhookTextMessage).Text()
			splittedBySeparator := strings.Split(mt, "|")
			counterpartyTitle := strings.Join(splittedBySeparator[:len(splittedBySeparator)-1], "|")
			counterpartyTitle = strings.TrimSpace(counterpartyTitle)
			chatEntity := whc.ChatEntity()
			var botAppUser bots.BotAppUser
			botAppUser, err = whc.GetAppUser()
			if err != nil {
				return m, err
			}
			user := botAppUser.(*models.AppUserEntity)
			var counterparties []models.Contact
			if counterparties, err = dal.Contact.GetLatestContacts(whc, 0, user.TotalContactsCount()); err != nil {
				return m, err
			}
			var counterpartyFound bool
			now := time.Now()
			for _, counterpartyItem := range counterparties {
				counterpartyItemTitle := counterpartyItem.FullName()
				if counterpartyItemTitle == counterpartyTitle {
					if balance, err = counterpartyItem.BalanceWithInterest(c, now); err != nil {
						err = errors.WithMessage(err, fmt.Sprintf("Failed to get balance with interest for contact %v", counterpartyItem.ID))
						return
					}
					theCounterparty = counterpartyItem
					counterpartyFound = true
					chatEntity.AddWizardParam(WIZARD_PARAM_COUNTERPARTY, strconv.FormatInt(counterpartyItem.ID, 10))
					break
				}
			}
			if !counterpartyFound {
				m = whc.NewMessageByCode(trans.MESSAGE_TEXT_UNKNOWN_COUNTERPARTY_ON_RETURN)
				return m, nil
			}
		} else {
			var counterparty models.Contact
			counterparty, err = getCounterparty(whc, counterpartyID)
			if balance, err = counterparty.BalanceWithInterest(c, time.Now()); err != nil {
				err = errors.WithMessage(err, fmt.Sprintf("Failed to get balance with interest for contact %v", counterparty.ID))
				return
			}
			theCounterparty = counterparty
		}

		mt := whc.Input().(bots.WebhookTextMessage).Text()
		for currency, value := range balance {
			if mt == _debtAmountButtonText(whc, currency, value, theCounterparty) {
				return askIfReturnedInFull(whc, theCounterparty, currency, value)
			}
		}
		if m.Text == "" {
			m = whc.NewMessageByCode(trans.MESSAGE_TEXT_UNKNOWN_DEBT)
		}
		return m, err
	},
}

func CreateReturnAndShowReceipt(whc bots.WebhookContext, returnToTransferID, counterpartyID int64, direction models.TransferDirection, returnAmount models.Amount) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	log.Debugf(c, "CreateReturnAndShowReceipt(returnToTransferID=%d, counterpartyID=%d)", returnToTransferID, counterpartyID)

	if returnAmount.Value < 0 {
		log.Warningf(c, "returnAmount.Value < 0: %v", returnAmount.Value)
		returnAmount.Value = returnAmount.Value.Abs()
	}

	creatorInfo := models.TransferCounterpartyInfo{
		UserID:    whc.AppUserIntID(),
		ContactID: counterpartyID,
	}

	if m, err = CreateTransferFromBot(whc, true, returnToTransferID, direction, creatorInfo, returnAmount, time.Time{}, models.NoInterest()); err != nil {
		return m, err
	}
	log.Debugf(c, "createReturnAndShowReceipt(): %v", m)
	return m, err
}

func getReturnDirectionFromDebtValue(currentDebt models.Amount) (models.TransferDirection, error) {
	switch {
	case currentDebt.Value < 0:
		return models.TransferDirectionUser2Counterparty, nil
	case currentDebt.Value > 0:
		return models.TransferDirectionCounterparty2User, nil
	}
	return models.TransferDirection(""), fmt.Errorf("Zero value for currency: [%v]", currentDebt.Currency)
}

func getReturnWizardParams(whc bots.WebhookContext) (counterpartyID, transferID int64, err error) {
	awaitingReplyTo := whc.ChatEntity().GetAwaitingReplyTo()
	params, err := url.ParseQuery(bots.AwaitingReplyToQuery(awaitingReplyTo))
	if err != nil {
		return counterpartyID, transferID, errors.Wrap(err, "Failed in AwaitingReplyToQuery()")
	}
	if counterpartyID, err = strconv.ParseInt(params.Get(WIZARD_PARAM_COUNTERPARTY), 10, 64); err != nil {
		return counterpartyID, transferID, errors.Wrap(err, "Failed to get counterparty ID")
	}
	transferID, _ = strconv.ParseInt(params.Get(WIZARD_PARAM_TRANSFER), 10, 64)
	return
}

func getCounterparty(whc bots.WebhookContext, counterpartyID int64) (counterparty models.Contact, err error) {
	//counterparty = new(models.Contact)
	if counterparty, err = facade.GetContactByID(whc.Context(), counterpartyID); err != nil {
		return
	}
	return
}
