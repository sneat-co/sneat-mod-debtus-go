package delayer4debtus

import "github.com/strongo/delaying"

var (
	FixTransfersIsOutstanding,
	CreateAndSendReceiptToCounterpartyByTelegram,
	DeleteContactTransfersDelayFunc,
	SetReminderIsSent,
	DiscardReminder,
	DiscardReminders,
	DiscardRemindersForTransfer,
	UpdateTransferWithCreatorReceiptTgMessageID,
	SendReceiptToCounterpartyByTelegram,
	UpdateInviteClaimedCount,
	UpdateTransfersWithCounterparty,
	UpdateTransferWithCounterparty,
	MarkReceiptAsSent,
	UpdateTransfersWithCreatorName,
	UpdateTransfersOnReturn,
	UpdateTransferOnReturn,
	OnReceiptSendFail,
	CreateReminderForTransferUser,
	OnReceiptSentSuccess delaying.Delayer
)
