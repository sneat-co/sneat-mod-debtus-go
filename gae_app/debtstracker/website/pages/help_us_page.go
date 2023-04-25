package pages

import (
	"github.com/julienschmidt/httprouter"
	"html/template"
	"net/http"

	"github.com/sneat-co/debtstracker-go/gae_app/bot/profiles/debtus/cmd/dtb_general"
	"github.com/sneat-co/debtstracker-translations/trans"
	"google.golang.org/appengine"
)

var helpUsPageTmpl *template.Template

func HelpUsPage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	locale, err := getLocale(appengine.NewContext(r), w, r)
	if err != nil {
		return
	}
	translator, data := pageContext(r, locale)

	for _, key := range []string{
		trans.WS_HELP_US_TITLE,
	} {
		data[key] = template.HTML(translator.Translate(key))
	}
	var tgBotCode string
	switch locale.Code5 {
	case "ru-RU":
		tgBotCode = "DebtsTrackerRuBot"
	default:
		tgBotCode = "DebtsTrackerBot"
	}
	content := dtb_general.YouCanHelp(translator, trans.MESSAGE_TEXT_YOU_CAN_HELP_BY_HTML, tgBotCode)
	data[trans.WS_HELP_US_CONTENT] = template.HTML(content)

	if helpUsPageTmpl == nil {
		helpUsPageTmpl = template.Must(template.ParseFiles(
			BASE_TEMPLATE,
			TEMPLATES_PATH+"help-us.html",
			TEMPLATES_PATH+"device-switcher.html",
			TEMPLATES_PATH+"device.js.html",
		))
	}
	RenderCachedPage(w, r, helpUsPageTmpl, locale, data, 0)
}
