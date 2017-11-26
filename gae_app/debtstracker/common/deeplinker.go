package common

import (
	"bytes"
	"fmt"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
)

type deeplink struct {
}

func (_ deeplink) AppHashPathToReceipt(receiptID int64) string {
	return fmt.Sprintf("receipt=%d", receiptID)
}

var Deeplink = deeplink{}

type Linker struct {
	userID int64
	locale string
	issuer string
	host   string
}

func NewLinker(environment strongo.Environment, userID int64, locale, issuer string) Linker {
	return Linker{
		userID: userID,
		locale: locale,
		issuer: issuer,
		host:   host(environment),
	}
}

func NewLinkerFromWhc(whc bots.WebhookContext) Linker {
	return NewLinker(whc.Environment(), whc.AppUserIntID(), whc.Locale().SiteCode(), formatIssuer(whc.BotPlatform().Id(), whc.GetBotCode()))
}

func host(environment strongo.Environment) string {
	switch environment {
	case strongo.EnvProduction:
		return "debtstracker.io"
	case strongo.EnvLocal:
		return "debtstracker.local"
	case strongo.EnvDevTest:
		return "debtstracker-dev1.appspot.com"
	}
	panic(fmt.Sprintf("Unknown environment: %v", environment))
}

func (l Linker) UrlToContact(contactID int64) string {
	return l.url("/contact", fmt.Sprintf("?id=%d", contactID), "")
}

func formatIssuer(botPlatform, botID string) string {
	return fmt.Sprintf("%v:%v", botPlatform, botID)
}

func (l Linker) url(path, query, hash string) string {
	var buffer bytes.Buffer
	buffer.WriteString("https://" + l.host + path + query)
	if hash != "" {
		buffer.WriteString(hash)
	}
	if query != "" || hash != "" {
		buffer.WriteString("&")
	}
	isAdmin := false // TODO: How to get isAdmin?
	token := auth.IssueToken(l.userID, l.issuer, isAdmin)
	buffer.WriteString("lang=" + l.locale)
	buffer.WriteString("&secret=" + token)
	return buffer.String()
}

func (l Linker) ToMainScreen(whc bots.WebhookContext) string {
	return l.url("/app/", "", "#")
}
