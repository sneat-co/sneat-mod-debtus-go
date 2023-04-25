package common

import (
	tgstore "github.com/bots-go-framework/bots-fw-telegram/store"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/sneat-co/debtstracker-translations/trans"
	"reflect"
	"time"

	"context"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/models"
	"github.com/strongo/app"
)

type DebtsTrackerAppContext struct {
}

var _ botsfw.BotAppContext = (*DebtsTrackerAppContext)(nil)

func (appCtx DebtsTrackerAppContext) AppUserEntityKind() string {
	return models.AppUserKind
}

func (appCtx DebtsTrackerAppContext) AppUserEntityType() reflect.Type {
	return reflect.TypeOf(&models.AppUserData{})
}

func (appCtx DebtsTrackerAppContext) NewBotAppUserEntity() botsfw.BotAppUser {
	return &models.AppUserData{
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
			return &models.DebtusTelegramChatData{
				TgChatBase: *tgstore.NewTelegramChatEntity(),
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
