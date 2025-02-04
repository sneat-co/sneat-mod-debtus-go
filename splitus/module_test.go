package splitus

import (
	"github.com/sneat-co/sneat-go-core/module"
	"github.com/sneat-co/sneat-mod-debtus-go/splitus/const4splitus"
	"testing"
)

func TestModule(t *testing.T) {
	m := Module()
	module.AssertModule(t, m, module.Expected{
		ModuleID:      const4splitus.ModuleID,
		HandlersCount: 2,
		DelayersCount: 0,
	})
}
