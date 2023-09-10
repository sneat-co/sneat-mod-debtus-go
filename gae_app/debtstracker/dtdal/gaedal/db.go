package gaedal

import (
	"context"
	"errors"
	"github.com/dal-go/dalgo/dal"
)

func GetDatabase(_ context.Context) (db dal.DB, err error) {
	return nil, errors.New("TODO: implement me: GetDatabase()")
}
