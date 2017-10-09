package gaedal

import (
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"github.com/strongo/app/gaedb"
)

func NewLoginPinIncompleteKey(c context.Context) *datastore.Key {
	return datastore.NewIncompleteKey(c, models.LoginPinKind, nil)
}

func NewLoginPinKey(c context.Context, id int64) *datastore.Key {
	return gaedb.NewKey(c, models.LoginPinKind, "", id, nil)
}
