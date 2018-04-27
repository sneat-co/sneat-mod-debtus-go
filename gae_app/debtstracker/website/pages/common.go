package pages

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"bitbucket.com/asterus/debtstracker-server/gae_app/bot"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/platforms/tgbots"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"github.com/strongo/app/gaestandard"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app"
	"github.com/strongo/log"
	"context"
	"google.golang.org/appengine"
)

func pageContext(r *http.Request, locale strongo.Locale) (translator strongo.SingleLocaleTranslator, data map[string]interface{}) {
	userVoiceID := "6ed87444-76e3-43ee-8b6e-fd28d345e79c" // English
	c := appengine.NewContext(r)

	switch locale.Code5 {
	case strongo.LOCALE_RU_RU:
		userVoiceID = "47c67b85-d064-4727-b149-bda58cfe6c2d"
	}

	appTranslator := common.TheAppContext.GetTranslator(c)
	translator = strongo.NewSingleMapTranslator(locale, appTranslator)

	if locale.Code5 != strongo.LOCALE_EN_US {
		translator = strongo.NewSingleLocaleTranslatorWithBackup(translator, strongo.NewSingleMapTranslator(strongo.LocaleEnUS, appTranslator))
	}

	env := gaestandard.GetEnvironmentFromHost(r.Host)
	if env == strongo.EnvUnknown {
		panic("Unknown host: " + r.Host)
	}
	botSettings, err := tgbots.GetBotSettingsByLang(gaestandard.GetEnvironment(c), bot.ProfileDebtus, locale.Code5)
	if err != nil {
		panic(err)
	}

	data = map[string]interface{}{
		"lang":          locale.SiteCode(),
		"userVoiceID":   userVoiceID,
		"TgBotID":       botSettings.Code,
		"SubLocalePath": strings.Replace(r.URL.EscapedPath(), fmt.Sprintf("/%v/", locale.SiteCode()), "/", 1),
		trans.WS_ALEX_T: translator.TranslateNoWarning(trans.WS_ALEX_T),
		trans.WS_MOTTO:  translator.Translate(trans.WS_MOTTO),
	}
	return translator, data
}

func getLocale(c context.Context, w http.ResponseWriter, r *http.Request) (locale strongo.Locale, err error) {
	getLocaleBySiteCode := func(localeCode string) {
		for _, supportedLocale := range strongo.LocalesByCode5 {
			if supportedLocale.SiteCode() == localeCode {
				locale = supportedLocale
				break
			}
		}
	}

	path := r.URL.Path
	if path == "/" {
		if localeCode, ok := c.Value("locale").(string); !ok {
			locale = strongo.LocaleEnUS
		} else {
			getLocaleBySiteCode(localeCode)
			if locale.Code5 == "" {
				locale = strongo.LocaleEnUS
			}
		}
		return
	} else {
		if strings.HasPrefix(path, "/ru/") {
			locale = strongo.LocaleRuRu
		} else if strings.HasPrefix(path, "/zh/") {
			locale = strongo.LocaleZhCn
		} else if strings.HasPrefix(path, "/ja/") {
			locale = strongo.LocaleJaJp
		} else if strings.HasPrefix(path, "/fa/") {
			locale = strongo.LocaleFaIr
		} else {
			nextSlashIndex := strings.Index(path[1:], "/")
			if nextSlashIndex == -1 {
				err = fmt.Errorf("Unsupported path: %v", path)
				w.WriteHeader(http.StatusNotFound)
				w.Header().Set("Content-Type", "text/plain")
				w.Write(([]byte)(err.Error()))
				return
			} else {
				localeCode := path[1 : nextSlashIndex+1]
				getLocaleBySiteCode(localeCode)
				if locale.Code5 == "" {
					w.WriteHeader(http.StatusNotFound)
					w.Header().Set("Content-Type", "text/plain")
					if _, err := w.Write(([]byte)(fmt.Sprintf("Unsupported locale: %v", localeCode))); err != nil {
						log.Errorf(c, err.Error())
					}
					return
				}
			}
		}
	}
	return
}

func RenderCachedPage(w http.ResponseWriter, r *http.Request, tmpl *template.Template, locale strongo.Locale, data map[string]interface{}, maxAge int) {
	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(buffer.Bytes())
		w.Write([]byte("<hr><div style=color:red;position:absolute;padding:10px;background-color:white>" + err.Error() + "</div>"))
		return
	}
	eTag := fmt.Sprintf("%x", md5.Sum(buffer.Bytes()))
	if match := r.Header.Get("If-None-Match"); match == eTag {
		w.WriteHeader(http.StatusNotModified)
	} else {
		header := w.Header()
		header.Set("Content-Language", locale.Code5)
		if maxAge >= 0 {
			if maxAge == 0 {
				maxAge = 600
			}
			header.Set("Cache-Control", fmt.Sprintf("public, max-age=%v", maxAge))
		}
		header.Set("ETag", eTag)
		w.Write(buffer.Bytes())
	}
}
