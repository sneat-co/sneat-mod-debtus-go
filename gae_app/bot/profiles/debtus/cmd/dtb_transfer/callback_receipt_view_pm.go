package dtb_transfer

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app/log"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"net/url"
	"html"
	"github.com/strongo/bots-framework/platforms/telegram"
)

//func CancelReceiptAction(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
//	return whc.NewMessage("TODO: Sorry, cancel is not implemented yet..."), nil
//}

const VIEW_RECEIPT_CALLBACK_COMMAND = "view-receipt"

var ViewReceiptCallbackCommand = bots.NewCallbackCommand(VIEW_RECEIPT_CALLBACK_COMMAND, viewReceiptCallbackAction)

func ShowReceipt(whc bots.WebhookContext, receiptID int64) (m bots.MessageFromBot, err error) {
	c := whc.Context()

	receipt, err := facade.MarkReceiptAsViewed(c, receiptID, whc.AppUserIntID())
	if err != nil {
		return m, err
	}

	transfer, err := dal.Transfer.GetTransferByID(c, receipt.TransferID)
	if err != nil {
		return m, err
	}

	m = whc.NewMessage("")

	var (
		mt           string
		counterparty models.Contact
	)
	counterpartyCounterparty := transfer.Creator()

	if counterpartyCounterparty.ContactID != 0 {
		counterparty, err = dal.Contact.GetContactByID(c, counterpartyCounterparty.ContactID)
	} else {
		if user, err := dal.User.GetUserByID(c, transfer.CreatorUserID); err != nil {
			return m, err
		} else {
			counterparty.ContactEntity = &models.ContactEntity{}
			counterparty.FirstName = user.FirstName
			counterparty.LastName = user.LastName
		}
	}

	if err != nil {
		return m, err
	}
	utm := common.NewUtmParams(whc, common.UTM_CAMPAIGN_REMINDER)
	mt = common.TextReceiptForTransfer(whc, transfer, whc.AppUserIntID(), common.ShowReceiptToAutodetect, utm)

	log.Debugf(c, "Receipt text: %v", mt)

	var inlineKeyboard *tgbotapi.InlineKeyboardMarkup

	if receipt.CreatorUserID == whc.AppUserIntID() {
		mt += "\n" + whc.Translate(trans.MESSAGE_TEXT_SELF_ACKNOWLEDGEMENT, html.EscapeString(transfer.Counterparty().ContactName))
	} else {
		isAcknowledgedAlready := !transfer.AcknowledgeTime.IsZero()

		if isAcknowledgedAlready {
			switch transfer.AcknowledgeStatus {
			case models.TransferAccepted:
				mt += "\n" + whc.Translate(trans.MESSAGE_TEXT_ALREADY_ACCEPTED_TRANSFER)
			case models.TransferDeclined:
				mt += "\n" + whc.Translate(trans.MESSAGE_TEXT_ALREADY_DECLINED_TRANSFER)
			default:
				log.Errorf(c, "!transfer.AcknowledgeTime.IsZero() && transfer.AcknowledgeStatus not in (accepted, declined)")
			}
		} else {
			mt += "\n" + whc.Translate(trans.MESSAGE_TEXT_PLEASE_ACKNOWLEDGE_TRANSFER)
		}
		receiptCode := common.EncodeID(receiptID)

		if !isAcknowledgedAlready {
			inlineKeyboard = &tgbotapi.InlineKeyboardMarkup{
				InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
					{
						{
							Text:         whc.Translate(trans.COMMAND_TEXT_ACCEPT),
							CallbackData: fmt.Sprintf("%v?id=%v&do=%v", ACKNOWLEDGE_RECEIPT_CALLBACK_COMMAND, receiptCode, dal.AckAccept),
						},
					},
					{
						{
							Text:         whc.Translate(trans.COMMAND_TEXT_DECLINE),
							CallbackData: fmt.Sprintf("%v?id=%v&do=%v", ACKNOWLEDGE_RECEIPT_CALLBACK_COMMAND, receiptCode, dal.AckDecline),
						},
					},
				},
			}
		}
	}

	log.Debugf(c, "mt: %v", mt)
	switch whc.InputType() {
	case bots.WebhookInputCallbackQuery:
		if m, err = whc.NewEditMessage(mt, bots.MessageFormatHTML); err != nil {
			return
		}
		m.DisableWebPagePreview = true
		if inlineKeyboard != nil {
			m.Keyboard = inlineKeyboard
		}
	case bots.WebhookInputText:
		m = whc.NewMessage(mt)
		if inlineKeyboard != nil {
			m.Keyboard = inlineKeyboard
		}
	default:
		if inputType, ok := bots.WebhookInputTypeNames[whc.InputType()]; ok {
			log.Errorf(c, "Unknown input type: %d=%v", whc.InputType(), inputType)
		} else {
			log.Errorf(c, "Unknown input type: %d", whc.InputType())
		}
	}

	if _, err = whc.Responder().SendMessage(c, m, bots.BotApiSendMessageOverHTTPS); err != nil {
		return m, err
	}

	{
		if m, err = whc.NewEditMessage(
			"\xF0\x9F\x93\xA4 "+whc.Translate(trans.MESSAGE_TEXT_RECEIPT_SENT_THROW_TELEGRAM)+"\n\xF0\x9F\x91\x93 "+whc.Translate(trans.MESSAGE_TEXT_RECEIPT_VIEWED_BY_COUNTERPARTY),
			bots.MessageFormatHTML,
		); err != nil {
			return
		}
		m.EditMessageUID = telegram_bot.NewChatMessageUID(transfer.CreatorTgChatID, int(transfer.CreatorTgReceiptByTgMsgID))
		//if _, err := whc.Responder().SendMessage(c, editCreatorMessage, bots.BotApiSendMessageOverHTTPS); err != nil {
		//	log.Errorf(c, "Failed to edit creator message: %v", err)
		//}
	}
	return m, err
}

func viewReceiptCallbackAction(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
	c := whc.Context()

	log.Debugf(c, "ViewReceiptAction(callbackUrl=%v)", callbackUrl)
	callbackQuery := callbackUrl.Query()

	localeCode5 := callbackQuery.Get("locale")
	if localeCode5 != "" {
		whc.SetLocale(localeCode5)
		if appUser, err := whc.GetAppUser(); err != nil {
			return m, err
		} else {
			appUser.SetPreferredLocale(localeCode5)
		}
	}
	receiptID, err := common.DecodeID(callbackQuery.Get("id"))
	if err != nil {
		return m, err
	}
	return ShowReceipt(whc, receiptID)
}

//func (_ viewReceiptCallback) onInvite(whc bots.WebhookContext, inviteCode string) (exit bool, transferID int64, transfer *models.Transfer, m bots.MessageFromBot, err error) {
//	c := whc.Context()
//	var invite *invites.Invite
//	if invite, err = invites.GetInvite(c, inviteCode); err != nil {
//		return
//	} else {
//		if invite == nil {
//			err = errors.New(fmt.Sprintf("Invite not found by code: %v", inviteCode))
//			return
//		}
//		if invite.CreatedByUserID == whc.AppUserIntID() {
//			if transferID, err = invite.RelatedIntID(); err != nil {
//				return
//			}
//			if transfer, err = dal.Transfer.GetTransferByID(c, transferID); err != nil {
//				return
//			}
//			sender := whc.GetSender()
//			mt := getInlineReceiptMessage(whc, true, fmt.Sprintf("%v %v", sender.GetFirstName(), sender.GetLastName()))
//			editedMessage := tgbotapi.NewEditMessageTextByInlineMessageID(
//				whc.InputCallbackQuery().GetInlineMessageID(),
//				mt+"\n\n"+whc.Translate(trans.MESSAGE_TEXT_FOR_COUNTERPARTY_ONLY, transfer.Contact().ContactName),
//			)
//			editedMessage.ParseMode = "HTML"
//			editedMessage.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
//				InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
//					transferReceiptChooseLocaleButtons(inviteCode, invite.CreatedOnID, invite.CreatedOnPlatform),
//				},
//			}
//			m.TelegramEditMessageText = &editedMessage
//			exit = true
//			return
//		}
//
//		if transferID, transfer, _, _, err = ClaimInviteOnTransfer(whc, whc.InputCallbackQuery().GetInlineMessageID(), inviteCode, invite); err != nil {
//			return
//		}
//	}
//	return
//}
