package fbm

import "testing"

func TestFbAppSecrets_App(t *testing.T) {
	secrets := &fbAppSecrets{}
	app := secrets.App()
	if app == nil {
		t.Error("fbAppSecrets{}.App() rturned nil")
	}
}
