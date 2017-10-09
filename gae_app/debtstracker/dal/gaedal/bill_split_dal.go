package gaedal

import (
	"golang.org/x/net/context"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
	"bitbucket.com/debtstracker/gae_app/debtstracker/dal"
)

type splitDalGae struct {
}

var _ dal.SplitDal = (*splitDalGae)(nil) // Make sure we implement interface

func (splitDalGae) GetSplitByID(c context.Context, splitID int64) (split models.Split, err error) {
	split.ID = splitID
	err = dal.DB.Get(c, &split)
	return
}

func (splitDalGae) InsertSplit(c context.Context, splitEntity models.SplitEntity) (split models.Split, err error) {
	split.SplitEntity = &splitEntity
	if err = dal.DB.Update(c, &split); err != nil {
		return
	}
	return
}
