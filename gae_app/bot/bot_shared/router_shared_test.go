package bot_shared

import (
	"testing"

	"github.com/strongo/bots-framework/core"
)

func TestAddSharedRoutes(t *testing.T) {
	router := bots.NewWebhookRouter(map[bots.WebhookInputType][]bots.Command{}, nil)
	AddSharedRoutes(router, BotParams{})
}
