package common

import (
	"bytes"
	"fmt"
	"github.com/strongo/i18n"
	"io"
	"strconv"

	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/auth"
)

func GetCounterpartyUrl(counterpartyID int64, currentUserID int64, locale i18n.Locale, utmParams UtmParams) string {
	var buffer bytes.Buffer
	WriteCounterpartyUrl(&buffer, counterpartyID, strconv.FormatInt(currentUserID, 10), locale, utmParams)
	return buffer.String()
}

func WriteCounterpartyUrl(writer io.Writer, counterpartyID int64, currentUserID string, locale i18n.Locale, utmParams UtmParams) {
	host := GetWebsiteHost(utmParams.Source)
	_, _ = writer.Write([]byte(fmt.Sprintf("https://%v/counterparty?id=%v&lang=%v", host, counterpartyID, locale.SiteCode())))
	// TODO: Commented due to Telegram issue with too long URL
	if !utmParams.IsEmpty() {
		_, _ = writer.Write([]byte(fmt.Sprintf("&%v", utmParams.ShortString())))
	}
	if currentUserID != "" && currentUserID != "0" {
		token := auth.IssueToken(currentUserID, formatIssuer(utmParams.Medium, utmParams.Source), false)
		_, _ = writer.Write([]byte(fmt.Sprintf("&secret=%v", token)))
	}
}
