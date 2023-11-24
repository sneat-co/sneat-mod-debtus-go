package common

import (
	"context"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/sneat-co/debtstracker-translations/trans"
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/debtstracker/models"
	"github.com/strongo/i18n"
	"github.com/strongo/strongoapp/appuser"
	"reflect"
	"time"
)

type DebtsTrackerAppContext struct {
}

func (appCtx DebtsTrackerAppContext) AppUserCollectionName() string {
	//TODO implement me
	panic("implement me")
}

func (appCtx DebtsTrackerAppContext) SetLocale(code5 string) error {
	//TODO implement me
	panic("implement me")
}

var _ botsfw.BotAppContext = (*DebtsTrackerAppContext)(nil)

func (appCtx DebtsTrackerAppContext) AppUserEntityKind() string {
	return models.AppUserKind
}

func (appCtx DebtsTrackerAppContext) AppUserEntityType() reflect.Type {
	return reflect.TypeOf(&models.AppUserData{})
}

func (appCtx DebtsTrackerAppContext) NewBotAppUserEntity() botsfwmodels.AppUserData {
	return &models.AppUserData{
		ContactDetails: models.ContactDetails{
			PhoneContact: models.PhoneContact{},
		},
		DtCreated: time.Now(),
	}
}

func (appCtx DebtsTrackerAppContext) GetBotChatEntityFactory(platform string) func() botsfwmodels.BotChatData {
	switch platform {
	case "telegram":
		panic("not implemented")
		//return func() botsfwmodels.ChatBaseData {
		//	return &models.DebtusTelegramChatData{
		//		TgChatBase: *botsfwtgmodels.NewTelegramChatEntity(),
		//	}
		//}
	default:
		panic("Unknown platform: " + platform)
	}
}

func (appCtx DebtsTrackerAppContext) NewAppUserData() appuser.BaseUserData {
	return appCtx.NewBotAppUserEntity()
}

func (appCtx DebtsTrackerAppContext) GetTranslator(c context.Context) i18n.Translator {
	return i18n.NewMapTranslator(c, trans.TRANS)
}

func (appCtx DebtsTrackerAppContext) SupportedLocales() i18n.LocalesProvider {
	return trans.DebtsTrackerLocales{}
}

var _ botsfw.BotAppContext = (*DebtsTrackerAppContext)(nil)

var TheAppContext = DebtsTrackerAppContext{}
