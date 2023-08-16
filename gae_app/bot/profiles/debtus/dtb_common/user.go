package dtb_common

import (
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/debtstracker/models"
)

func GetUser(whc botsfw.WebhookContext) (user models.AppUser, err error) {
	panic("not implemented")
	//var appUser botsfwmodels.BotAppUser
	//if appUser, err = whc.GetAppUser(); err != nil {
	//	return
	//}
	//user.Data = appUser.(*models.AppUserData)
	//user.ID, err = strconv.ParseInt(whc.AppUserID(), 10, 64)
	//return
}
