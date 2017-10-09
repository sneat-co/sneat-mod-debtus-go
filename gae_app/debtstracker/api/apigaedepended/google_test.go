package apigaedepended

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"net/http"
	"testing"
)

func TestInitApiGaeDepended(t *testing.T) {
	i, j := 0, 0
	handleFunc = func(pattern string, handler func(http.ResponseWriter, *http.Request)) {
		i += 1
	}
	dal.HandleWithContext = func(handler dal.ContextHandler) func(w http.ResponseWriter, r *http.Request) {
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
