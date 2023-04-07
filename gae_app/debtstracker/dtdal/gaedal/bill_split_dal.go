package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
)

type splitDalGae struct {
}

var _ dtdal.SplitDal = (*splitDalGae)(nil) // Make sure we implement interface

func (splitDalGae) GetSplitByID(c context.Context, splitID int64) (split models.Split, err error) {
	split.ID = splitID
	err = dtdal.DB.Get(c, &split)
	return
}

func (splitDalGae) InsertSplit(c context.Context, splitEntity models.SplitEntity) (split models.Split, err error) {
	split.SplitEntity = &splitEntity
	if err = dtdal.DB.Update(c, &split); err != nil {
		return
	}
	return
}
