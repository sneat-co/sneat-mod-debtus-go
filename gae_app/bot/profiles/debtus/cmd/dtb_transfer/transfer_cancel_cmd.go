package dtb_transfer

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_general"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-framework/core"
)

const CANCEL_TRANSFER_WIZARD_COMMAND = "cancel-transfer-wizard"

var CancelTransferWizardCommand = bots.Command{
	Code:     CANCEL_TRANSFER_WIZARD_COMMAND,
	Commands: trans.Commands(trans.COMMAND_TEXT_CANCEL, "/cancel", emoji.NO_ENTRY_SIGN_ICON),
	Action:   cancelTransferWizardCommandAction,
}

func cancelTransferWizardCommandAction(whc bots.WebhookContext) (bots.MessageFromBot, error) {
	whc.ChatEntity().SetAwaitingReplyTo("")
	var m bots.MessageFromBot
	//userKey, _, err := whc.GetUser()
	//if err != nil {
	//	return m, err
	//}
	//var transfers []models.Transfer
	//ctx := whc.Context()
	//transferKeys, err := datastore.NewQuery(models.TransferKind).Filter("UserID =", userKey.IntID()).Limit(1).GetAll(ctx, &transfers)
	//if err != nil {
	//	return m, err
	//}
	m = whc.NewMessageByCode(trans.MESSAGE_TEXT_TRANSFER_CREATION_CANCELED)
	//if len(transferKeys) == 0 {
	//	m = tgbotapi.NewMessage(whc.ChatID(), Translate(trans.MESSAGE_TEXT_NOTHING_TO_CANCEL, whc))
	//} else {
	//	err := datastore.Delete(ctx, transferKeys[0])
	//	if err != nil {
	//		return m, err
	//	}
	//	//transfer := transfers[0]
	//}
	dtb_general.SetMainMenuKeyboard(whc, &m)
	return m, nil
}
