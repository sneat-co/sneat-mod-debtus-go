package admin

import (
	"github.com/strongo/app"
	"net/http"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_fbm"
)

func InitAdmin() {
	strongo.AddHttpHandler("/admin/latest", LatestPage)
	strongo.AddHttpHandler("/admin/clean", CleanupPage)
	strongo.AddHttpHandler("/admin/mass-invites", LatestPage)
	strongo.AddHttpHandler("/admin/fix/transfers", FixTransfersHandler)
	http.HandleFunc("/admin/fbm/set", dtb_fbm.SetupFbm)
}



