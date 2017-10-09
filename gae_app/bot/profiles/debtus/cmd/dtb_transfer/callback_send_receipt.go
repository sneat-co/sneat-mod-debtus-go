package dtb_transfer

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/app/log"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"html"
	"net/url"
	"strconv"
	"strings"
)

var SendReceiptCallbackCommand = bots.NewCallbackCommand(SEND_RECEIPT_CALLBACK_PATH, CallbackSendReceipt)

func CallbackSendReceipt(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	q := callbackUrl.Query()
	sendBy := q.Get("by")

	log.Debugf(c, "CallbackSendReceipt(callbackUrl=%v)", callbackUrl)
	var (
		transferID int64
		transfer   models.Transfer
	)
	transferID, err = common.DecodeID(q.Get(WIZARD_PARAM_TRANSFER))
	if err != nil {
		return m, errors.Wrap(err, "Faield to decode transferID to int")
	}
	transfer, err = dal.Transfer.GetTransferByID(c, transferID)
	if err != nil {
		return m, errors.Wrap(err, "Failed to get transfer by ID")
	}
	//chatEntity := whc.ChatEntity() //TODO: Need this to get appUser, has to be refactored
	//appUser, err := whc.GetAppUser()
	counterparty, err := dal.Contact.GetContactByID(c, transfer.Counterparty().ContactID)
	if err != nil {
		return m, err
	}
	if IsTransferNotificationsBlockedForChannel(counterparty.ContactEntity, sendBy) {
		m = whc.NewMessage(trans.MESSAGE_TEXT_USER_BLOCKED_TRANSFER_NOTIFICATIONS_BY)
		return m, err
	}
	chatEntity := whc.ChatEntity()
	switch sendBy {
	case SEND_RECEIPT_BY_CHOOSE_CHANNEL:
		return createSendReceiptOptionsMessage(whc, transfer)
	case RECEIPT_ACTION__DO_NOT_SEND:
		log.Debugf(c, "CallbackSendReceipt(): do-not-send")
		if m, err = whc.NewEditMessage(whc.Translate(trans.MESSAGE_TEXT_RECEIPT_WILL_NOT_BE_SENT), bots.MessageFormatHTML); err != nil {
			return
		}

		// TODO: do type assertion
		if callbackMessage := whc.Input().(telegram_bot.TelegramWebhookCallbackQuery).TelegramCallbackMessage; callbackMessage != nil && callbackMessage().Text == m.Text {
			m.Text += " (double clicked)"
		}
		m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
			[]tgbotapi.InlineKeyboardButton{
				{
					Text:         whc.Translate(trans.COMMAND_TEXT_I_HAVE_CHANGED_MY_MIND),
					CallbackData: fmt.Sprintf("%v?by=%v&%v=%v", SEND_RECEIPT_CALLBACK_PATH, SEND_RECEIPT_BY_CHOOSE_CHANNEL, WIZARD_PARAM_TRANSFER, common.EncodeID(transferID)),
				},
			},
		)
		return m, err
	case string(models.InviteByTelegram):
		panic(fmt.Sprintf("Unsupported option: %v", models.InviteByTelegram))
	case string(models.InviteBySms):

		if counterparty.PhoneNumber > 0 {
			return sendReceiptBySms(whc, counterparty.PhoneContact, transfer, counterparty)
		} else {
			var updateMessage bots.MessageFromBot
			if updateMessage, err = whc.NewEditMessage(whc.Translate(trans.MESSAGE_TEXT_LETS_SEND_SMS), bots.MessageFormatHTML); err != nil {
				return
			}
			if _, err = whc.Responder().SendMessage(c, updateMessage, bots.BotApiSendMessageOverHTTPS); err != nil {
				log.Errorf(c, errors.WithMessage(err, "failed to update Telegram message").Error())
				err = nil
			}

			chatEntity.SetAwaitingReplyTo(ASK_PHONE_NUMBER_FOR_RECEIPT_COMMAND)
			chatEntity.AddWizardParam(WIZARD_PARAM_TRANSFER, strconv.FormatInt(transferID, 10))
			mt := strings.Join([]string{
				whc.Translate(trans.MESSAGE_TEXT_ASK_PHONE_NUMBER_OF_COUNTERPARTY, html.EscapeString(transfer.Counterparty().ContactName)),
				whc.Translate(trans.MESSAGE_TEXT_USE_CONTACT_TO_SEND_PHONE_NUMBER, emoji.PAPERCLIP_ICON),
				whc.Translate(trans.MESSAGE_TEXT_ABOUT_PHONE_NUMBER_FORMAT),
				whc.Translate(trans.MESSAGE_TEXT_THIS_NUMBER_WILL_BE_USED_TO_SEND_RECEIPT),
			}, "\n\n")
			//mt += "\n\n" + whc.Translate(trans.MESSAGE_TEXT_VIEW_MY_NUMBER_IN_INTERNATIONAL_FORMAT)

			m = whc.NewMessage(mt)
			m.Format = bots.MessageFormatHTML
			keyboard := [][]tgbotapi.KeyboardButton{
				[]tgbotapi.KeyboardButton{
					{RequestContact: true, Text: whc.Translate(trans.COMMAND_TEXT_VIEW_MY_NUMBER_IN_INTERNATIONAL_FORMAT)},
				},
			}
			lastName := whc.GetSender().GetLastName()
			if lastName == "Trakhimenok" || lastName == "Paltseva" {
				for k, _ := range common.TwilioTestNumbers {
					keyboard = append(keyboard, []tgbotapi.KeyboardButton{{Text: k}})

				}
			}
			m.Keyboard = &tgbotapi.ReplyKeyboardMarkup{
				Keyboard: keyboard,
			}
		}
	case string(models.InviteByEmail):
		chatEntity.SetAwaitingReplyTo(ASK_EMAIL_FOR_RECEIPT_COMMAND)
		chatEntity.AddWizardParam(WIZARD_PARAM_TRANSFER, strconv.FormatInt(transferID, 10))
		m = whc.NewMessage(whc.Translate(trans.MESSAGE_TEXT_INVITE_ASK_EMAIL_FOR_RECEIPT, transfer.Counterparty().ContactName))
	default:
		err = errors.New("Unknown channel to send receipt: " + sendBy)
		log.Errorf(c, err.Error())
	}
	return m, err
}

func IsTransferNotificationsBlockedForChannel(counterparty *models.ContactEntity, channel string) bool {
	for _, blockedBy := range counterparty.NoTransferUpdatesBy {
		if blockedBy == channel {
			return true
		}
	}
	return false
}
