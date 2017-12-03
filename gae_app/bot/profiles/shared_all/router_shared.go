package shared_all

import (
	"github.com/strongo/bots-framework/core"
)

func AddSharedRoutes(router bots.WebhooksRouter, botParams BotParams) {
	startCommand := createStartCommand(botParams)
	helpRootCommand := createHelpRootCommand(botParams)
	router.AddCommands(bots.WebhookInputText, []bots.Command{
		startCommand,
		helpRootCommand,
		ReferrersCommand,
		onboardingAskLocaleCommand,
		aboutDrawCommand,
	})
	router.AddCommands(bots.WebhookInputCallbackQuery, []bots.Command{
		onStartCallbackCommand(botParams),
		helpRootCommand,
		joinDrawCommand,
		aboutDrawCommand,
		askPreferredLocaleFromSettingsCallback,
		setLocaleCallbackCommand(botParams),
	})
	router.AddCommands(bots.WebhookInputLeftChatMembers, []bots.Command{
		leftChatCommand,
	})
	router.AddCommands(bots.WebhookInputSticker, []bots.Command{
		bots.IgnoreCommand,
	})
	router.AddCommands(bots.WebhookInputReferral, []bots.Command{
		startCommand,
	})
	router.AddCommands(bots.WebhookInputConversationStarted, []bots.Command{
		startCommand,
	})
}
