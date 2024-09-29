package debtus

import (
	"github.com/sneat-co/sneat-go-core/module"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/api4debtus"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/const4debtus"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/facade4debtus"
)

func Module() module.Module {
	return module.NewModule(const4debtus.ModuleID,
		module.RegisterRoutes(api4debtus.RegisterHttpRoutes),
		module.RegisterDelays(facade4debtus.InitDelays4debtus),
	)
}
