package splitus

import (
	"github.com/sneat-co/sneat-go-core/module"
	"github.com/sneat-co/sneat-mod-debtus-go/splitus/api4splitusbot"
	"github.com/sneat-co/sneat-mod-debtus-go/splitus/const4splitus"
	"github.com/strongo/strongoapp"
	"net/http"
)

const moduleID = const4splitus.ModuleID

func Module() module.Module {
	return module.NewModule(moduleID,
		module.RegisterRoutes(func(handle module.HTTPHandleFunc) {
			// TODO: This should be unified with the rest of APIs
			api4splitusbot.InitApiForSplitus(func(method, path string, handler strongoapp.HttpHandlerWithContext) {
				handle(method, path, func(writer http.ResponseWriter, request *http.Request) {
					handler(request.Context(), writer, request)
				})
			})
		}),
		//module.RegisterDelays(facade4debtus.InitDelays4debtus),
	)
}
