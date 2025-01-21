package dtdal

import (
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/models4debtus"
	"strconv"
)

type TransferSourceBot struct {
	platform string
	botID    string
	chatID   string
}

func (s TransferSourceBot) PopulateTransfer(t *models4debtus.TransferData) {
	t.CreatedOnPlatform = s.platform
	t.CreatedOnID = s.botID
	if s.platform == "telegram" {
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
