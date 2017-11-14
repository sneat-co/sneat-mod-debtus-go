package models

import (
	"github.com/strongo/app/db"
	"time"
)

const GaClientKind = "UserGaClient"

type GaClientEntity struct {
	Created   time.Time
	UserAgent string `datastore:",noindex"`
	IpAddress string `datastore:",noindex"`
}

type GaClient struct {
	db.StringID
	*GaClientEntity
}

func (GaClient) Kind() string {
	return GaClientKind
}

func (gaClient GaClient) Entity() interface{} {
	return gaClient.GaClientEntity
}

func (gaClient *GaClient) SetEntity(entity interface{}) {
	gaClient.GaClientEntity = entity.(*GaClientEntity)
}


var _ db.EntityHolder = (*GaClient)(nil)