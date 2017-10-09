package dtb_common

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/bots-framework/core"
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
