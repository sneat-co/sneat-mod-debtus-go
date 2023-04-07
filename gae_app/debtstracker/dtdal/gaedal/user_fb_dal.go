package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/db/gaedb"
	"google.golang.org/appengine/datastore"
)

func NewUserFacebookKey(c context.Context, fbAppOrPageID, fbUserOrPageScopeID string) *datastore.Key {
	if fbAppOrPageID == "" {
		panic("fbAppOrPageID is empty string")
	}
	if fbUserOrPageScopeID == "" {
		panic("fbUserOrPageScopeID is empty string")
	}
	return gaedb.NewKey(c, models.UserFacebookKind, fbAppOrPageID+":"+fbUserOrPageScopeID, 0, nil)
}

type UserFacebookDalGae struct {
}

func NewUserFacebookDalGae() UserFacebookDalGae {
	return UserFacebookDalGae{}
}

func (UserFacebookDalGae) SaveFbUser(c context.Context, fbUser models.UserFacebook) (err error) {
	key := NewUserFacebookKey(c, fbUser.FbAppOrPageID, fbUser.FbUserOrPageScopeID)
	if _, err = gaedb.Put(c, key, fbUser.UserFacebookEntity); err != nil {
		return
	}
	return
}

func (UserFacebookDalGae) DeleteFbUser(c context.Context, fbAppOrPageID, fbUserOrPageScopeID string) (err error) {
	key := NewUserFacebookKey(c, fbAppOrPageID, fbUserOrPageScopeID)
	if err = gaedb.Delete(c, key); err != nil {
		return
	}
	return
}

func (UserFacebookDalGae) GetFbUserByFbID(c context.Context, fbAppOrPageID, fbUserOrPageScopeID string) (fbUser models.UserFacebook, err error) {
	var entity models.UserFacebookEntity
	if err = gaedb.Get(c, NewUserFacebookKey(c, fbAppOrPageID, fbUserOrPageScopeID), &entity); err != nil {
		if err == datastore.ErrNoSuchEntity {
			err = db.NewErrNotFoundByStrID(models.UserFacebookKind, fbUserOrPageScopeID, err)
		}
		return
	}
	fbUser = models.UserFacebook{
		FbAppOrPageID:       fbAppOrPageID,
		FbUserOrPageScopeID: fbUserOrPageScopeID,
		UserFacebookEntity:  &entity,
	}
	return
}
