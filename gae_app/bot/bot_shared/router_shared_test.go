package bot_shared

import (
	"github.com/strongo/bots-framework/core"
	"testing"
)

func TestAddSharedRoutes(t *testing.T) {
	router := bots.NewWebhookRouter(map[bots.WebhookInputType][]bots.Command{}, nil)
	AddSharedRoutes(router, BotParams{})
}
