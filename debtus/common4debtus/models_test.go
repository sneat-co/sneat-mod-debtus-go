package common4debtus

import (
	"context"
	"github.com/sneat-co/sneat-go-core/utm"
	"github.com/strongo/i18n"
	"regexp"
	"testing"
)

func TestGetCounterpartyUrl(t *testing.T) {
	var (
		utm utm.Params
	)
	counterpartyUrl, _ := GetCounterpartyUrl(context.Background(), "123", "", i18n.LocaleRuRu, utm)

	re := regexp.MustCompile(`^https://debtusbot\.io/counterparty\?id=\d+&lang=\w{2}$`)
	if !re.MatchString(counterpartyUrl) {
		t.Errorf("Unexpected counterpart URL:\n%v", counterpartyUrl)
	}
}
