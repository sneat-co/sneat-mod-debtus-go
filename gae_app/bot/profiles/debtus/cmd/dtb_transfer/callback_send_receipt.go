package dtb_transfer

import (
	"fmt"
	"github.com/bots-go-framework/bots-api-telegram/tgbotapi"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/sneat-co/debtstracker-translations/trans"
	"html"
	"net/url"
	"strconv"
	"strings"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bitbucket.org/asterus/debtstracker-server/gae_app/general"
	"errors"
	"github.com/bots-go-framework/bots-fw-telegram"
	"github.com/sneat-co/debtstracker-translations/emoji"
	"github.com/strongo/log"
)

var SendReceiptCallbackCommand = botsfw.NewCallbackCommand(SEND_RECEIPT_CALLBACK_PATH, CallbackSendReceipt)

func CallbackSendReceipt(whc botsfw.WebhookContext, callbackUrl *url.URL) (m botsfw.MessageFromBot, err error) {
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
		return m, fmt.Errorf("failed to decode transferID to int: %w", err)
	}
	transfer, err = facade.Transfers.GetTransferByID(c, tx, transferID)
	if err != nil {
		return m, fmt.Errorf("failed to get transfer by ID: %w", err)
	}
	//chatEntity := whc.ChatEntity() //TODO: Need this to get appUser, has to be refactored
	//appUser, err := whc.GetAppUser()
	counterparty, err := facade.GetContactByID(c, transfer.Data.Counterparty().ContactID)
	if err != nil {
		return m, err
	}
	if IsTransferNotificationsBlockedForChannel(counterparty.Data, sendBy) {
		m = whc.NewMessage(trans.MESSAGE_TEXT_USER_BLOCKED_TRANSFER_NOTIFICATIONS_BY)
		return m, err
	}
	chatEntity := whc.ChatEntity()
	switch sendBy {
	case SEND_RECEIPT_BY_CHOOSE_CHANNEL:
		return createSendReceiptOptionsMessage(whc, transfer)
	case RECEIPT_ACTION__DO_NOT_SEND:
		log.Debugf(c, "CallbackSendReceipt(): do-not-send")
		if m, err = whc.NewEditMessage(whc.Translate(trans.MESSAGE_TEXT_RECEIPT_WILL_NOT_BE_SENT), botsfw.MessageFormatHTML); err != nil {
			return
		}

		// TODO: do type assertion with botsfw.CallbackQuery interface
		if callbackMessage := whc.Input().(telegram.TgWebhookCallbackQuery).TelegramCallbackMessage; callbackMessage != nil && callbackMessage().Text == m.Text {
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
	case string(models.InviteByLinkToTelegram):
		return showLinkForReceiptInTelegram(whc, transfer)
	case string(models.InviteBySms):

		if counterparty.Data.PhoneNumber > 0 {
			return sendReceiptBySms(whc, counterparty.Data.PhoneContact, transfer, counterparty)
		} else {
			var updateMessage botsfw.MessageFromBot
			if updateMessage, err = whc.NewEditMessage(whc.Translate(trans.MESSAGE_TEXT_LETS_SEND_SMS), botsfw.MessageFormatHTML); err != nil {
				return
			}
			if _, err = whc.Responder().SendMessage(c, updateMessage, botsfw.BotAPISendMessageOverHTTPS); err != nil {
				log.Errorf(c, fmt.Errorf("failed to update Telegram message: %w", err).Error())
				err = nil
			}

			chatEntity.SetAwaitingReplyTo(ASK_PHONE_NUMBER_FOR_RECEIPT_COMMAND)
			chatEntity.AddWizardParam(WIZARD_PARAM_TRANSFER, strconv.FormatInt(transferID, 10))
			mt := strings.Join([]string{
				whc.Translate(trans.MESSAGE_TEXT_ASK_PHONE_NUMBER_OF_COUNTERPARTY, html.EscapeString(transfer.Data.Counterparty().ContactName)),
				whc.Translate(trans.MESSAGE_TEXT_USE_CONTACT_TO_SEND_PHONE_NUMBER, emoji.PAPERCLIP_ICON),
				whc.Translate(trans.MESSAGE_TEXT_ABOUT_PHONE_NUMBER_FORMAT),
				whc.Translate(trans.MESSAGE_TEXT_THIS_NUMBER_WILL_BE_USED_TO_SEND_RECEIPT),
			}, "\n\n")
			//mt += "\n\n" + whc.Translate(trans.MESSAGE_TEXT_VIEW_MY_NUMBER_IN_INTERNATIONAL_FORMAT)

			m = whc.NewMessage(mt)
			m.Format = botsfw.MessageFormatHTML
			keyboard := [][]tgbotapi.KeyboardButton{
				{
					{RequestContact: true, Text: whc.Translate(trans.COMMAND_TEXT_VIEW_MY_NUMBER_IN_INTERNATIONAL_FORMAT)},
				},
			}
			lastName := whc.GetSender().GetLastName()
			if lastName == "Trakhimenok" || lastName == "Paltseva" {
				for k := range common.TwilioTestNumbers {
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
		m = whc.NewMessage(whc.Translate(trans.MESSAGE_TEXT_INVITE_ASK_EMAIL_FOR_RECEIPT, transfer.Data.Counterparty().ContactName))
	default:
		err = errors.New("Unknown channel to send receipt: " + sendBy)
		log.Errorf(c, err.Error())
	}
	return m, err
}

func showLinkForReceiptInTelegram(whc botsfw.WebhookContext, transfer models.Transfer) (m botsfw.MessageFromBot, err error) {
	receipt := models.NewReceiptEntity(whc.AppUserIntID(), transfer.ID, transfer.Data.Counterparty().UserID, whc.Locale().Code5, "link", "telegram", general.CreatedOn{
		CreatedOnPlatform: whc.BotPlatform().ID(),
		CreatedOnID:       whc.GetBotCode(),
	})
	var receiptID int64
	if receiptID, err = dtdal.Receipt.CreateReceipt(whc.Context(), &receipt); err != nil {
		return m, err
	}
	receiptUrl := GetUrlForReceiptInTelegram(whc.GetBotCode(), receiptID, whc.Locale().Code5)
	m.Text = "Send this link to counterparty:\n\n" + fmt.Sprintf(`<a href="%v">%v</a>`, receiptUrl, receiptUrl) + "\n\nPlease be aware that the first person opening this link will be treated as counterparty for this debt."
	m.Format = botsfw.MessageFormatHTML
	m.IsEdit = true
	return
}

func IsTransferNotificationsBlockedForChannel(counterparty *models.ContactEntity, channel string) bool {
	for _, blockedBy := range counterparty.NoTransferUpdatesBy {
		if blockedBy == channel {
			return true
		}
	}
	return false
}
