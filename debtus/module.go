package debtus

import (
	"github.com/sneat-co/sneat-go-core/module"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/api/api4debtus"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/api/api4transfers"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/const4debtus"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/facade4debtus"
	"github.com/strongo/strongoapp"
	"net/http"
)

const moduleID = const4debtus.ModuleID

func Module() module.Module {
	return module.NewModule(moduleID,
		module.RegisterRoutes(func(handle module.HTTPHandleFunc) {
			// TODO: This should be unified with the rest of APIs
			api4debtus.InitApiForDebtus(func(method, path string, handler strongoapp.HttpHandlerWithContext) {
				handle(method, path, func(writer http.ResponseWriter, request *http.Request) {
					handler(request.Context(), writer, request)
				})
			})
		}),
		module.RegisterRoutes(func(handle module.HTTPHandleFunc) {
			// TODO: This should be unified with the rest of APIs
			api4transfers.InitApiForTransfers(func(method, path string, handler strongoapp.HttpHandlerWithContext) {
				handle(method, path, func(writer http.ResponseWriter, request *http.Request) {
					handler(request.Context(), writer, request)
				})
			})
		}),
		module.RegisterDelays(facade4debtus.InitDelays4debtus),
	)
}
