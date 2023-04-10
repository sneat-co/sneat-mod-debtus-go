package shared_all

func AddSharedRoutes(router botsfw.WebhooksRouter, botParams BotParams) {
	startCommand := createStartCommand(botParams)
	helpRootCommand := createHelpRootCommand(botParams)
	router.AddCommands(bots.WebhookInputText, []botsfw.Command{
		startCommand,
		helpRootCommand,
		ReferrersCommand,
		createOnboardingAskLocaleCommand(botParams),
		aboutDrawCommand,
	})
	router.AddCommands(bots.WebhookInputCallbackQuery, []botsfw.Command{
		onStartCallbackCommand(botParams),
		helpRootCommand,
		joinDrawCommand,
		aboutDrawCommand,
		askPreferredLocaleFromSettingsCallback,
		setLocaleCallbackCommand(botParams),
	})
	router.AddCommands(bots.WebhookInputLeftChatMembers, []botsfw.Command{
		leftChatCommand,
	})
	router.AddCommands(bots.WebhookInputSticker, []botsfw.Command{
		bots.IgnoreCommand,
	})
	router.AddCommands(bots.WebhookInputReferral, []botsfw.Command{
		startCommand,
	})
	router.AddCommands(bots.WebhookInputConversationStarted, []botsfw.Command{
		startCommand,
	})
}
