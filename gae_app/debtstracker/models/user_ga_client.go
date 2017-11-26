package models

import (
	"time"

	"github.com/strongo/db"
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

func (GaClient) NewEntity() interface{} {
	return new(GaClientEntity)
}

func (gaClient *GaClient) SetEntity(entity interface{}) {
	if entity == nil {
		gaClient.GaClientEntity = nil

	} else {
		gaClient.GaClientEntity = entity.(*GaClientEntity)

	}
}

var _ db.EntityHolder = (*GaClient)(nil)
