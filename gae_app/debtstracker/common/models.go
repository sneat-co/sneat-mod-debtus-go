package common

import (
	"bytes"
	"fmt"
	"io"

	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/auth"
	"github.com/strongo/app"
)

func GetCounterpartyUrl(counterpartyID int64, currentUserID int64, locale strongo.Locale, utmParams UtmParams) string {
	var buffer bytes.Buffer
	WriteCounterpartyUrl(&buffer, counterpartyID, currentUserID, locale, utmParams)
	return buffer.String()
}

func WriteCounterpartyUrl(writer io.Writer, counterpartyID int64, currentUserID int64, locale strongo.Locale, utmParams UtmParams) {
	host := GetWebsiteHost(utmParams.Source)
	_, _ = writer.Write([]byte(fmt.Sprintf("https://%v/counterparty?id=%v&lang=%v", host, counterpartyID, locale.SiteCode())))
	// TODO: Commented due to Telegram issue with too long URL
	if !utmParams.IsEmpty() {
		_, _ = writer.Write([]byte(fmt.Sprintf("&%v", utmParams.ShortString())))
	}
	if currentUserID != 0 {
		token := auth.IssueToken(currentUserID, formatIssuer(utmParams.Medium, utmParams.Source), false)
		_, _ = writer.Write([]byte(fmt.Sprintf("&secret=%v", token)))
	}
}
