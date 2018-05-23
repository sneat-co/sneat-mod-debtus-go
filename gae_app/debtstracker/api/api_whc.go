package api

import (
	"net/http"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
)

type ApiWebhookContext struct {
	appUser      *models.AppUserEntity
	appUserIntID int64
	botChatID    int64
	chatEntity   bots.BotChat
	*bots.WebhookContextBase
}

var _ bots.WebhookContext = (*ApiWebhookContext)(nil)

func (ApiWebhookContext) IsInGroup() bool {
	panic("not supported")
}

func NewApiWebhookContext(r *http.Request, appUser *models.AppUserEntity, userID, botChatID int64, chatEntity bots.BotChat) ApiWebhookContext {
	var botSettings bots.BotSettings
	whc := ApiWebhookContext{
		appUser:      appUser,
		appUserIntID: userID,
		botChatID:    botChatID,
		chatEntity:   chatEntity,
		WebhookContextBase: bots.NewWebhookContextBase(
			r,
			common.TheAppContext,
			telegram.Platform{},
			*bots.NewBotContext(dal.BotHost, botSettings),
			nil, // webhookInput
			bots.BotCoreStores{},
			nil, // GaMeasurement
			func() bool { return false },
			nil,
		),
	}
	whc.SetLocale(chatEntity.GetPreferredLanguage())
	return whc
}

func (whc ApiWebhookContext) AppUserIntID() int64 {
	return whc.appUserIntID
}

func (whc ApiWebhookContext) BotChatIntID() int64 {
	return whc.botChatID
}

func (whc ApiWebhookContext) ChatEntity() bots.BotChat {
	return whc.chatEntity
}

func (whc ApiWebhookContext) GetAppUser() (bots.BotAppUser, error) {
	return whc.appUser, nil
}

func (whc ApiWebhookContext) Init(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (whc ApiWebhookContext) IsNewerThen(chatEntity bots.BotChat) bool {
	return true
}

func (whc ApiWebhookContext) MessageText() string {
	return ""
}

func (whc ApiWebhookContext) NewEditMessage(text string, format bots.MessageFormat) (m bots.MessageFromBot, err error) {
	panic("Not implemented")
}

func (whc ApiWebhookContext) Responder() bots.WebhookResponder {
	panic("Not implemented")
}

func (whc ApiWebhookContext) UpdateLastProcessed(chatEntity bots.BotChat) error {
	panic("Not implemented")
}
