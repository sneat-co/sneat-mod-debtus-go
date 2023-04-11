package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/db/gaedb"
	"google.golang.org/appengine/v2/datastore"
)

func NewLoginPinIncompleteKey(c context.Context) *datastore.Key {
	return datastore.NewIncompleteKey(c, models.LoginPinKind, nil)
}

func NewLoginPinKey(c context.Context, id int64) *datastore.Key {
	return gaedb.NewKey(c, models.LoginPinKind, "", id, nil)
}
