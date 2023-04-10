package shared_all

import (
	"testing"
)

func TestAddSharedRoutes(t *testing.T) {
	router := botsfw.NewWebhookRouter(map[bots.WebhookInputType][]botsfw.Command{}, nil)
	AddSharedRoutes(router, BotParams{})
}
