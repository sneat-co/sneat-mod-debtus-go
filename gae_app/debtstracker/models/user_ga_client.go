package models

import (
	"time"
	"github.com/strongo/app/db"
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
