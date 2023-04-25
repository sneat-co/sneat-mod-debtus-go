package dtb_common

import (
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/models"
)

func GetUser(whc botsfw.WebhookContext) (user models.AppUser, err error) {
	var appUser botsfw.BotAppUser
	if appUser, err = whc.GetAppUser(); err != nil {
		return
	}
	user.Data = appUser.(*models.AppUserData)
	user.ID = whc.AppUserIntID()
	return
}
