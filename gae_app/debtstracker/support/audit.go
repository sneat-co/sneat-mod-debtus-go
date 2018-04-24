package support

import (
	"time"

	"github.com/strongo/nds"
	"context"
	"google.golang.org/appengine/datastore"
)

const AuditKind = "Audit"

type AuditEntity struct {
	Action  string
	Created time.Time
	Message string `datastore:",noindex"`
	Related []string
}

type Audit struct {
	ID int64
	AuditEntity
}

func NewAuditEntity(action, message string, related ...string) AuditEntity {
	return AuditEntity{
		Created: time.Now(),
		Action:  action,
		Message: message,
		Related: related,
	}
}

type AuditStorage interface {
	LogAuditRecord(action, message string, related ...string) error
}

type AuditGaeStore struct {
	c context.Context
}

func NewAuditGaeStore(c context.Context) AuditGaeStore {
	return AuditGaeStore{c: c}
}

func (s AuditGaeStore) LogAuditRecord(action, message string, related ...string) (audit Audit, err error) {
	audit.AuditEntity = NewAuditEntity(action, message, related...)
	var key *datastore.Key
	key, err = nds.Put(s.c, datastore.NewIncompleteKey(s.c, AuditKind, nil), audit.AuditEntity)
	if err != nil {
		return
	}
	audit.ID = key.IntID()
	return
}
