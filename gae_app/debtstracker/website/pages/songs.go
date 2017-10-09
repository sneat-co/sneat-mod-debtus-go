package pages

import (
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app"
	"html/template"
	"net/http"
)

var iouADanceTmpl *template.Template

func AnnieIOUaDancePage(w http.ResponseWriter, r *http.Request) {
	if iouADanceTmpl == nil {
		iouADanceTmpl = template.Must(template.ParseFiles(
			BASE_TEMPLATE,
			TEMPLATES_PATH+"song-iou-a-dance.html",
			TEMPLATES_PATH+"device-switcher.html",
			TEMPLATES_PATH+"device.js.html",
		))
	}

	translator, data := pageContext(r, strongo.LocaleEnUS)
	for _, key := range []string{
		trans.WS_SHORT_DESC,
		trans.WS_LIVE_DEMO,
	} {
		data[key] = template.HTML(translator.Translate(key))
	}
	data["SubLocalePath"] = "/"
	RenderCachedPage(w, r, iouADanceTmpl, strongo.LocaleEnUS, data, 0)
}

var iouDappyTmpl *template.Template

func IOWDappyPage(w http.ResponseWriter, r *http.Request) {
	if iouDappyTmpl == nil {
		iouDappyTmpl = template.Must(template.ParseFiles(
			BASE_TEMPLATE,
			TEMPLATES_PATH+"song-iou-dappy.html",
			TEMPLATES_PATH+"device-switcher.html",
			TEMPLATES_PATH+"device.js.html",
		))
	}

	translator, data := pageContext(r, strongo.LocaleEnUS)
	data["SubLocalePath"] = "/"
	for _, key := range []string{
		trans.WS_SHORT_DESC,
		trans.WS_LIVE_DEMO,
	} {
		data[key] = template.HTML(translator.Translate(key))
	}
	RenderCachedPage(w, r, iouDappyTmpl, strongo.LocaleEnUS, data, 0)
}
