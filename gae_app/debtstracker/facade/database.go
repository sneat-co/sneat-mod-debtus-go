package facade

import (
	"context"
	"fmt"
	"github.com/strongo/dalgo/dal"
)

// GetDatabase returns debts tracker database
func GetDatabase(ctx context.Context) (db dal.Database, err error) {
	return nil, fmt.Errorf("TODO: Implement GetDatabase(ctx)")
}

func RunReadwriteTransaction(ctx context.Context, f func(ctx context.Context, tx dal.ReadwriteTransaction) error) error {
	db, err := GetDatabase(ctx)
	if err != nil {
		return err
	}
	return db.RunReadwriteTransaction(ctx, f)
}

func RunReadonlyTransaction(ctx context.Context, f func(ctx context.Context, tx dal.ReadTransaction) error) error {
	db := GetDatabase(ctx)
	return db.RunReadonlyTransaction(ctx, f)
}
