package facade

import (
	"context"
	"fmt"
	"github.com/dal-go/dalgo/dal"
)

// GetDatabase returns debts tracker database
func GetDatabase(ctx context.Context) (db dal.Database, err error) {
	return nil, fmt.Errorf("TODO: Implement GetDatabase(ctx)")
}

func DB() dal.Database {
	panic("TODO: Implement DB()")
}

func RunReadwriteTransaction(ctx context.Context, f func(ctx context.Context, tx dal.ReadwriteTransaction) error) error {
	db, err := GetDatabase(ctx)
	if err != nil {
		return err
	}
	return db.RunReadwriteTransaction(ctx, f)
}
