package facade

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo2firestore"
)

// GetDatabase returns debts tracker database
func GetDatabase(ctx context.Context) (db dal.Database, err error) {
	client, err := firestore.NewClient(ctx, "projectID")
	if err != nil {
		panic(err)
	}
	return dalgo2firestore.NewDatabase("sneat", client), nil
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
