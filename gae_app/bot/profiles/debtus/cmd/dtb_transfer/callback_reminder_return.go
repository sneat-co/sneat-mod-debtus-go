package dtb_transfer

import (
	"fmt"
	"github.com/crediterra/money"
	"net/url"
	"strconv"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_general"
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/debtus/dtb_common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"errors"
	"github.com/sneat-co/debtstracker-translations/emoji"
	"github.com/strongo/app"
	"github.com/strongo/log"
)

var ReturnCallbackCommand = botsfw.NewCallbackCommand(dtb_common.CALLBACK_DEBT_RETURNED_PATH, ProcessReturnAnswer)

func ProcessReturnAnswer(whc botsfw.WebhookContext, callbackUrl *url.URL) (m botsfw.MessageFromBot, err error) {
	//
	c := whc.Context()
	log.Debugf(c, "ProcessReturnAnswer()")
	q := callbackUrl.Query()
	reminderID, err := common.DecodeID(q.Get("reminder"))
	var transferID int64
	if err != nil {
		if q.Get("reminder") == "" { // TODO: Remove this obsolete branch
			if transferID, err = common.DecodeID(q.Get("id")); err != nil {
				return m, errors.Wrap(err, "Failed to decode transfer ID")
			}
		} else {
			return m, errors.Wrap(err, "Failed to decode reminder ID")
		}
	} else {
		if reminder, err := dtdal.Reminder.SetReminderStatus(c, reminderID, 0, models.ReminderStatusUsed, time.Now()); err != nil {
			return m, err
		} else {
			transferID = reminder.TransferID
		}
	}

	howMuch := q.Get("how-much")
	transfer, err := facade.Transfers.GetTransferByID(c, transferID)
	if err != nil {
		return m, err
	}
	switch howMuch {
	case "":
		panic("Missing how-much parameter")
	case dtb_common.RETURNED_FULLY:
		return ProcessFullReturn(whc, transfer)
	case dtb_common.RETURNED_PARTIALLY:
		return ProcessPartialReturn(whc, transfer)
	case dtb_common.RETURNED_NOTHING:
		return ProcessNoReturn(whc, reminderID, transfer)
	default:
		panic(fmt.Sprintf("Unknown how-much: %v", howMuch))
	}
}

const commandCodeEnableReminderAgain = "enable-reminder-again"

var EnableReminderAgainCallbackCommand = botsfw.NewCallbackCommand(commandCodeEnableReminderAgain, func(whc botsfw.WebhookContext, callbackUrl *url.URL) (m botsfw.MessageFromBot, err error) {
	c := whc.Context()
	log.Debugf(c, "EnableReminderAgainCallbackCommand()")
	q := callbackUrl.Query()
	var (
		reminderID int64
		transfer   models.Transfer
	)
	if reminderID, err = common.DecodeID(q.Get("reminder")); err != nil {
		err = errors.WithMessage(err, "Can't decode parameter 'reminder'")
		return
	}
	if transfer.ID, err = common.DecodeID(q.Get("transfer")); err != nil {
		err = errors.WithMessage(err, "Can't decode parameter 'transfer'")
		return
	}

	if transfer, err = facade.Transfers.GetTransferByID(c, transfer.ID); err != nil {
		return
	}

	return askWhenToRemindAgain(whc, reminderID, transfer)
})

func ProcessFullReturn(whc botsfw.WebhookContext, transfer models.Transfer) (m botsfw.MessageFromBot, err error) {
	amountValue := transfer.GetOutstandingValue(time.Now())
	if amountValue == 0 {
		return dtb_general.EditReminderMessage(whc, transfer, whc.Translate(trans.MESSAGE_TEXT_TRANSFER_ALREADY_FULLY_RETURNED))
	} else if amountValue < 0 {
		err = fmt.Errorf("data integrity error -> transfer.GetOutstandingValue():%v < 0", amountValue)
		return
	}

	amount := money.NewAmount(transfer.GetAmount().Currency, amountValue)

	var (
		counterpartyID int64
		direction      models.TransferDirection
	)
	userID := whc.AppUserIntID()
	if transfer.CreatorUserID == userID {
		counterpartyID = transfer.Counterparty().ContactID
		switch transfer.Direction() {
		case models.TransferDirectionCounterparty2User:
			direction = models.TransferDirectionUser2Counterparty
		case models.TransferDirectionUser2Counterparty:
			direction = models.TransferDirectionCounterparty2User
		default:
			return m, fmt.Errorf("Transfer %v has unknown direction '%v'.", transfer.ID, transfer.Direction())
		}
	} else if transfer.Counterparty().UserID == userID {
		switch transfer.Direction() {
		case models.TransferDirectionCounterparty2User:
		case models.TransferDirectionUser2Counterparty:
		default:
			return m, fmt.Errorf("Transfer %v has unknown direction '%v'.", transfer.ID, transfer.Direction())
		}
		counterpartyID = transfer.Creator().ContactID
		direction = transfer.Direction()
	}

	if m, err = dtb_general.EditReminderMessage(whc, transfer, whc.Translate(trans.MESSAGE_TEXT_REPLIED_DEBT_RETURNED_FULLY)); err != nil {
		return
	}

	if _, err = whc.Responder().SendMessage(whc.Context(), m, botsfw.BotAPISendMessageOverHTTPS); err != nil {
		return m, err
	}

	if m, err = CreateReturnAndShowReceipt(whc, transfer.ID, counterpartyID, direction, amount); err != nil {
		return m, err
	}

	reportReminderIsActed(whc, "reminder-acted-returned-fully")

	//TODO: edit message
	return m, err
}

func ProcessPartialReturn(whc botsfw.WebhookContext, transfer models.Transfer) (bots.MessageFromBot, error) {
	var counterpartyID int64
	switch whc.AppUserIntID() {
	case transfer.CreatorUserID:
		counterpartyID = transfer.Counterparty().ContactID
	case transfer.Counterparty().UserID:
		counterpartyID = transfer.Creator().ContactID
	default:
		panic(fmt.Sprintf("whc.whc.AppUserIntID()=%v not in (transfer.Counterparty().ContactID=%v, transfer.Creator().ContactID=%v)",
			whc.AppUserIntID(), transfer.Counterparty().ContactID, transfer.Creator().ContactID))
	}
	chatEntity := whc.ChatEntity()
	chatEntity.SetAwaitingReplyTo("")
	chatEntity.AddWizardParam(WIZARD_PARAM_COUNTERPARTY, strconv.FormatInt(counterpartyID, 10))
	chatEntity.AddWizardParam(WIZARD_PARAM_TRANSFER, strconv.FormatInt(transfer.ID, 10))
	chatEntity.AddWizardParam("currency", string(transfer.Currency))

	reportReminderIsActed(whc, "reminder-acted-returned-partially")

	return AskHowMuchHaveBeenReturnedCommand.Action(whc)
}

func askWhenToRemindAgain(whc botsfw.WebhookContext, reminderID int64, transfer models.Transfer) (m botsfw.MessageFromBot, err error) {
	if m, err = dtb_general.EditReminderMessage(whc, transfer, whc.Translate(trans.MESSAGE_TEXT_ASK_WHEN_TO_REMIND_AGAIN)); err != nil {
		return
	}
	callbackData := fmt.Sprintf("%v?id=%v&in=%v", dtb_common.CALLBACK_REMIND_AGAIN, common.EncodeID(reminderID), "%v")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			{
				Text:         emoji.CALENDAR_ICON + " " + whc.Translate(trans.COMMAND_TEXT_SET_DATE),
				CallbackData: fmt.Sprintf("%v?id=%v", SET_NEXT_REMINDER_DATE_COMMAND, common.EncodeID(reminderID)),
			},
		},
		[]tgbotapi.InlineKeyboardButton{
			{Text: whc.Translate(trans.COMMAND_TEXT_TOMORROW), CallbackData: fmt.Sprintf(callbackData, "24h")},
			{Text: whc.Translate(trans.COMMAND_TEXT_DAY_AFTER_TOMORROW), CallbackData: fmt.Sprintf(callbackData, "48h")},
		},
		[]tgbotapi.InlineKeyboardButton{
			{Text: whc.Translate(trans.COMMAND_TEXT_IN_1_WEEK), CallbackData: fmt.Sprintf(callbackData, "168h")},
			{Text: whc.Translate(trans.COMMAND_TEXT_IN_1_MONTH), CallbackData: fmt.Sprintf(callbackData, "720h")},
		},
		[]tgbotapi.InlineKeyboardButton{
			{Text: whc.Translate(trans.COMMAND_TEXT_DISABLE_REMINDER), CallbackData: fmt.Sprintf(callbackData, dtb_common.C_REMIND_IN_DISABLE)},
		},
	)

	if whc.GetBotSettings().Env == strongo.EnvDevTest {
		keyboard.InlineKeyboard = append(
			[][]tgbotapi.InlineKeyboardButton{
				{
					{
						Text:         whc.Translate(trans.COMMAND_TEXT_IN_FEW_MINUTES),
						CallbackData: fmt.Sprintf(callbackData, "1m"),
					},
				},
			},
			keyboard.InlineKeyboard...,
		)
	}
	m.IsEdit = true
	m.Keyboard = keyboard
	return
}

func ProcessNoReturn(whc botsfw.WebhookContext, reminderID int64, transfer models.Transfer) (m botsfw.MessageFromBot, err error) {
	return askWhenToRemindAgain(whc, reminderID, transfer)
}

const (
	SET_NEXT_REMINDER_DATE_COMMAND = "set-next-reminder-date"
)

var SetNextReminderDateCallbackCommand = botsfw.Command{
	Code: SET_NEXT_REMINDER_DATE_COMMAND,
	CallbackAction: func(whc botsfw.WebhookContext, callbackUrl *url.URL) (m botsfw.MessageFromBot, err error) {
		c := whc.Context()

		reminderID, err := common.DecodeID(callbackUrl.Query().Get("id"))
		if err != nil {
			return m, errors.Wrapf(err, "Failed to decode transfer id")
		}

		chatEntity := whc.ChatEntity()
		chatEntity.SetAwaitingReplyTo(SET_NEXT_REMINDER_DATE_COMMAND)
		chatEntity.AddWizardParam(WIZARD_PARAM_REMINDER, strconv.FormatInt(reminderID, 10))

		reminder, err := dtdal.Reminder.GetReminderByID(c, reminderID)
		if err != nil {
			return m, errors.Wrap(err, "Failed to get reminder by id")
		}
		transfer, err := facade.Transfers.GetTransferByID(c, reminder.TransferID)
		if err != nil {
			return m, errors.Wrap(err, "Failed to get transfer by id")
		}

		if m, err = dtb_general.EditReminderMessage(whc, transfer, whc.Translate(trans.MESSAGE_TEXT_ASK_WHEN_TO_REMIND_AGAIN)); err != nil {
			return
		}

		if _, err = whc.Responder().SendMessage(c, m, botsfw.BotAPISendMessageOverHTTPS); err != nil {
			return m, err
		}

		m = whc.NewMessageByCode(trans.MESSAGE_TEXT_ASK_DATE_TO_REMIND)

		return m, err
	},
	Action: func(whc botsfw.WebhookContext) (bots.MessageFromBot, error) {
		m, date, err := processSetDate(whc)
		if !date.IsZero() {
			chatEntity := whc.ChatEntity()

			encodedReminderID := chatEntity.GetWizardParam(WIZARD_PARAM_REMINDER)
			reminderID, err := strconv.ParseInt(encodedReminderID, 10, 64)
			if err != nil {
				return m, errors.Wrap(err, "Failed to decode reminder id")
			}
			now := time.Now()
			sinceToday := now.Sub(now.Truncate(24 * time.Hour))

			date = date.Add(sinceToday)
			remindInDuration := date.Sub(now)
			return rescheduleReminder(whc, reminderID, remindInDuration)
		}
		return m, err
	},
}
