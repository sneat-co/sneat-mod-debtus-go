package dtb_transfer

import (
	"fmt"
	"github.com/bots-go-framework/bots-api-telegram/tgbotapi"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/crediterra/money"
	"github.com/sneat-co/debtstracker-translations/trans"
	"net/url"
	"strconv"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_general"
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/debtus/dtb_common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
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
	reminderID, err := common.DecodeIntID(q.Get("reminder"))
	var transferID int
	if err != nil {
		if q.Get("reminder") == "" { // TODO: Remove this obsolete branch
			if transferID, err = common.DecodeIntID(q.Get("id")); err != nil {
				return m, fmt.Errorf("failed to decode transfer ID: %w", err)
			}
		} else {
			return m, fmt.Errorf("failed to decode reminder ID: %w", err)
		}
	} else {
		if reminder, err := dtdal.Reminder.SetReminderStatus(c, reminderID, 0, models.ReminderStatusUsed, time.Now()); err != nil {
			return m, err
		} else {
			transferID = reminder.TransferID
		}
	}

	howMuch := q.Get("how-much")
	transfer, err := facade.Transfers.GetTransferByID(c, nil, transferID)
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
		reminderID int
		transfer   models.Transfer
	)
	if reminderID, err = common.DecodeIntID(q.Get("reminder")); err != nil {
		err = fmt.Errorf("can't decode parameter 'reminder': %w", err)
		return
	}
	if transfer.ID, err = common.DecodeIntID(q.Get("transfer")); err != nil {
		err = fmt.Errorf("can't decode parameter 'transfer': %w", err)
		return
	}

	if transfer, err = facade.Transfers.GetTransferByID(c, nil, transfer.ID); err != nil {
		return
	}

	return askWhenToRemindAgain(whc, reminderID, transfer)
})

func ProcessFullReturn(whc botsfw.WebhookContext, transfer models.Transfer) (m botsfw.MessageFromBot, err error) {
	amountValue := transfer.Data.GetOutstandingValue(time.Now())
	if amountValue == 0 {
		return dtb_general.EditReminderMessage(whc, transfer, whc.Translate(trans.MESSAGE_TEXT_TRANSFER_ALREADY_FULLY_RETURNED))
	} else if amountValue < 0 {
		err = fmt.Errorf("data integrity error -> transfer.GetOutstandingValue():%v < 0", amountValue)
		return
	}

	amount := money.NewAmount(transfer.Data.GetAmount().Currency, amountValue)

	var (
		counterpartyID int64
		direction      models.TransferDirection
	)
	userID := whc.AppUserIntID()
	if transfer.Data.CreatorUserID == userID {
		counterpartyID = transfer.Data.Counterparty().ContactID
		switch transfer.Data.Direction() {
		case models.TransferDirectionCounterparty2User:
			direction = models.TransferDirectionUser2Counterparty
		case models.TransferDirectionUser2Counterparty:
			direction = models.TransferDirectionCounterparty2User
		default:
			return m, fmt.Errorf("transfer %v has unknown direction '%v'", transfer.ID, transfer.Data.Direction())
		}
	} else if transfer.Data.Counterparty().UserID == userID {
		switch transfer.Data.Direction() {
		case models.TransferDirectionCounterparty2User:
		case models.TransferDirectionUser2Counterparty:
		default:
			return m, fmt.Errorf("transfer %v has unknown direction '%v'.", transfer.ID, transfer.Data.Direction())
		}
		counterpartyID = transfer.Data.Creator().ContactID
		direction = transfer.Data.Direction()
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

func ProcessPartialReturn(whc botsfw.WebhookContext, transfer models.Transfer) (botsfw.MessageFromBot, error) {
	var counterpartyID int64
	switch whc.AppUserIntID() {
	case transfer.Data.CreatorUserID:
		counterpartyID = transfer.Data.Counterparty().ContactID
	case transfer.Data.Counterparty().UserID:
		counterpartyID = transfer.Data.Creator().ContactID
	default:
		panic(fmt.Sprintf("whc.whc.AppUserIntID()=%v not in (transfer.Counterparty().ContactID=%v, transfer.Creator().ContactID=%v)",
			whc.AppUserIntID(), transfer.Data.Counterparty().ContactID, transfer.Data.Creator().ContactID))
	}
	chatEntity := whc.ChatEntity()
	chatEntity.SetAwaitingReplyTo("")
	chatEntity.AddWizardParam(WIZARD_PARAM_COUNTERPARTY, strconv.FormatInt(counterpartyID, 10))
	chatEntity.AddWizardParam(WIZARD_PARAM_TRANSFER, strconv.Itoa(transfer.ID))
	chatEntity.AddWizardParam("currency", string(transfer.Data.Currency))

	reportReminderIsActed(whc, "reminder-acted-returned-partially")

	return AskHowMuchHaveBeenReturnedCommand.Action(whc)
}

func askWhenToRemindAgain(whc botsfw.WebhookContext, reminderID int, transfer models.Transfer) (m botsfw.MessageFromBot, err error) {
	if m, err = dtb_general.EditReminderMessage(whc, transfer, whc.Translate(trans.MESSAGE_TEXT_ASK_WHEN_TO_REMIND_AGAIN)); err != nil {
		return
	}
	callbackData := fmt.Sprintf("%v?id=%v&in=%v", dtb_common.CALLBACK_REMIND_AGAIN, common.EncodeIntID(reminderID), "%v")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			{
				Text:         emoji.CALENDAR_ICON + " " + whc.Translate(trans.COMMAND_TEXT_SET_DATE),
				CallbackData: fmt.Sprintf("%v?id=%v", SET_NEXT_REMINDER_DATE_COMMAND, common.EncodeIntID(reminderID)),
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

func ProcessNoReturn(whc botsfw.WebhookContext, reminderID int, transfer models.Transfer) (m botsfw.MessageFromBot, err error) {
	return askWhenToRemindAgain(whc, reminderID, transfer)
}

const (
	SET_NEXT_REMINDER_DATE_COMMAND = "set-next-reminder-date"
)

var SetNextReminderDateCallbackCommand = botsfw.Command{
	Code: SET_NEXT_REMINDER_DATE_COMMAND,
	CallbackAction: func(whc botsfw.WebhookContext, callbackUrl *url.URL) (m botsfw.MessageFromBot, err error) {
		c := whc.Context()

		reminderID, err := common.DecodeIntID(callbackUrl.Query().Get("id"))
		if err != nil {
			return m, fmt.Errorf("failed to decode transfer id: %w", err)
		}

		chatEntity := whc.ChatEntity()
		chatEntity.SetAwaitingReplyTo(SET_NEXT_REMINDER_DATE_COMMAND)
		chatEntity.AddWizardParam(WIZARD_PARAM_REMINDER, strconv.Itoa(reminderID))

		reminder, err := dtdal.Reminder.GetReminderByID(c, nil, reminderID)
		if err != nil {
			return m, fmt.Errorf("failed to get reminder by id: %w", err)
		}
		transfer, err := facade.Transfers.GetTransferByID(c, nil, reminder.TransferID)
		if err != nil {
			return m, fmt.Errorf("failed to get transfer by id: %w", err)
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
	Action: func(whc botsfw.WebhookContext) (botsfw.MessageFromBot, error) {
		m, date, err := processSetDate(whc)
		if !date.IsZero() {
			chatEntity := whc.ChatEntity()

			encodedReminderID := chatEntity.GetWizardParam(WIZARD_PARAM_REMINDER)
			reminderID, err := strconv.Atoi(encodedReminderID)
			if err != nil {
				return m, fmt.Errorf("failed to decode reminder id: %w", err)
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
