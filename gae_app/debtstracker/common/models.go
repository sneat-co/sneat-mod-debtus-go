package common

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bytes"
	"fmt"
	"github.com/strongo/app"
	"io"
)

func GetCounterpartyUrl(counterpartyID, currentUserID int64, locale strongo.Locale, utmParams UtmParams) string {
	var buffer bytes.Buffer
	WriteCounterpartyUrl(&buffer, counterpartyID, currentUserID, locale, utmParams)
	return buffer.String()
}

func WriteCounterpartyUrl(writer io.Writer, counterpartyID, currentUserID int64, locale strongo.Locale, utmParams UtmParams) {
	host := GetWebsiteHost(utmParams.Source)
	writer.Write([]byte(fmt.Sprintf("https://%v/counterparty?id=%v&lang=%v", host, counterpartyID, locale.SiteCode())))
	// TODO: Commented due to Telegram issue with too long URL
	if !utmParams.IsEmpty() {
		writer.Write([]byte(fmt.Sprintf("&%v", utmParams.ShortString())))
	}
	if currentUserID != 0 {
		token := auth.IssueToken(currentUserID, formatIssuer(utmParams.Medium, utmParams.Source), false)
		writer.Write([]byte(fmt.Sprintf("&secret=%v", token)))
	}
}
