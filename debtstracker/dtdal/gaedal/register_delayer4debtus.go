package gaedal

import (
	"github.com/sneat-co/sneat-mod-debtus-go/debtstracker/dtdal/delayer4debtus"
	"github.com/strongo/delaying"
)

func RegisterDelayers4Debtus() {
	delayer4debtus.DeleteContactTransfersDelayFunc = delaying.MustRegisterFunc(DeleteContactTransfersFuncKey, delayedDeleteContactTransfers)
	delayer4debtus.UpdateInviteClaimedCount = delaying.MustRegisterFunc("UpdateInviteClaimedCount", delayedUpdateInviteClaimedCount)
	delayer4debtus.MarkReceiptAsSent = delaying.MustRegisterFunc("MarkReceiptAsSent", delayedMarkReceiptAsSent)
	delayer4debtus.FixTransfersIsOutstanding = delaying.MustRegisterFunc("FixTransfersIsOutstanding", delayedFixTransfersIsOutstanding)
	delayer4debtus.UpdateTransferOnReturn = delaying.MustRegisterFunc("UpdateTransferOnReturn", delayedUpdateTransfersOnReturn)
	delayer4debtus.UpdateTransfersWithCounterparty = delaying.MustRegisterFunc("UpdateTransfersWithCounterparty", delayedUpdateTransfersWithCounterparty)
	delayer4debtus.UpdateTransferWithCounterparty = delaying.MustRegisterFunc("UpdateTransferWithCounterparty", delayedUpdateTransferWithCounterparty)
	delayer4debtus.UpdateTransfersWithCreatorName = delaying.MustRegisterFunc("UpdateTransfersWithCreatorName", delayedUpdateTransfersWithCreatorName)
	delayer4debtus.UpdateTransfersOnReturn = delaying.MustRegisterFunc("UpdateTransfersOnReturn", delayedUpdateTransfersOnReturn)
	delayer4debtus.UpdateTransferOnReturn = delaying.MustRegisterFunc("UpdateTransferOnReturn", delayedUpdateTransferOnReturn)
}
