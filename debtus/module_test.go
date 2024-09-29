package debtus

import (
	"github.com/sneat-co/sneat-go-core/module"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/const4debtus"
	"testing"
)

func TestModule(t *testing.T) {
	m := Module()
	module.AssertModule(t, m, module.Expected{
		ModuleID:      const4debtus.ModuleID,
		HandlersCount: 0,
		DelayersCount: 2,
	})
}
