package common

import (
	"regexp"
	"testing"

	"github.com/strongo/app"
)

func TestGetCounterpartyUrl(t *testing.T) {
	var (
		utm UtmParams
	)
	counterpartyUrl := GetCounterpartyUrl(123, 0, strongo.LocaleRuRu, utm)

	re := regexp.MustCompile(`^https://debtstracker\.io/counterparty\?id=\d+&lang=\w{2}$`)
	if !re.MatchString(counterpartyUrl) {
		t.Errorf("Unexpected counterpart URL:\n%v", counterpartyUrl)
	}
}
