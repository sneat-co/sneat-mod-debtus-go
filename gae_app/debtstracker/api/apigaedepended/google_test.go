package apigaedepended

import (
	strongo "github.com/strongo/app"
	"net/http"
	"testing"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
)

func TestInitApiGaeDepended(t *testing.T) {
	i, j := 0, 0
	handleFunc = func(pattern string, handler func(http.ResponseWriter, *http.Request)) {
		i += 1
	}
	dtdal.HandleWithContext = func(handler strongo.ContextHandler) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			j += 1
		}
	}
	InitApiGaeDepended()
	if i != 2 {
		t.Errorf("i:%d != 2", i)
	}
	if j != 0 {
		t.Errorf("j=%v", j)
	}
}
