package common

import (
	"reflect"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/bots-go-framework/bots-fw-telegram"
	"github.com/strongo/app"
)

type DebtsTrackerAppContext struct {
}

var _ botsfw.BotAppContext = (*DebtsTrackerAppContext)(nil)

func (appCtx DebtsTrackerAppContext) AppUserEntityKind() string {
	return models.AppUserKind
}

func (appCtx DebtsTrackerAppContext) AppUserEntityType() reflect.Type {
	return reflect.TypeOf(&models.AppUserEntity{})
}

func (appCtx DebtsTrackerAppContext) NewBotAppUserEntity() botsfw.BotAppUser {
	return &models.AppUserEntity{
		ContactDetails: models.ContactDetails{
			PhoneContact: models.PhoneContact{},
		},
		DtCreated: time.Now(),
	}
}

func (appCtx DebtsTrackerAppContext) GetBotChatEntityFactory(platform string) func() botsfw.BotChat {
	switch platform {
	case "telegram":
		return func() botsfw.BotChat {
			return &models.DtTelegramChatEntity{
				TgChatEntityBase: *telegram.NewTelegramChatEntity(),
			}
		}
	default:
		panic("Unknown platform: " + platform)
	}
}

func (appCtx DebtsTrackerAppContext) NewAppUserEntity() strongo.AppUser {
	return appCtx.NewBotAppUserEntity()
}

func (appCtx DebtsTrackerAppContext) GetTranslator(c context.Context) strongo.Translator {
	return strongo.NewMapTranslator(c, trans.TRANS)
}

func (appCtx DebtsTrackerAppContext) SupportedLocales() strongo.LocalesProvider {
	return trans.DebtsTrackerLocales{}
}

var _ botsfw.BotAppContext = (*DebtsTrackerAppContext)(nil)

var TheAppContext = DebtsTrackerAppContext{}
