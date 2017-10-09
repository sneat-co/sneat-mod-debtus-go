package common

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bytes"
	"fmt"
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/templates/inspiration/builtin/html"
	"io"
	"strconv"
	"strings"
)

func GetBalanceUrlForUser(userID int64, locale strongo.Locale, createdOnPlatform, createdOnID string) string {
	return getUrlForUser(userID, locale, "debts", createdOnPlatform, createdOnID)
}

func GetHistoryUrlForUser(userID int64, locale strongo.Locale, createdOnPlatform, createdOnID string) string {
	return getUrlForUser(userID, locale, "history", createdOnPlatform, createdOnID)
}

func getUrlForUser(userID int64, locale strongo.Locale, page, createdOnPlatform, createdOnID string) string {
	token := auth.IssueToken(userID, createdOnPlatform+":"+createdOnID, false)
	host := GetWebsiteHost(createdOnID)
	url := fmt.Sprintf("https://%v/app/#", host)
	switch page {
	case "history":
		url += "/history/"
	case "debts":
		url += "/debts/"
	default:
		url += "page=" + page
	}
	return url + fmt.Sprintf("&lang=%v&secret=%v", locale.SiteCode(), token)
}

func GetTransferUrlForUser(transferID, userID int64, locale strongo.Locale, utmParams UtmParams) string {
	var buffer bytes.Buffer
	WriteTransferUrlForUser(&buffer, transferID, userID, locale, utmParams)
	return buffer.String()
}

func WriteTransferUrlForUser(writer io.Writer, transferID, userID int64, locale strongo.Locale, utmParams UtmParams) {
	host := GetWebsiteHost(utmParams.Source)
	writer.Write([]byte(fmt.Sprintf(
		"https://%v/transfer?id=%v&lang=%v",
		host, strconv.FormatInt(transferID, 10), locale.SiteCode(),
	)))
	if !utmParams.IsEmpty() {
		writer.Write([]byte(fmt.Sprintf("&%v", utmParams.ShortString())))
	}
	if userID != 0 {
		token := auth.IssueToken(userID, formatIssuer(utmParams.Medium, utmParams.Source), false)
		writer.Write([]byte(fmt.Sprintf("&secret=%v", token)))
	}
}

func GetChooseCurrencyUrlForUser(userID int64, locale strongo.Locale, createdOnPlatform, createdOnID, contextData string) string {
	token := auth.IssueToken(userID, createdOnPlatform+":"+createdOnID, false)
	host := GetWebsiteHost(createdOnID)
	return fmt.Sprintf(
		"https://%v/app/#/choose-currency?lang=%v&%v&secret=%v",
		host, locale.SiteCode(), contextData, token,
	)
}

func GetWebsiteHost(createdOnID string) string {
	createdOnID = strings.ToLower(createdOnID)
	if strings.Contains(createdOnID, "dev") {
		return "debtstracker-dev1.appspot.com"
	} else if strings.Contains(createdOnID, ".local") {
		return "debtstracker.local"
	} else {
		return "debtstracker.io"
	}
}

func GetPathAndQueryForInvite(inviteCode string, utmParams UtmParams) string {
	return fmt.Sprintf("ack?invite=%v#%v", html.URLQueryEscaper(inviteCode), utmParams)
}

func GetNewDebtPageUrl(whc bots.WebhookContext, direction models.TransferDirection, utmCampaign string) string {
	botID := whc.GetBotCode()
	botPlatform := whc.BotPlatform().Id()
	token := auth.IssueToken(whc.AppUserIntID(), formatIssuer(botPlatform, botID), false)
	host := GetWebsiteHost(botID)
	//utmParams := NewUtmParams(whc, utmCampaign)
	return fmt.Sprintf(
		"https://%v/open/new-debt?d=%v&lang=%v&secret=%v",
		host, direction, whc.Locale().SiteCode(), token, //utmParams, - commented out as: Viber response.Status=3: keyboard is not valid. is too long (length: 274, maximum allowed: 250)]
	)
}
