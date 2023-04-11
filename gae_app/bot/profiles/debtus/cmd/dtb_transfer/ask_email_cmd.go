package dtb_transfer

import (
	"strconv"
	"strings"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bitbucket.org/asterus/debtstracker-server/gae_app/general"
	"bitbucket.org/asterus/debtstracker-server/gae_app/invites"
	"github.com/strongo/log"
)

const ASK_EMAIL_FOR_RECEIPT_COMMAND = "ask-email-for-receipt"

var AskEmailForReceiptCommand = botsfw.Command{
	Code: ASK_EMAIL_FOR_RECEIPT_COMMAND,
	Action: func(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error) {
		c := whc.Context()

		log.Debugf(c, "AskEmailForReceiptCommand.Action()")
		email := whc.Input().(botsfw.WebhookTextMessage).Text()
		if strings.Index(email, "@") < 0 {
			return whc.NewMessage(whc.Translate(trans.MESSAGE_TEXT_INVALID_EMAIL)), nil
		}

		chatEntity := whc.ChatEntity()
		var transferID int64
		transferID, err = strconv.ParseInt(chatEntity.GetWizardParam(WIZARD_PARAM_TRANSFER), 10, 64)
		if err != nil {
			return m, err
		}
		transfer, err := facade.Transfers.GetTransferByID(c, transferID)
		if err != nil {
			return m, err
		}
		m, err = sendReceiptByEmail(whc, email, "", transfer)
		return
	},
}

func sendReceiptByEmail(whc botsfw.WebhookContext, toEmail, toName string, transfer models.Transfer) (m botsfw.MessageFromBot, err error) {
	c := whc.Context()
	receiptEntity := models.NewReceiptEntity(whc.AppUserIntID(), transfer.ID, transfer.Counterparty().UserID, whc.Locale().Code5, string(models.InviteByEmail), toEmail, general.CreatedOn{
		CreatedOnPlatform: whc.BotPlatform().ID(),
		CreatedOnID:       whc.GetBotCode(),
	})
	receiptID, err := dtdal.Receipt.CreateReceipt(c, &receiptEntity)

	emailID := ""
	if emailID, err = invites.SendReceiptByEmail(
		whc.ExecutionContext(),
		models.Receipt{IntegerID: db.NewIntID(receiptID), ReceiptEntity: &receiptEntity},
		whc.GetSender().GetFirstName(),
		toName,
		toEmail,
	); err != nil {
		return m, err
	}

	m = whc.NewMessageByCode(trans.MESSAGE_TEXT_RECEIPT_SENT_THROW_EMAIL, emailID)

	return m, err
}
