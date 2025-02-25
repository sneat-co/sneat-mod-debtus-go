package debtus

import (
	"github.com/sneat-co/sneat-go-core/module"
	"testing"
)

func TestModule(t *testing.T) {
	m := Module()
	module.AssertModule(t, m, module.Expected{
		ModuleID:      moduleID,
		HandlersCount: 4,
		DelayersCount: 2,
	})
}
