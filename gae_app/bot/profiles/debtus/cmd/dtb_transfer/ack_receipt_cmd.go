package dtb_transfer

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_general"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/log"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/measurement-protocol"
	"html"
)

func AcknowledgeReceipt(whc bots.WebhookContext, receiptID int64, operation string) (m bots.MessageFromBot, err error) {
	c := whc.Context()

	_, transfer, isCounterpartiesJustConnected, err := facade.AcknowledgeReceipt(c, receiptID, whc.AppUserIntID(), operation)
	if err != nil {
		if errors.Cause(err) == facade.ErrSelfAcknowledgement {
			m = whc.NewMessage(whc.Translate(trans.MESSAGE_TEXT_SELF_ACKNOWLEDGEMENT, html.EscapeString(transfer.Counterparty().ContactName)))
			return m, nil
		}
		return m, err
	} else {

		{ // Reporting to Google Analytics
			gaMeasurement := whc.GaMeasurement()

			gaMeasurement.Queue(measurement.NewEventWithLabel(
				"receipts",
				"receipt-acknowledged",
				operation,
				whc.GaCommon(),
			))

			if isCounterpartiesJustConnected {
				gaMeasurement.Queue(measurement.NewEvent(
					"counterparties",
					"counterparties-connected",
					whc.GaCommon(),
				))
			}
		}

		var operationMessage string
		switch operation {
		case dal.AckAccept:
			operationMessage = whc.Translate(trans.MESSAGE_TEXT_TRANSFER_ACCEPTED_BY_YOU)
		case dal.AckDecline:
			operationMessage = whc.Translate(trans.MESSAGE_TEXT_TRANSFER_DECLINED_BY_YOU)
		default:
			err = errors.New("Expected accept or decline as operation, got: " + operation)
			return
		}

		utm := common.NewUtmParams(whc, common.UTM_CAMPAIGN_RECEIPT)
		if whc.InputType() == bots.WebhookInputCallbackQuery {
			if m, err = whc.NewEditMessage(common.TextReceiptForTransfer(whc, transfer, 0, common.ShowReceiptToCounterparty, utm)+"\n\n"+operationMessage, bots.MessageFormatHTML); err != nil {
				return
			}
		} else {
			m = whc.NewMessage(operationMessage + "\n\n" + common.TextReceiptForTransfer(whc, transfer, 0, common.ShowReceiptToCounterparty, utm))
			m.Keyboard = dtb_general.MainMenuKeyboardOnReceiptAck(whc)
			m.Format = bots.MessageFormatHTML
		}

		if transfer.Creator().TgChatID != 0 {
			askMsgToCreator := whc.NewMessage("")
			askMsgToCreator.ToChat = bots.ChatIntID(transfer.Creator().TgChatID)
			var operationMsg string
			counterpartyName := transfer.Counterparty().ContactName
			switch operation {
			case "accept":
				operationMsg = whc.Translate(trans.MESSAGE_TEXT_TRANSFER_ACCEPTED_BY_COUNTERPARTY, html.EscapeString(counterpartyName))
			case "decline":
				operationMsg = whc.Translate(trans.MESSAGE_TEXT_TRANSFER_DECLINED_BY_COUNTERPARTY, html.EscapeString(counterpartyName))
			default:
				err = errors.New("Expected accept or decline as operation, got: " + operation)
			}
			askMsgToCreator.Text = operationMsg + "\n\n" + common.TextReceiptForTransfer(whc, transfer, transfer.CreatorUserID, common.ShowReceiptToAutodetect, utm)

			if transfer.Creator().TgBotID != whc.GetBotCode() {
				log.Warningf(c, "TODO: transferEntity.Creator().TgBotID != whc.GetBotCode(): "+askMsgToCreator.Text)
			} else {
				if _, err = whc.Responder().SendMessage(c, askMsgToCreator, bots.BotApiSendMessageOverHTTPS); err != nil {
					log.Errorf(c, "Failed to send acknowledge to creator: %v", err)
					err = nil // This is not that critical to report the error to user
				}
			}
		}
		// Seems we can edit message just once after callback :(
		//if transferEntity.CounterpartyTgReceiptInlineMessageID != "" {
		//	mt = common.TextReceiptForTransfer(whc, transferID, transferEntity, transferEntity.CounterpartyCounterpartyID)
		//	editMessage := tgbotapi.NewEditMessageTextByInlineMessageID(transferEntity.CounterpartyTgReceiptInlineMessageID, mt + fmt.Sprintf("\n\n Acknowledged by %v", transferEntity.Contact().ContactName))
		//
		//	if values, err := editMessage.Values(); err != nil {
		//		log.Errorf(c, "Failed to get values for editMessage: %v", err)
		//	} else {
		//		log.Debugf(c, "editMessage.Values(): %v", values)
		//	}
		//	updateMessage := whc.NewMessage("")
		//	updateMessage.TelegramEditMessageText = &editMessage
		//	_, err := whc.Responder().SendMessage(c, updateMessage, bots.BotApiSendMessageOverHTTPS)
		//	if err != nil {
		//		log.Errorf(c, "Failed to update counterparty receipt message: %v", err)
		//	}
		//}
		return m, err
	}
}
