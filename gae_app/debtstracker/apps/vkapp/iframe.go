package vkapp

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/website/pages"
	"html/template"
	"net/http"
	//"github.com/strongo/app"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/platforms/vk"
	"github.com/strongo/app"
	"github.com/julienschmidt/httprouter"
)

func InitVkIFrameApp(router *httprouter.Router) {
	router.HandlerFunc("GET", "/apps/vk/iframe", IFrameHandler)
}

func IFrameHandler(w http.ResponseWriter, r *http.Request) {
	if vkIFrameTemplate == nil {
		vkIFrameTemplate = template.Must(
			template.ParseFiles(
				pages.TEMPLATES_PATH+"vk-iframe.html",
				pages.TEMPLATES_PATH+"device-switcher.html",
				pages.TEMPLATES_PATH+"device.js.html",
			),
		)
	}
	query := r.URL.Query()
	apiID := query.Get("api_id")
	_, ok := vk.BotsBy.ByCode[apiID]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unknown app id"))
		return
	}

	lang := "ru"
	if query.Get("language") == "3" {
		lang = "en"
	}

	data := map[string]interface{}{
		"vkApiId": apiID,
		"lang":    lang,
		"hash":    query.Get("hash"),
	}

	pages.RenderCachedPage(w, r, vkIFrameTemplate, strongo.LocaleRuRu, data, 0)
}

var vkIFrameTemplate *template.Template
