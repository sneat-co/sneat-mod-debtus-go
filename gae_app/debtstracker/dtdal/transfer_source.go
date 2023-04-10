package dtdal

import (
	"strconv"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/bots-go-framework/bots-fw-telegram"
)

type TransferSourceBot struct {
	platform string
	botID    string
	chatID   string
}

func (s TransferSourceBot) PopulateTransfer(t *models.TransferEntity) {
	t.CreatedOnPlatform = s.platform
	t.CreatedOnID = s.botID
	if s.platform == telegram.PlatformID {
		t.Creator().TgBotID = s.botID
		var err error
		t.Creator().TgChatID, err = strconv.ParseInt(s.chatID, 10, 64)
		if err != nil {
			panic(err.Error())
		}
	}
}

var _ TransferSource = (*TransferSourceBot)(nil)

func NewTransferSourceBot(platform, botID, chatID string) TransferSourceBot {
	if botID == "" {
		panic("botID is not set")
	}
	if chatID == "" {
		panic("chatID is not set")
	}
	return TransferSourceBot{
		platform: platform,
		botID:    botID,
		chatID:   chatID,
	}
}
