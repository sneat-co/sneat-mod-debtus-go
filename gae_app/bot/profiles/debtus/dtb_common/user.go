package dtb_common

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
)

func GetUser(whc botsfw.WebhookContext) (user models.AppUser, err error) {
	var appUser botsfw.BotAppUser
	if appUser, err = whc.GetAppUser(); err != nil {
		return
	}
	user.AppUserEntity = appUser.(*models.AppUserEntity)
	user.ID = whc.AppUserIntID()
	return
}
