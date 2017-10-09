package dtb_common

import (
	"github.com/strongo/bots-framework/core"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
)

func GetUser(whc bots.WebhookContext) (user models.AppUser, err error) {
	var appUser bots.BotAppUser
	if appUser, err = whc.GetAppUser(); err != nil {
		return
	}
	user.AppUserEntity = appUser.(*models.AppUserEntity)
	user.ID = whc.AppUserIntID()
	return
}
