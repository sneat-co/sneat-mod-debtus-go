package dtb_transfer

import (
	"bytes"
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_general"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/analytics"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/app"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/bots-framework/platforms/viber"
	"github.com/strongo/decimal"
	"github.com/strongo/log"
	"golang.org/x/net/html"
)

var transferRegex = regexp.MustCompile(`(?i)((?P<verb>\w+) )?(?P<amount>\d+)\s*(?P<currency>\w{3})?\s*(?P<direction>from|to)\s+(?P<contact>.+)[\s\.]*`)

//\s*(?P<when>today|yesterday|(last|this)\s+\w+|on\s+(\d{1,2}[\./]\d{1,2}[\./]\d{4}))?

func IsCurrencyIcon(s string) bool {
	for _, icon := range CURRENCY_ICONS {
		if s == icon {
			return true
		}
	}
	return false
}

var CURRENCY_ICONS = []string{
	emoji.CD_ICON,
	emoji.BOOK_ICON,
	emoji.TEACUP_ICON,
	emoji.BEER_ICON,
	emoji.FORK_AND_KNIFE_ICON,
	emoji.HAMMER_ICON,
	emoji.TAXI_ICON,
	emoji.BICYCLE_ICON,
	emoji.HOURGLASS_ICON,
	emoji.APPAREL_ICON,
	emoji.SMOKING_ICON,
}

var currenciesByPriority = []models.Currency{
	models.CURRENCY_EUR,
	models.CURRENCY_USD,
	models.CURRENCY_GBP,
	models.CURRENCY_JPY,
	models.CURRENCY_IRR,
	models.CURRENCY_RUB,

	models.CURRENCY_UAH,
	models.CURRENCY_BYN,
	models.CURRENCY_TJS,
	models.CURRENCY_UZS,
	models.Currency(emoji.CD_ICON),
	models.Currency(emoji.BOOK_ICON),
	models.Currency(emoji.BEER_ICON),
	models.Currency(emoji.TEACUP_ICON),
	models.Currency(emoji.HOURGLASS_ICON),
	models.Currency(emoji.TAXI_ICON),

	models.Currency(emoji.BICYCLE_ICON),
	models.Currency(emoji.HAMMER_ICON),
	models.Currency(emoji.FORK_AND_KNIFE_ICON),
	models.Currency(emoji.DRESS_ICON),
	models.Currency(emoji.HIGH_HEELED_SHOES_ICON),
	models.Currency(emoji.TSHIRT_ICON),
}

func AskTransferCurrencyButtons(whc bots.WebhookContext) [][]string {
	user, _ := whc.GetAppUser()
	user.GetPreferredLocale()

	var (
		row, col int
	)

	result := make([][]string, 3)

	var runesInRow int

	var alreadyAddedCurrencies []models.Currency

	addCurrencyAndNewLineIfNeeded := func(currency models.Currency) {
		result[row] = append(result[row], currency.SignAndCode())
		runesInRow += utf8.RuneCountInString(currency.SignAndCode()) // TODO: Proper runes count
		col += 1
		if runesInRow > 16 || col >= 6 {
			row += 1
			if len(result) == row {
				result = append(result, []string{})
			}
			col = 0
			runesInRow = 0
		}
	}

	appUser := user.(*models.AppUserEntity)

	for _, currency := range appUser.GetCurrencies() {
		curr := models.Currency(currency)
		addCurrencyAndNewLineIfNeeded(curr)
		alreadyAddedCurrencies = append(alreadyAddedCurrencies, curr)
	}

	alreadyAdded := func(currency models.Currency) bool {
		for _, curr := range alreadyAddedCurrencies {
			if curr == currency {
				return true
			}
		}
		return false
	}

	for _, currency := range currenciesByPriority {
		if alreadyAdded(currency) {
			continue
		}
		addCurrencyAndNewLineIfNeeded(currency)
	}

	result = append(result, []string{whc.Translate(trans.COMMAND_TEXT_CANCEL)})

	return result
}

func AskTransferAmountCommand(code, messageTextFormat string, nextCommand bots.Command) bots.Command {
	return bots.Command{
		Code:    code,
		Replies: []bots.Command{nextCommand},
		Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
			c := whc.Context()

			//amount := 0
			//whc.chatEntity.AwaitingReplyTo = fmt.Sprintf("%v>%v?%v&amount=%v", whc.AwaitingReplyToPath(), code, whc.AwaitingReplyToQuery(), amount)

			chatEntity := whc.ChatEntity()
			awaitingReplyTo := chatEntity.GetAwaitingReplyTo()
			awaitingReplyToPath := bots.AwaitingReplyToPath(awaitingReplyTo)
			switch {
			case chatEntity.IsAwaitingReplyTo(code):
				switch whc.Input().(type) {
				case bots.WebhookTextMessage:
					mt := strings.TrimSpace(whc.Input().(bots.WebhookTextMessage).Text())
					if mt == "." || mt == "0" || strings.Index(mt, emoji.NO_ENTRY_SIGN_ICON) >= 0 {
						return CancelTransferWizardCommand.Action(whc)
					}
					if strings.Count(mt, ",") == 1 && strings.Count(mt, ".") == 0 {
						// handles numbers like 12,34
						mt = strings.Replace(mt, ",", ".", 1)
					} else if strings.Count(mt, ".") == 1 && strings.Count(mt, ",") > 0 && strings.Index(mt, ",") < strings.Index(mt, ".") {
						// handles numbers like 12,345.67
						mt = strings.Replace(mt, ",", "", -1)
					}
					if _, err := strconv.ParseFloat(mt, 64); err != nil {
						err = nil
						m = whc.NewMessage(emoji.NO_ENTRY_SIGN_ICON +
							" " + whc.Translate(trans.MESSAGE_TEXT_INVALID_FLOAT) +
							"\n\n" + whc.Translate(messageTextFormat, html.EscapeString(chatEntity.GetWizardParam("currency"))))
						m.Format = bots.MessageFormatHTML
					} else {
						chatEntity.AddWizardParam("value", mt)
						return nextCommand.Action(whc)
					}
				case bots.WebhookContactMessage:
					m.Text = whc.Translate("Please enter amount now, and then contact.")
					return
				default:
					m.Text = whc.Translate("Please enter amount now.")
					return
				}
			case strings.Contains(awaitingReplyToPath, code):
				//if strings.Contains(messageText, "%v") {
				//	amountValue, err := strconv.ParseFloat(params.Get("amount"), 64)
				//	if err != nil {
				//		return m, err
				//	}
				//	amount := models.AmountTotal{Currency: models.Currency(params.Get("currency")), Value: amountValue}
				//	messageText = fmt.Sprintf(messageText, amount)
				//}
				return m, fmt.Errorf("Command %v is incorrectly matched, whc.AwaitingReplyToPath(): %v", code, awaitingReplyToPath)
			default:
				chatEntity.PushStepToAwaitingReplyTo(code)
				currencyText := chatEntity.GetWizardParam("currency")
				if currencyText == "" {
					awaitingReplyToQuery := bots.AwaitingReplyToQuery(awaitingReplyTo)
					log.Warningf(c, "No currency in params: %v", awaitingReplyToQuery)
				}
				m = whc.NewMessageByCode(messageTextFormat, html.EscapeString(currencyText))
				if len(currencyText) == 3 && currencyText == strings.ToUpper(currencyText) {
					m.Keyboard = &tgbotapi.ReplyKeyboardHide{HideKeyboard: true}
				} else {
					m.Keyboard = tgbotapi.NewReplyKeyboardUsingStrings(
						[][]string{
							{
								"1", "2", "3", "4", "5",
							},
							{
								"6", "7", "8", "9", "10",
							},
							{
								emoji.NO_ENTRY_SIGN_ICON + " " + whc.Translate(trans.COMMAND_TEXT_CANCEL),
							},
						},
					)
				}
			}
			m.Format = bots.MessageFormatHTML
			return m, nil
		},
	}
}

type _onContactSelectedAction func(whc bots.WebhookContext, counterparty models.Contact) (m bots.MessageFromBot, err error)

func CreateAskTransferCounterpartyCommand(
	isReturn bool,
	code, title, icon, messageText string,
	replies []bots.Command,
	newContactCommand bots.Command,
	onContactSelectedAction _onContactSelectedAction,
) bots.Command {
	if newContactCommand.Code != "" {
		replies = append(replies, newContactCommand)
	}
	return bots.Command{
		Code:    code,
		Title:   title,
		Icon:    icon,
		Replies: replies,
		Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
			c := whc.Context()
			//amount := 0
			//whc.chatEntity.AwaitingReplyTo = fmt.Sprintf("%v>%v?%v&amount=%v", whc.AwaitingReplyToPath(), code, whc.AwaitingReplyToQuery(), amount)

			log.Debugf(c, "AskTransferCounterpartyCommand.Action(command.code=%v)", code)
			chatEntity := whc.ChatEntity()
			awaitingReplyTo := chatEntity.GetAwaitingReplyTo()
			awaitingReplyToPath := bots.AwaitingReplyToPath(awaitingReplyTo)
			switch {
			case strings.HasSuffix(awaitingReplyToPath, code): // If ends with it's own code display list of counterparties
				log.Debugf(c, "strings.HasSuffix(awaitingReplyToPath, code)")
				input := whc.Input()
				switch input.(type) {
				case bots.WebhookContactMessage:
					chatEntity.PushStepToAwaitingReplyTo(newContactCommand.Code)
					return newContactCommand.Action(whc)
				case bots.WebhookTextMessage:
					mt := whc.Input().(bots.WebhookTextMessage).Text()
					if mt == "." {
						return cancelTransferWizardCommandAction(whc)
					}
					var contactIDs []int64
					if contactIDs, err = dal.Contact.GetContactIDsByTitle(c, whc.AppUserIntID(), mt, true); err != nil {
						return m, err
					}
					if mt == whc.Translate(trans.COMMAND_TEXT_SHOW_ALL_CONTACTS) {
						log.Debugf(c, "mt == whc.Translate(trans.COMMAND_TEXT_SHOW_ALL_CONTACTS)")
						m, err = listCounterpartiesAsButtons(whc, models.AppUser{}, isReturn, messageText, newContactCommand)
					} else {
						log.Debugf(c, "mt != whc.Translate(trans.COMMAND_TEXT_SHOW_ALL_CONTACTS), len(contactIDs): %v", len(contactIDs))
						switch len(contactIDs) {
						case 1:
							contactID := contactIDs[0]
							chatEntity.AddWizardParam(WIZARD_PARAM_COUNTERPARTY, strconv.FormatInt(contactID, 10))
							var contact models.Contact
							if contact, err = facade.GetContactByID(c, contactID); err != nil {
								return
							}
							m, err = onContactSelectedAction(whc, contact)
						case 0:
							m = whc.NewMessage(whc.Translate(trans.MESSAGE_TEXT_UNKNOWN_COUNTERPARTY))
						default:
							m = whc.NewMessage(whc.Translate("Too many counterparties found: %v", contactIDs))
						}
					}
				default:
					err = fmt.Errorf("Unsupported message type: %T", input)
					return
				}
				return m, err
			case strings.Contains(awaitingReplyToPath, code):
				log.Debugf(c, "strings.Contains(awaitingReplyToPath, code)")
				return m, fmt.Errorf("Command %v is incorrectly matched, whc.AwaitingReplyToPath(): %v", code, awaitingReplyToPath)
			default:
				log.Debugf(c, "default:")
				var user models.AppUser
				if user, err = facade.User.GetUserByID(c, whc.AppUserIntID()); err != nil {
					return
				}
				if isReturn && user.BalanceCount <= 3 && user.TotalContactsCount() <= 3 {
					// If there is little debts in total show selection of debts immediately
					counterparties, err := dal.Contact.GetLatestContacts(whc, 0, user.TotalContactsCount())
					if err != nil {
						return m, err
					}
					var buttons [][]string

					var isTooManyRows bool
					now := time.Now()
					for _, counterparty := range counterparties {
						balance, err := counterparty.BalanceWithInterest(c, now)
						if err != nil {
							log.Errorf(c, "Failed to get balance with interest for contact %v: %v", counterparty.ID, err)
							buttons = append(buttons, []string{emoji.ERROR_ICON + " ERROR: " + counterparty.FullName()})
							continue
						}
						if (len(buttons) + len(balance)) > 4 {
							isTooManyRows = true
							log.Warningf(c, "Consider performance optimization - duplicate queries to get counterparties")
							break
						}
						for currency, value := range balance {
							buttons = append(buttons, []string{_debtAmountButtonText(whc, currency, value, counterparty)})
						}
					}
					if !isTooManyRows {
						m = askToChooseDebt(whc, buttons)
						return m, err
					}
				}

				chatEntity.PushStepToAwaitingReplyTo(code)
				m, err = listCounterpartiesAsButtons(whc, user, isReturn, messageText, newContactCommand)
				return m, err
			}
		},
	}
}

const COUNTERPARTY_BUTTONS_LIMIT = 4

func listCounterpartiesAsButtons(whc bots.WebhookContext, user models.AppUser, isReturn bool, messageText string, newCounterpartyCommand bots.Command,
) (m bots.MessageFromBot, err error) {
	c := whc.Context()

	log.Debugf(c, "listCounterpartiesAsButtons")
	queryString, err := url.ParseQuery(bots.AwaitingReplyToQuery(whc.ChatEntity().GetAwaitingReplyTo()))
	if err != nil {
		return m, err
	}
	if len(queryString) > 0 {
		currency := queryString.Get("currency")
		valueS := queryString.Get("value")
		value, err := decimal.ParseDecimal64p2(valueS)
		if err != nil {
			return m, err
		}
		amount := models.Amount{Currency: models.Currency(currency), Value: value}
		m = whc.NewMessage(fmt.Sprintf(whc.Translate(messageText), amount))
	} else {
		m = whc.NewMessage(whc.Translate(messageText))
	}
	m.Format = bots.MessageFormatHTML
	if user.AppUserEntity == nil {
		if user, err = facade.User.GetUserByID(c, whc.AppUserIntID()); err != nil {
			return
		}
	}
	var showAllContactsText = whc.Translate(trans.COMMAND_TEXT_SHOW_ALL_CONTACTS)

	buttons := [][]string{}
	var counterparties2buttons = func(counterparties []models.UserContactJson, isShowingAll bool) {
		for _, counterparty := range counterparties {
			buttons = append(buttons, []string{counterparty.Name})
		}
		var controlButtons []string
		if !isShowingAll && len(counterparties) < user.TotalContactsCount() {
			controlButtons = append(controlButtons, showAllContactsText)
		}
		if newCounterpartyCommand.Code != "" {
			controlButtons = append(controlButtons, newCounterpartyCommand.DefaultTitle(whc))
		}
		if len(controlButtons) > 0 {
			buttons = append(buttons, controlButtons)
		}
	}
	if webhookMessage, ok := whc.Input().(bots.WebhookTextMessage); ok && webhookMessage.Text() == showAllContactsText {
		counterparties2buttons(user.Contacts(), true)
	} else {
		switch user.BalanceCount {
		case 0: // User have no active debts
			if user.TotalContactsCount() > 0 {
				counterparties := user.LatestCounterparties(COUNTERPARTY_BUTTONS_LIMIT)
				counterparties2buttons(counterparties, false)
			} else {
				return newCounterpartyCommand.Action(whc)
			}
		default: // User have active debts (balance is not 0.

			counterpartiesToShow := user.ActiveContactsWithBalance()
			if len(counterpartiesToShow) <= COUNTERPARTY_BUTTONS_LIMIT {
				latestCounterparties := user.LatestCounterparties(COUNTERPARTY_BUTTONS_LIMIT)
				for _, latestCounterparty := range latestCounterparties {
					var isInWithDebts bool
					for _, counterpartyToShow := range counterpartiesToShow {
						if counterpartyToShow.ID == latestCounterparty.ID {
							isInWithDebts = true
							break
						}
					}
					if !isInWithDebts {
						counterpartiesToShow = append(counterpartiesToShow, latestCounterparty)
					}
					if len(counterpartiesToShow) >= COUNTERPARTY_BUTTONS_LIMIT {
						break
					}
				}
			}
			counterparties2buttons(counterpartiesToShow, false)
		}
	}
	if len(buttons) > 0 {
		keyboard := tgbotapi.NewReplyKeyboardUsingStrings(buttons)
		keyboard.OneTimeKeyboard = true
		m.Keyboard = keyboard
	}
	return m, nil
}

type TransferWizard struct {
	params url.Values
}

func NewTransferWizard(whc bots.WebhookContext) (TransferWizard, error) {
	awaitingReplyTo := whc.ChatEntity().GetAwaitingReplyTo()
	log.Debugf(whc.Context(), "AwaitingReplyTo: %v", awaitingReplyTo)
	params, err := url.ParseQuery(bots.AwaitingReplyToQuery(awaitingReplyTo))
	return TransferWizard{params: params}, err
}

func (w TransferWizard) CounterpartyID(c context.Context) int64 {
	s := w.params.Get(WIZARD_PARAM_COUNTERPARTY)
	if s == "" {
		s = w.params.Get(WIZARD_PARAM_CONTACT)
	}
	if s == "" {
		log.Debugf(c, "Wizard params: %v", w.params)
		return 0
	} else {
		counterpartyId, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			panic(err.Error())
		}
		return counterpartyId
	}
}

func TransferWizardCompletedCommand(code string) bots.Command {
	return bots.Command{
		Code: code,
		Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
			c := whc.Context()

			if chatEntity := whc.ChatEntity(); strings.Contains(chatEntity.GetAwaitingReplyTo(), "counterparty=0") {
				return dtb_general.MainMenuAction(whc, trans.MESSGE_TEXT_DEBT_ERROR_FIXED_START_OVER, false)
			}

			log.Infof(c, "TransferWizardCompletedCommand(code=%v).Action()", code)
			transferWizard, err := NewTransferWizard(whc)
			if err != nil {
				return m, err
			}

			var direction models.TransferDirection
			switch {
			case strings.HasPrefix(code, "transfer-from-"):
				direction = models.TransferDirectionUser2Counterparty
			case strings.HasPrefix(code, "transfer-to-"):
				direction = models.TransferDirectionCounterparty2User
			default:
				return m, fmt.Errorf("Can not decide direction due to unknown code: %v", code)
			}

			counterpartyId := transferWizard.CounterpartyID(c)
			params := transferWizard.params

			value, err := decimal.ParseDecimal64p2(params.Get("value"))
			if err != nil {
				return m, err
			}
			value = value.Abs()
			currencyCode := params.Get("currency")
			currency := models.Currency(currencyCode)

			var dueOn time.Time
			due := params.Get("due")
			if due != "" {
				if strings.Count(due, "-") == 2 {
					if dueOn, err = time.Parse(TRANSFER_WIZARD_DUE_DATE_FORMAT, due); err != nil {
						return m, errors.Wrap(err, "Failed to parse due date")
					}
					hour := time.Now().Hour()
					if hour <= 22 {
						hour += 1
					}
					hours, _ := time.ParseDuration(fmt.Sprintf("%vh", hour))
					dueOn = dueOn.Add(hours)
				} else {
					if dueIn, err := time.ParseDuration(due); err != nil {
						return m, errors.Wrap(err, "Fauled to parse due duration")
					} else if dueIn > time.Duration(0) {
						dueOn = time.Now().Add(dueIn)
					}
				}
			}

			amount := models.Amount{Currency: currency, Value: value}

			creatorInfo := models.TransferCounterpartyInfo{
				UserID:      whc.AppUserIntID(),
				ContactID:   counterpartyId,
				ContactName: "",
			}
			//if note := params.Get(TRANSFER_WIZARD_PARAM_NOTE); note != "" {
			//	creatorInfo.Note = note
			//}

			if comment := params.Get(TRANSFER_WIZARD_PARAM_COMMENT); comment != "" {
				creatorInfo.Comment = comment
			}

			var transferInterest models.TransferInterest

			if interest := params.Get(TRANSFER_WIZARD_PARAM_INTEREST); interest != "" {
				if transferInterest, err = getInterestData(interest); err != nil {
					return m, err
				}
			}

			m, err = CreateTransferFromBot(whc, false, 0, direction, creatorInfo, amount, dueOn, transferInterest)
			if err != nil {
				return m, err
			}

			//SetMainMenuKeyboard(whc, &m)
			whc.ChatEntity().SetAwaitingReplyTo("")
			return m, nil
		},
	}
}

const TRANSFER_WIZARD_DUE_DATE_FORMAT = "2006-01-02"

func CreateTransferFromBot(
	whc bots.WebhookContext,
	isReturn bool,
	returnToTransferID int64,
	direction models.TransferDirection,
	creatorInfo models.TransferCounterpartyInfo,
	amount models.Amount,
	dueOn time.Time,
	transferInterest models.TransferInterest,
) (
	m bots.MessageFromBot,
	err error,
) {
	c := whc.Context()

	if returnToTransferID != 0 && !isReturn {
		panic("returnToTransferID != 0 && !isReturn")
	}

	pleaseWaitMessageConfig := whc.NewMessage(emoji.HOURGLASS_ICON + " " + whc.Translate(trans.MESSAGE_TEXT_TRANSFER_IS_CREATING))
	pleaseWaitMessageConfig.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(whc.Translate(trans.COMMAND_TEXT_PLEASE_WAIT), dtb_general.PLEASE_WAIT_COMMAND),
		},
	)

	pleaseWaitMessage, err := whc.Responder().SendMessage(c, pleaseWaitMessageConfig, bots.BotAPISendMessageOverHTTPS)

	if err != nil {
		return m, err
	}

	from, to := facade.TransferCounterparties(direction, creatorInfo)

	var appUser models.AppUser
	if appUser, err = facade.User.GetUserByID(c, whc.AppUserIntID()); err != nil {
		return
	}
	newTransfer := facade.NewTransferInput(whc.Environment(),
		GetTransferSource(whc),
		appUser,
		"",
		isReturn,
		returnToTransferID,
		from, to,
		amount,
		dueOn,
		transferInterest,
	)

	output, err := facade.Transfers.CreateTransfer(whc.Context(), newTransfer)

	if err != nil {
		switch err {
		case facade.ErrNoOutstandingTransfers:
			m.Text = whc.Translate(trans.MT_NO_OUTSTANDING_TRANSFERS)
			log.Warningf(c, "Attempt to create return but no outstanding debts: %v", err)
			return
		case facade.ErrAttemptToCreateDebtWithInterestAffectingOutstandingTransfers:
			err = nil
			buf := new(bytes.Buffer)
			buf.WriteString(whc.Translate(trans.MT_ATTEMPT_TO_CREATE_DEBT_WITH_INTEREST_AFFECTING_OUTSTANDING) + "\n")
			now := time.Now()
			if outstandingTransfer, err := dal.Transfer.LoadOutstandingTransfers(c, now, appUser.ID, creatorInfo.ContactID, amount.Currency, newTransfer.Direction().Reverse()); err != nil {
				buf.WriteString(errors.WithMessage(err, "failed to load outstanding transfers").Error() + "\n")
			} else if len(outstandingTransfer) == 0 {
				return m, errors.WithMessage(err, "got facade.ErrAttemptToCreateDebtWithInterestAffectingOutstandingTransfers but no outstanding transfers found")
			} else {
				for _, ot := range outstandingTransfer {
					fmt.Fprintf(buf, "\tDebt #%v for %v => outstanding: %v\n", ot.ID, ot.GetAmount(), ot.GetOutstandingAmount(now))
				}
			}
			m.Text = buf.String()
			return m, err
		}
		log.Errorf(c, "Failed to create transfer: %v", err)
		if errors.Cause(err) == facade.ErrNotImplemented {
			m.Text = whc.Translate(trans.MESSAGE_TEXT_NOT_IMPLEMENTED_YET) + "\n\n" + err.Error()
			err = nil
		}
		return m, err
	}

	log.Debugf(c, "isReturn: %v, transfer.IsReturn: %v", isReturn, output.Transfer.IsReturn)

	{ // Reporting to Google Analytics
		ga := whc.GA()

		gaEventLabel := string(output.Transfer.Currency)
		if len([]rune(gaEventLabel)) > 16 {
			gaEventLabel = string([]rune(string(output.Transfer.Currency))[:16])
		}
		var action string
		if isReturn {
			if len(output.ReturnedTransfers) == 1 && !output.ReturnedTransfers[0].IsOutstanding {
				action = "debt-returned-fully"
			} else {
				action = "debt-returned-partially"
			}
		} else {
			action = "debt-new-created"
		}
		gaEvent := ga.GaEventWithLabel(analytics.EventCategoryTransfers, action, gaEventLabel)
		gaEvent.Value = uint(math.Abs(output.Transfer.AmountInCents.AsFloat64()) + 0.5)

		if gaErr := ga.Queue(gaEvent); gaErr != nil {
			log.Warningf(c, "Failed to log event: %v", gaErr)
		} else {
			log.Infof(c, "GA event queued: %v", gaEvent)
		}

		if !output.Transfer.DtDueOn.IsZero() {
			gaEvent = ga.GaEvent(analytics.EventCategoryTransfers, analytics.EventActionDebtDueDateSet)
			//Do not set event value!: gaEvent.Value = uint(transfer.DtDueOn.Sub(time.Now()) / time.Hour)
			if gaErr := ga.Queue(gaEvent); gaErr != nil {
				log.Warningf(c, "Failed to log event: %v", gaErr)
			} else {
				log.Infof(c, "GA event queued: %v", gaEvent)
			}
		}
	}

	{
		utm := common.NewUtmParams(whc, common.UTM_CAMPAIGN_RECEIPT)
		receiptMessageText := common.TextReceiptForTransfer(whc, output.Transfer, whc.AppUserIntID(), common.ShowReceiptToAutodetect, utm)

		switch whc.BotPlatform().ID() {
		case telegram.PlatformID:
			var receiptMessageFromBot bots.MessageFromBot
			if receiptMessageFromBot, err = whc.NewEditMessage(receiptMessageText, bots.MessageFormatHTML); err != nil {
				return receiptMessageFromBot, err
			}
			receiptMessageFromBot.EditMessageUID = telegram.NewChatMessageUID(0, pleaseWaitMessage.TelegramMessage.(tgbotapi.Message).MessageID)
			_, err = whc.Responder().SendMessage(c, receiptMessageFromBot, bots.BotAPISendMessageOverHTTPS)
			if err != nil {
				return m, err
			}
			if receiptSendOptionsMessage, err := createSendReceiptOptionsMessage(whc, output.Transfer); err != nil {
				return m, err
			} else {
				if response, err := whc.Responder().SendMessage(c, receiptSendOptionsMessage, bots.BotAPISendMessageOverHTTPS); err != nil {
					return m, err
				} else {
					tgMessage := response.TelegramMessage.(tgbotapi.Message)
					if err = dal.Transfer.DelayUpdateTransferWithCreatorReceiptTgMessageID(whc.Context(), whc.GetBotCode(), output.Transfer.ID, tgMessage.Chat.ID, int64(tgMessage.MessageID)); err != nil {
						return m, err
					}
					whc.ChatEntity().SetAwaitingReplyTo("")
				}
			}
		case viber.PlatformID:
			receiptMessageFromBot := whc.NewMessage(receiptMessageText)
			whc.Responder().SendMessage(c, receiptMessageFromBot, bots.BotAPISendMessageOverHTTPS)
		default:
			panic("Unsupported bot platform: " + whc.BotPlatform().ID())
		}
	}

	return dtb_general.MainMenuAction(whc, "", false)
	//
	//
	//return m, err
}

func sendReceiptByTelegramButton(transferEncodedID string, translator strongo.SingleLocaleTranslator) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonSwitchInlineQuery(
		translator.Translate(trans.COMMAND_TEXT_SEND_RECEIPT_BY_TELEGRAM),
		fmt.Sprintf("receipt?id=%v", transferEncodedID),
	)
}

func createSendReceiptOptionsMessage(whc bots.WebhookContext, transfer models.Transfer) (m bots.MessageFromBot, err error) {
	c := whc.Context()

	log.Debugf(c, "createSendReceiptOptionsMessage(transferID=%v)", transfer.ID)
	mt := whc.Translate(trans.MESSAGE_TEXT_YOU_CAN_SEND_RECEIPT, html.EscapeString(transfer.Counterparty().ContactName))
	var utmCampaign string
	if transfer.IsReturn {
		utmCampaign = common.UTM_CAMPAIGN_DEBT_RETURNED
	} else {
		utmCampaign = common.UTM_CAMPAIGN_DEBT_CREATED
	}
	utmParams := common.NewUtmParams(whc, utmCampaign)
	transferUrlForUser := common.GetTransferUrlForUser(transfer.ID, whc.AppUserIntID(), whc.Locale(), utmParams)
	mt = strings.Replace(mt, "<a receipt>", fmt.Sprintf(`<a href="%v">`, transferUrlForUser), 1)
	mt = strings.Replace(mt, "<a counterparty>", fmt.Sprintf(`<a href="%v">`, common.GetCounterpartyUrl(transfer.Counterparty().ContactID, whc.AppUserIntID(), whc.Locale(), utmParams)), 1)

	if whc.InputType() == bots.WebhookInputCallbackQuery {
		if m, err = whc.NewEditMessage(mt, bots.MessageFormatHTML); err != nil {
			return
		}
	} else {
		m = whc.NewMessage(mt)
		m.Format = bots.MessageFormatHTML
	}

	transferEncodedID := common.EncodeID(transfer.ID)
	transferDecodedID, err := common.DecodeID(transferEncodedID)
	if err != nil {
		panic(fmt.Sprintf("Failed to decode transferEncodedID:%v that was encoded from %v", transferEncodedID, transfer.ID))
	}
	if transferDecodedID != transfer.ID {
		panic("transferDecodedID != transferRawID")
	}
	log.Debugf(c, "transferID: %v, transferEncodedID: %v", transfer.ID, transferEncodedID)

	m.DisableWebPagePreview = true
	var telegramKeyboard tgbotapi.InlineKeyboardMarkup
	var isCounterpartyUserHasTelegram bool
	if transfer.Creator().ContactID != 0 {
		if user, err := facade.User.GetUserByID(c, transfer.Counterparty().UserID); err != nil {
			err = errors.Wrapf(err, "Failed to get counterparty user by ID=%v", transfer.Counterparty().UserID)
			return m, err
		} else {
			isCounterpartyUserHasTelegram = user.HasTelegramAccount()
			log.Debugf(c, "isCounterpartyUserHasTelegram: %v, transfer.Creator().ContactID: %v, user.GetTelegramUserIDs(): %v", isCounterpartyUserHasTelegram, transfer.Creator().ContactID, user.GetTelegramUserIDs())
		}
	} else {
		log.Debugf(c, "isCounterpartyUserHasTelegram: %v, transfer.Creator().ContactID: %v", isCounterpartyUserHasTelegram, transfer.Creator().ContactID)
	}

	if isCounterpartyUserHasTelegram {
		m.Text = emoji.HOURGLASS_ICON + " " + fmt.Sprintf(whc.Translate(trans.MESSAGE_TEXT_RECEIPT_IS_SENDING_BY_TELEGRAM), transfer.Counterparty().ContactName)
	} else {
		telegramKeyboard.InlineKeyboard = [][]tgbotapi.InlineKeyboardButton{
			{sendReceiptByTelegramButton(transferEncodedID, whc)},
		}
		utmParams := common.UtmParams{
			Source:   telegram.PlatformID,
			Medium:   common.UTM_MEDIUM_BOT,
			Campaign: common.UTM_CAMPAIGN_TRANSFER_SEND_RECEIPT,
		}
		transferUrl := common.GetTransferUrlForUser(transfer.ID, whc.AppUserIntID(), whc.Locale(), utmParams)

		transferUrl += "&send=menu"

		telegramKeyboard.InlineKeyboard = append(
			telegramKeyboard.InlineKeyboard,
			[]tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonURL(
					whc.Translate(trans.COMMAND_TEXT_COUNTERPARTY_HAS_NO_TELEGRAM),
					transferUrl,
				),
			},
			[]tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData(
					whc.Translate(trans.COMMAND_TEXT_SEND_RECEIPT_BY_SMS),
					SendReceiptCallbackData(transfer.ID, "sms"),
				),
			},
			[]tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData(
					whc.Translate(trans.COMMAND_TEXT_GET_LINK_FOR_RECEIPT_IN_TELEGRAM),
					SendReceiptCallbackData(transfer.ID, string(models.InviteByLinkToTelegram)),
				),
			},
		)
	}
	telegramKeyboard.InlineKeyboard = append(
		telegramKeyboard.InlineKeyboard,
		[]tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(
				whc.Translate(trans.COMMAND_TEXT_DO_NOT_SEND_RECEIPT),
				SendReceiptCallbackData(transfer.ID, RECEIPT_ACTION__DO_NOT_SEND), // TODO: Replace path with constant
			),
		},
	)
	m.Keyboard = &telegramKeyboard
	return m, err
}

//var TomorrowCommand = Command{
//	code: "tomorrow",
//	title: COMMAND_TEXT_TOMORROW,
//}

func GetTransferSource(whc bots.WebhookContext) dal.TransferSource {
	return dal.NewTransferSourceBot(whc.BotPlatform().ID(), whc.GetBotCode(), whc.MustBotChatID())
}

//const CALLBACK_COUNTERPARTY_WITHOUT_TG = "counterparty-no-tg"
//
//var CounterpartyNoTelegramCommand = bots.Command{
//	ByCode: CALLBACK_COUNTERPARTY_WITHOUT_TG,
//	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
//		q := callbackUrl.Query()
//		transferEncodedID := q.Get(WIZARD_PARAM_TRANSFER)
//		transferID, err := common.DecodeID(transferEncodedID)
//		if err != nil {
//			return m, err
//		}
//
//		kbMarkup := tgbotapi.NewInlineKeyboardMarkup(
//			[]tgbotapi.InlineKeyboardButton{
//				{Text: whc.Translate(trans.COMMAND_TEXT_SEND_RECEIPT_BY_SMS), CallbackData: SendReceiptCallbackData(transferID, "sms")},
//			},
//		)
//		hide := q.Get("hide")
//		if hide == "" {
//			//localeSiteCode := whc.Locale().SiteCode()
//			//sendReceiptPageUrl := func(by string) string {
//			//	return fmt.Sprintf("https://debtstracker.io/app/send-receipt?by=%v&%v=%v&lang=%v", WIZARD_PARAM_TRANSFER, by, transferID, localeSiteCode)
//			//}
//			callbackHide := func(by string) string {
//				return fmt.Sprintf("%v?hide=%v&%v=%v", CALLBACK_COUNTERPARTY_WITHOUT_TG, by, WIZARD_PARAM_TRANSFER, transferEncodedID)
//			}
//			kbMarkup.InlineKeyboard = append(
//				kbMarkup.InlineKeyboard,
//				[]tgbotapi.InlineKeyboardButton{{Text: whc.Translate(trans.COMMAND_TEXT_SEND_RECEIPT_BY_VK), CallbackData: callbackHide("vk")}},
//				[]tgbotapi.InlineKeyboardButton{{Text: whc.Translate(trans.COMMAND_TEXT_SEND_RECEIPT_BY_FB), CallbackData: callbackHide("fb")}},
//				[]tgbotapi.InlineKeyboardButton{{Text: whc.Translate(trans.COMMAND_TEXT_SEND_RECEIPT_BY_OK), CallbackData: callbackHide("ok")}},
//				//[]tgbotapi.InlineKeyboardButton{{Text: whc.Translate(trans.COMMAND_TEXT_SEND_RECEIPT_BY_TWT), CallbackData: callbackHide("twt")}},
//			)
//			shuffled := make([][]tgbotapi.InlineKeyboardButton, len(kbMarkup.InlineKeyboard))
//			for i, v := range rand.Perm(len(shuffled)) {
//				 shuffled[v] = kbMarkup.InlineKeyboard[i]
//			}
//			kbMarkup.InlineKeyboard = shuffled
//		} else {
//			kbMarkup.InlineKeyboard = append(
//				[][]tgbotapi.InlineKeyboardButton{
//					[]tgbotapi.InlineKeyboardButton{sendReceiptByTelegramButton(transferEncodedID, whc)},
//				},
//				kbMarkup.InlineKeyboard...,
//			)
//			whc.GA().Queue(measurement.NewEvent("receipt", "send-by-"+hide, whc.GaCommon()))
//		}
//
//		m = telegram.NewEditMessageKeyboard(whc, kbMarkup)
//		m.Text = whc.Translate(trans.MESSAGE_TEXT_RECEIPT_AVAILABLE_CHANNELS)
//		return m, err
//	},
//}
