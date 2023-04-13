package dtb_transfer

import (
	"fmt"
	"github.com/bots-go-framework/bots-api-telegram/tgbotapi"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/crediterra/money"
	"github.com/sneat-co/debtstracker-translations/trans"
	"net/url"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"errors"
	"github.com/sneat-co/debtstracker-translations/emoji"
	"github.com/strongo/decimal"
	"github.com/strongo/log"
	"golang.org/x/net/html"
)

//var StartReturnWizardCommand = botsfw.Command{
//	Code: "start-return-wizard",
//	Action: func(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error) {
//	},
//}

const RETURN_WIZARD_COMMAND = "return-wizard"

var StartReturnWizardCommand = botsfw.Command{
	Code:     RETURN_WIZARD_COMMAND,
	Commands: trans.Commands(trans.COMMAND_RETURNED),
	Replies:  []botsfw.Command{AskReturnCounterpartyCommand, AskToChooseDebtToReturnCommand},
	Action: func(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error) {
		log.Debugf(whc.Context(), "StartReturnWizardCommand.Action()")
		whc.ChatEntity().SetAwaitingReplyTo(RETURN_WIZARD_COMMAND)
		return AskReturnCounterpartyCommand.Action(whc)
	},
}

func askIfReturnedInFull(whc botsfw.WebhookContext, counterparty models.Contact, currency money.Currency, value decimal.Decimal64p2) (m botsfw.MessageFromBot, err error) {
	amount := money.Amount{Currency: money.Currency(currency), Value: value}
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
		whc.Translate(mt), html.EscapeString(counterparty.Data.FullName()), amount) +
		"\n\n" + whc.Translate(trans.MESSAGE_TEXT_IS_IT_RETURNED_IN_FULL))
	m.Format = botsfw.MessageFormatHTML
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
	[]botsfw.Command{
		AskToChooseDebtToReturnCommand,
		AskIfReturnedInFullCommand,
	},
	botsfw.Command{}, //newContactCommand - We do not allow to create a new contact on return
	func(whc botsfw.WebhookContext, counterparty models.Contact) (m botsfw.MessageFromBot, err error) {
		c := whc.Context()

		log.Debugf(c, "StartReturnWizardCommand.onCounterpartySelectedAction(counterparty.ID=%v)", counterparty.ID)
		var balanceWithInterest money.Balance
		balanceWithInterest, err = counterparty.Data.BalanceWithInterest(c, time.Now())
		if err != nil {
			err = fmt.Errorf("failed to get counterparty balance with interest: %w", err)
			return
		}
		//TODO: Display MESSAGE_TEXT_COUNTERPARTY_OWES_YOU_SINGLE_DEBT or MESSAGE_TEXT_YOU_OWE_TO_COUNTERPARTY_SINGLE_DEBT
		switch len(balanceWithInterest) {
		case 1:
			for currency, value := range balanceWithInterest {
				return askIfReturnedInFull(whc, counterparty, currency, value)
			}
		case 0:
			errorMessage := whc.Translate(trans.MESSAGE_TEXT_COUNTERPARTY_HAS_EMPTY_BALANCE, counterparty.Data.FullName())
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

func askToChooseDebt(whc botsfw.WebhookContext, buttons [][]string) (m botsfw.MessageFromBot) {
	if len(buttons) > 0 {
		whc.ChatEntity().PushStepToAwaitingReplyTo(ASK_TO_CHOOSE_DEBT_TO_RETURN_COMMAND)
		m = whc.NewMessage(whc.Translate(trans.MESSAGE_TEXT_CHOOSE_DEBT_THAT_HAS_BEEN_RETURNED))
		m.Keyboard = tgbotapi.NewReplyKeyboardUsingStrings(buttons)
	} else {
		m = whc.NewMessage(whc.Translate(trans.MESSAGE_TEXT_NO_DEBTS_TO_RETURN))
	}
	return m
}

func _debtAmountButtonText(whc botsfw.WebhookContext, currency money.Currency, value decimal.Decimal64p2, counterparty models.Contact) string {
	amount := money.Amount{Currency: currency, Value: value.Abs()}
	var mt string
	switch {
	case value > 0:
		mt = trans.BUTTON_TEXT_SOMEONE_OWES_TO_YOU_AMOUNT
	case value < 0:
		mt = trans.BUTTON_TEXT_YOU_OWE_AMOUNT_TO_SOMEONE
	default:
		mt = "ERROR (%v) - zero value: %v"
	}
	return fmt.Sprintf(whc.Translate(mt), counterparty.Data.FullName(), amount)
}

const ASK_IF_RETURNED_IN_FULL_COMMAND = "ask-if-return-in-full"

var AskIfReturnedInFullCommand = botsfw.Command{
	Code:    ASK_IF_RETURNED_IN_FULL_COMMAND,
	Replies: []botsfw.Command{AskHowMuchHaveBeenReturnedCommand},
	Action: func(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error) {
		chatEntity := whc.ChatEntity()
		if chatEntity.IsAwaitingReplyTo(ASK_IF_RETURNED_IN_FULL_COMMAND) {
			switch whc.Input().(botsfw.WebhookTextMessage).Text() {
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

func processReturnCommand(whc botsfw.WebhookContext, returnValue decimal.Decimal64p2) (m botsfw.MessageFromBot, err error) {
	if returnValue < 0 {
		panic(fmt.Sprintf("returnValue < 0: %v", returnValue))
	}
	c := whc.Context()
	chatEntity := whc.ChatEntity()
	var (
		counterpartyID int64
		transferID     int
	)
	if counterpartyID, transferID, err = getReturnWizardParams(whc); err != nil {
		return m, err
	}
	counterparty, err := getCounterparty(whc, counterpartyID)
	if err != nil {
		return m, err
	}
	counterpartyBalanceWithInterest, err := counterparty.Data.BalanceWithInterest(c, time.Now())
	if err != nil {
		err = fmt.Errorf("failed to get balance with interest for contact %v: %v", counterparty.ID, err)
		return
	}
	awaitingUrl, err := url.Parse(chatEntity.GetAwaitingReplyTo())
	if err != nil {
		return m, err
	}
	currency := money.Currency(awaitingUrl.Query().Get("currency"))

	if transferID != 0 && returnValue > 0 {
		var transfer models.Transfer
		if transfer, err = facade.Transfers.GetTransferByID(whc.Context(), nil, transferID); err != nil {
			return
		}

		returnAmount := money.NewAmount(currency, returnValue)
		if outstandingAmount := transfer.Data.GetOutstandingAmount(time.Now()); outstandingAmount.Value < returnValue {
			m.Text = whc.Translate(trans.MESSAGE_TEXT_RETURN_IS_TOO_BIG, returnAmount, outstandingAmount, outstandingAmount.Value)
			return
		}
	}

	if previousBalance, ok := counterpartyBalanceWithInterest[currency]; ok {
		if returnValue == 0 {
			returnValue = previousBalance.Abs()
		}
		previousBalance := money.Amount{Currency: currency, Value: previousBalance}
		direction, err := getReturnDirectionFromDebtValue(previousBalance)
		if err != nil {
			return m, err
		}
		return CreateReturnAndShowReceipt(whc, transferID, counterpartyID, direction, money.NewAmount(currency, returnValue))
	} else {
		return m, fmt.Errorf("Contact has no currency in balance. counterpartyID=%v,  currency='%v'", counterpartyID, currency)
	}
}

const ASK_HOW_MUCH_HAVE_BEEN_RETURNED = "ask-how-much-have-been-returned"

var AskHowMuchHaveBeenReturnedCommand = botsfw.Command{
	Code: ASK_HOW_MUCH_HAVE_BEEN_RETURNED,
	Action: func(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error) {
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

func TryToProcessHowMuchHasBeenReturned(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error) {
	if amountValue, err := decimal.ParseDecimal64p2(whc.Input().(botsfw.WebhookTextMessage).Text()); err != nil {
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

var AskToChooseDebtToReturnCommand = botsfw.Command{
	Code: ASK_TO_CHOOSE_DEBT_TO_RETURN_COMMAND,
	Replies: []botsfw.Command{
		AskIfReturnedInFullCommand,
	},
	Action: func(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error) {
		c := whc.Context()
		counterpartyID, _, _ := getReturnWizardParams(whc)
		var (
			theCounterparty models.Contact
			balance         money.Balance
		)
		if counterpartyID == 0 {
			// Let's try to get counterpartyEntity from message text
			mt := whc.Input().(botsfw.WebhookTextMessage).Text()
			splittedBySeparator := strings.Split(mt, "|")
			counterpartyTitle := strings.Join(splittedBySeparator[:len(splittedBySeparator)-1], "|")
			counterpartyTitle = strings.TrimSpace(counterpartyTitle)
			chatEntity := whc.ChatEntity()
			var botAppUser botsfw.BotAppUser
			botAppUser, err = whc.GetAppUser()
			if err != nil {
				return m, err
			}
			user := botAppUser.(*models.AppUserData)
			var counterparties []models.Contact
			if counterparties, err = dtdal.Contact.GetLatestContacts(whc, nil, 0, user.TotalContactsCount()); err != nil {
				return m, err
			}
			var counterpartyFound bool
			now := time.Now()
			for _, counterpartyItem := range counterparties {
				counterpartyItemTitle := counterpartyItem.Data.FullName()
				if counterpartyItemTitle == counterpartyTitle {
					if balance, err = counterpartyItem.Data.BalanceWithInterest(c, now); err != nil {
						err = fmt.Errorf("failed to get balance with interest for contact %v: %w", counterpartyItem.ID, err)
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
			if balance, err = counterparty.Data.BalanceWithInterest(c, time.Now()); err != nil {
				err = fmt.Errorf("failed to get balance with interest for contact %v: %w", counterparty.ID, err)
				return
			}
			theCounterparty = counterparty
		}

		mt := whc.Input().(botsfw.WebhookTextMessage).Text()
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

func CreateReturnAndShowReceipt(whc botsfw.WebhookContext, returnToTransferID int, counterpartyID int64, direction models.TransferDirection, returnAmount money.Amount) (m botsfw.MessageFromBot, err error) {
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

func getReturnDirectionFromDebtValue(currentDebt money.Amount) (models.TransferDirection, error) {
	switch {
	case currentDebt.Value < 0:
		return models.TransferDirectionUser2Counterparty, nil
	case currentDebt.Value > 0:
		return models.TransferDirectionCounterparty2User, nil
	}
	return models.TransferDirection(""), fmt.Errorf("Zero value for currency: [%v]", currentDebt.Currency)
}

func getReturnWizardParams(whc botsfw.WebhookContext) (counterpartyID int64, transferID int, err error) {
	awaitingReplyTo := whc.ChatEntity().GetAwaitingReplyTo()
	params, err := url.ParseQuery(botsfw.AwaitingReplyToQuery(awaitingReplyTo))
	if err != nil {
		return counterpartyID, transferID, fmt.Errorf("failed in AwaitingReplyToQuery(): %w", err)
	}
	if counterpartyID, err = strconv.ParseInt(params.Get(WIZARD_PARAM_COUNTERPARTY), 10, 64); err != nil {
		return counterpartyID, transferID, fmt.Errorf("failed to get counterparty ID: %w", err)
	}
	transferID, _ = strconv.Atoi(params.Get(WIZARD_PARAM_TRANSFER))
	return
}

func getCounterparty(whc botsfw.WebhookContext, counterpartyID int64) (counterparty models.Contact, err error) {
	//counterparty = new(models.Contact)
	if counterparty, err = facade.GetContactByID(whc.Context(), nil, counterpartyID); err != nil {
		return
	}
	return
}
