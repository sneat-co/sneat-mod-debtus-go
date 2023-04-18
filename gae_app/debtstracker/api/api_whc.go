package api

import (
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/strongo/log"
	"net/http"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/bots-go-framework/bots-fw-telegram"
)

type ApiWebhookContext struct {
	appUser      *models.AppUserData
	appUserIntID int64
	botChatID    int64
	chatEntity   botsfw.BotChat
	*botsfw.WebhookContextBase
}

var _ botsfw.WebhookContext = (*ApiWebhookContext)(nil)

func (ApiWebhookContext) IsInGroup() bool {
	panic("not supported")
}

func NewApiWebhookContext(r *http.Request, appUser *models.AppUserData, userID, botChatID int64, chatEntity botsfw.BotChat) ApiWebhookContext {
	var botSettings botsfw.BotSettings
	whc := ApiWebhookContext{
		appUser:      appUser,
		appUserIntID: userID,
		botChatID:    botChatID,
		chatEntity:   chatEntity,
		WebhookContextBase: botsfw.NewWebhookContextBase(
			r,
			common.TheAppContext,
			telegram.Platform{},
			*botsfw.NewBotContext(dtdal.BotHost, botSettings),
			nil, // webhookInput
			botsfw.BotCoreStores{},
			nil, // GaMeasurement
			func() bool { return false },
			nil,
		),
	}
	if err := whc.SetLocale(chatEntity.GetPreferredLanguage()); err != nil {
		log.Errorf(r.Context(), "failed to set locale: %v", err)
	}
	return whc
}

func (whc ApiWebhookContext) AppUserIntID() int64 {
	return whc.appUserIntID
}

func (whc ApiWebhookContext) BotChatIntID() int64 {
	return whc.botChatID
}

func (whc ApiWebhookContext) ChatEntity() botsfw.BotChat {
	return whc.chatEntity
}

func (whc ApiWebhookContext) GetAppUser() (botsfw.BotAppUser, error) {
	return whc.appUser, nil
}

func (whc ApiWebhookContext) Init(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (whc ApiWebhookContext) IsNewerThen(chatEntity botsfw.BotChat) bool {
	return true
}

func (whc ApiWebhookContext) MessageText() string {
	return ""
}

func (whc ApiWebhookContext) NewEditMessage(text string, format botsfw.MessageFormat) (m botsfw.MessageFromBot, err error) {
	panic("Not implemented")
}

func (whc ApiWebhookContext) Responder() botsfw.WebhookResponder {
	panic("Not implemented")
}

func (whc ApiWebhookContext) UpdateLastProcessed(chatEntity botsfw.BotChat) error {
	panic("Not implemented")
}
