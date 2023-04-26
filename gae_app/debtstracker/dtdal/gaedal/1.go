package gaedal

import "github.com/strongo/app/delaying"

func InitDelaying(mustRegisterFunc func(key string, i any) delaying.Function) {
	delayUpdateBillDependencies = mustRegisterFunc("delayUpdateBillDependencies", delayedUpdateBillDependencies)
	delayUpdateInviteClaimedCount = mustRegisterFunc("UpdateInviteClaimedCount", delayedUpdateInviteClaimedCount)
	delayedUpdateUserWithContact = mustRegisterFunc("updateUserWithContact", updateUserWithContact)
	delayUpdateUserWithBill = mustRegisterFunc(updateUserWithBillKeyName, delayedUpdateUserWithBill)
	delayUpdateTransfersWithCounterparty = mustRegisterFunc(DELAY_UPDATE_TRANSFERS_WITH_COUNTERPARTY, delayedUpdateTransfersWithCounterparty)
	delayUpdateTransferWithCounterparty = mustRegisterFunc(DELAY_UPDATE_1_TRANSFER_WITH_COUNTERPARTY, delayedUpdateTransferWithCounterparty)
	delayMarkReceiptAsSent = mustRegisterFunc("delayMarkReceiptAsSent", delayedMarkReceiptAsSent)
	delayUpdateTransfersWithCreatorName = mustRegisterFunc(UPDATE_TRANSFERS_WITH_CREATOR_NAME, delayedUpdateTransfersWithCreatorName)
	delayUpdateTransfersOnReturn = mustRegisterFunc("updateTransfersOnReturn", updateTransfersOnReturn)
	delayUpdateTransferOnReturn = mustRegisterFunc("updateTransferOnReturn", updateTransferOnReturn)
	delaySetUserPreferredLocale = mustRegisterFunc("delayedSetUserPreferredLocale", delayedSetUserPreferredLocale)
	delayedOnReceiptSentSuccess = mustRegisterFunc("onReceiptSentSuccess", onReceiptSentSuccess)
	delayedOnReceiptSendFail = mustRegisterFunc("onReceiptSendFail", onReceiptSendFail)
	delayCreateReminderForTransferUser = mustRegisterFunc("delayedCreateReminderForTransferUser", delayedCreateReminderForTransferUser)
	delayedSendReceiptToCounterpartyByTelegram = mustRegisterFunc("delayedSendReceiptToCounterpartyByTelegram", sendReceiptToCounterpartyByTelegram)
	delayUpdateTransferWithCreatorReceiptTgMessageID = mustRegisterFunc("UpdateTransferWithCreatorReceiptTgMessageID", delayedUpdateTransferWithCreatorReceiptTgMessageID)
	delayDiscardRemindersForTransfer = mustRegisterFunc("discardRemindersForTransfer", discardRemindersForTransfer)
	delayUpdateUserHasDueTransfers = mustRegisterFunc("delayUpdateUserHasDueTransfers", delayedUpdateUserHasDueTransfers)
	delayUpdateUsersWithBill = mustRegisterFunc(updateUsersWithBillKeyName, updateUsersWithBill)
	delayUpdateGroupWithBill = mustRegisterFunc("delayedUpdateWithBill", delayedUpdateGroupWithBill)
	delayDiscardReminders = mustRegisterFunc("discardReminders", discardReminders)
	delaySetReminderIsSent = mustRegisterFunc("setReminderIsSent", setReminderIsSent)
	delayDiscardReminder = mustRegisterFunc("DiscardReminder", delayedDiscardReminder)
	delayDeleteContactTransfersDelayFunc = mustRegisterFunc(DeleteContactTransfersFuncKey, delayedDeleteContactTransfers) // TODO: Duplicate of delayDeleteContactTransfers ?
	delayCreateAndSendReceiptToCounterpartyByTelegram = mustRegisterFunc("delayCreateAndSendReceiptToCounterpartyByTelegram", delayedCreateAndSendReceiptToCounterpartyByTelegram)
	delayFixTransfersIsOutstanding = mustRegisterFunc("fix-transfers-is-outstanding", fixTransfersIsOutstanding)
}

var (
	delayFixTransfersIsOutstanding,
	delayCreateAndSendReceiptToCounterpartyByTelegram,
	delayDeleteContactTransfersDelayFunc,
	delayDiscardReminder,
	delaySetReminderIsSent,
	delayDiscardReminders,
	delayUpdateGroupWithBill,
	delayUpdateUsersWithBill,
	delayUpdateUserHasDueTransfers,
	delayDiscardRemindersForTransfer,
	delayUpdateTransferWithCreatorReceiptTgMessageID,
	delayedSendReceiptToCounterpartyByTelegram,
	delayedUpdateUserWithContact,
	delayUpdateInviteClaimedCount,
	delayUpdateBillDependencies,
	delayUpdateUserWithBill,
	delayUpdateTransfersWithCounterparty,
	delayUpdateTransferWithCounterparty,
	delayMarkReceiptAsSent,
	delayUpdateTransfersWithCreatorName,
	delayUpdateTransfersOnReturn,
	delayUpdateTransferOnReturn,
	delaySetUserPreferredLocale,
	delayedOnReceiptSendFail,
	delayCreateReminderForTransferUser,
	delayedOnReceiptSentSuccess delaying.Function
)
