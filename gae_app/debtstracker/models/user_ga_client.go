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
	db.NoIntID
	ID string
	*GaClientEntity
}
