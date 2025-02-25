package dal4debtus

import (
	"context"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/models4debtus"
)

func GetDebtusUser(ctx context.Context, tx dal.ReadSession, debtusUser models4debtus.DebtusUserEntry) (err error) {
	return tx.Get(ctx, debtusUser.Record)
}
