package pages

import (
	"github.com/DebtsTracker/translations/trans"
	"google.golang.org/appengine"
	"html/template"
	"net/http"
)

var adsPageTmpl *template.Template

func AdsPage(w http.ResponseWriter, r *http.Request) {
	locale, err := getLocale(appengine.NewContext(r), w, r)
	if err != nil {
		return
	}
	translator, data := pageContext(r, locale)

	for _, key := range []string{
		trans.WS_ADS_TITLE,
		trans.WS_ADS_CONTENT,
	} {
		data[key] = template.HTML(translator.Translate(key))
	}

	if adsPageTmpl == nil {
		adsPageTmpl = template.Must(template.ParseFiles(
			BASE_TEMPLATE,
			TEMPLATES_PATH+"ads.html",
			TEMPLATES_PATH+"device-switcher.html",
			TEMPLATES_PATH+"device.js.html",
		))
	}
	RenderCachedPage(w, r, adsPageTmpl, locale, data, 0)
}
