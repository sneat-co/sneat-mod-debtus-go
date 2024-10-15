package gaedal

import (
	"github.com/strongo/delaying"
)

func InitDelaying(mustRegisterFunc func(key string, i any) delaying.Delayer) {

	delayerUpdateInviteClaimedCount = mustRegisterFunc("UpdateInviteClaimedCount", delayedUpdateInviteClaimedCount)
	delayerUpdateTransfersWithCounterparty = mustRegisterFunc(DELAY_UPDATE_TRANSFERS_WITH_COUNTERPARTY, delayedUpdateTransfersWithCounterparty)
	delayerUpdateTransferWithCounterparty = mustRegisterFunc(DELAY_UPDATE_1_TRANSFER_WITH_COUNTERPARTY, delayedUpdateTransferWithCounterparty)
	delayerMarkReceiptAsSent = mustRegisterFunc("delayerMarkReceiptAsSent", delayedMarkReceiptAsSent)
	delayerUpdateTransfersWithCreatorName = mustRegisterFunc(UPDATE_TRANSFERS_WITH_CREATOR_NAME, delayedUpdateTransfersWithCreatorName)
	delayerUpdateTransfersOnReturn = mustRegisterFunc("updateTransfersOnReturn", updateTransfersOnReturn)
	delayerUpdateTransferOnReturn = mustRegisterFunc("updateTransferOnReturn", updateTransferOnReturn)
	delayerOnReceiptSentSuccess = mustRegisterFunc("onReceiptSentSuccess", onReceiptSentSuccess)
	delayerOnReceiptSendFail = mustRegisterFunc("delayedOnReceiptSendFail", delayedOnReceiptSendFail)
	delayerCreateReminderForTransferUser = mustRegisterFunc("delayedCreateReminderForTransferUser", delayedCreateReminderForTransferUser)
	delayerSendReceiptToCounterpartyByTelegram = mustRegisterFunc("delayedSendReceiptToCounterpartyByTelegram", delayedSendReceiptToCounterpartyByTelegram)
	delayerUpdateTransferWithCreatorReceiptTgMessageID = mustRegisterFunc("delayedUpdateTransferWithCreatorReceiptTgMessageID", delayedUpdateTransferWithCreatorReceiptTgMessageID)
	delayerDiscardRemindersForTransfer = mustRegisterFunc("delayedDiscardRemindersForTransfer", delayedDiscardRemindersForTransfer)
	delayerDiscardReminders = mustRegisterFunc("delayedDiscardReminders", delayedDiscardReminders)
	delayerSetReminderIsSent = mustRegisterFunc("delayedSetReminderIsSent", delayedSetReminderIsSent)
	delayerDiscardReminder = mustRegisterFunc("delayedDiscardReminder", delayedDiscardReminder)
	delayerDeleteContactTransfersDelayFunc = mustRegisterFunc(DeleteContactTransfersFuncKey, delayedDeleteContactTransfers) // TODO: Duplicate of delayDeleteContactTransfers ?
	delayerCreateAndSendReceiptToCounterpartyByTelegram = mustRegisterFunc("delayedCreateAndSendReceiptToCounterpartyByTelegram", delayedCreateAndSendReceiptToCounterpartyByTelegram)
	delayerFixTransfersIsOutstanding = mustRegisterFunc("delayedFixTransfersIsOutstanding", delayedFixTransfersIsOutstanding)
}

var (
	delayerFixTransfersIsOutstanding,
	delayerCreateAndSendReceiptToCounterpartyByTelegram,
	delayerDeleteContactTransfersDelayFunc,
	delayerDiscardReminder,
	delayerSetReminderIsSent,
	delayerDiscardReminders,

	delayerDiscardRemindersForTransfer,
	delayerUpdateTransferWithCreatorReceiptTgMessageID,
	delayerSendReceiptToCounterpartyByTelegram,
	delayerUpdateInviteClaimedCount,
	delayerUpdateTransfersWithCounterparty,
	delayerUpdateTransferWithCounterparty,
	delayerMarkReceiptAsSent,
	delayerUpdateTransfersWithCreatorName,
	delayerUpdateTransfersOnReturn,
	delayerUpdateTransferOnReturn,
	delayerOnReceiptSendFail,
	delayerCreateReminderForTransferUser,
	delayerOnReceiptSentSuccess delaying.Delayer
)
