package common

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"golang.org/x/net/context"
	"reflect"
	"time"
)

type DebtsTrackerAppContext struct {
}

var _ bots.BotAppContext = (*DebtsTrackerAppContext)(nil)

func (appCtx DebtsTrackerAppContext) AppUserEntityKind() string {
	return models.AppUserKind
}

func (appCtx DebtsTrackerAppContext) AppUserEntityType() reflect.Type {
	return reflect.TypeOf(&models.AppUserEntity{})
}

func (appCtx DebtsTrackerAppContext) NewBotAppUserEntity() bots.BotAppUser {
	return &models.AppUserEntity{
		ContactDetails: models.ContactDetails{
			PhoneContact: models.PhoneContact{},
		},
		DtCreated: time.Now(),
	}
}

func (appCtx DebtsTrackerAppContext) GetBotChatEntityFactory(platform string) func() bots.BotChat {
	switch platform {
	case "telegram":
		return func() bots.BotChat {
			return &models.DtTelegramChatEntity{
				TelegramChatEntityBase: *telegram_bot.NewTelegramChatEntity(),
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

var _ bots.BotAppContext = (*DebtsTrackerAppContext)(nil)

var TheAppContext = DebtsTrackerAppContext{}
