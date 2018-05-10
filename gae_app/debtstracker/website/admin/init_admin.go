package admin

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_fbm"
	"github.com/julienschmidt/httprouter"
)

type router interface {
	GET(path string, handle httprouter.Handle)
}

func InitAdmin(router router) {
	router.GET("/admin/latest", LatestPage)
	router.GET("/admin/clean", CleanupPage)
	//strongo.AddHttpHandler("/admin/mass-invites", LatestPage)
	router.GET("/admin/fix/transfers", FixTransfersHandler)
	router.GET("/admin/fbm/set", dtb_fbm.SetupFbm)
}
