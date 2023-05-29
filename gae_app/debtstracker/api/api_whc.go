package api

import (
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/strongo/log"
	"net/http"

	"github.com/bots-go-framework/bots-fw-telegram"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/common"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/dtdal"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/models"
)

type ApiWebhookContext struct {
	appUser      *models.AppUserData
	appUserIntID int64
	botChatID    int64
	chatEntity   botsfwmodels.ChatData
	*botsfw.WebhookContextBase
}

var _ botsfw.WebhookContext = (*ApiWebhookContext)(nil)

func (ApiWebhookContext) IsInGroup() bool {
	panic("not supported")
}

func NewApiWebhookContext(r *http.Request, appUser *models.AppUserData, userID, botChatID int64, chatData botsfwmodels.ChatData) ApiWebhookContext {
	var botSettings botsfw.BotSettings
	whc := ApiWebhookContext{
		appUser:      appUser,
		appUserIntID: userID,
		botChatID:    botChatID,
		chatEntity:   chatData,
		WebhookContextBase: botsfw.NewWebhookContextBase(
			r,
			common.TheAppContext,
			telegram.Platform,
			*botsfw.NewBotContext(dtdal.BotHost, botSettings),
			nil, // webhookInput
			nil,
			nil, // records fields setter
			nil, // GaMeasurement
			func() bool { return false },
			nil,
		),
	}
	if err := whc.SetLocale(chatData.GetPreferredLanguage()); err != nil {
		log.Errorf(r.Context(), "failed to set locale: %v", err)
	}
	return whc
}

func (whc ApiWebhookContext) AppUserIntID() int64 {
	return whc.AppUserInt64ID()
}

func (whc ApiWebhookContext) AppUserData() (botsfwmodels.AppUserData, error) {
	//TODO implement me
	panic("implement me")
}

func (whc ApiWebhookContext) BotChatIntID() int64 {
	return whc.botChatID
}

func (whc ApiWebhookContext) ChatEntity() botsfwmodels.ChatData {
	return whc.chatEntity
}

func (whc ApiWebhookContext) GetAppUser() (botsfwmodels.AppUserData, error) {
	return whc.appUser, nil
}

func (whc ApiWebhookContext) Init(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (whc ApiWebhookContext) IsNewerThen(chatEntity botsfwmodels.ChatData) bool {
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

func (whc ApiWebhookContext) UpdateLastProcessed(chatEntity botsfwmodels.ChatData) error {
	panic("Not implemented")
}
