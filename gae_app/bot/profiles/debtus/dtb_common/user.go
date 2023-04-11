package dtb_common

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/bots-go-framework/bots-fw/botsfw"
)

func GetUser(whc botsfw.WebhookContext) (user models.AppUser, err error) {
	var appUser botsfw.BotAppUser
	if appUser, err = whc.GetAppUser(); err != nil {
		return
	}
	user.Data = appUser.(*models.AppUserEntity)
	user.ID = whc.AppUserIntID()
	return
}
