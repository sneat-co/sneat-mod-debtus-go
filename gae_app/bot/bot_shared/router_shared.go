package bot_shared

import (
	"github.com/strongo/bots-framework/core"
)

func AddSharedRoutes(router bots.WebhooksRouter, botParams BotParams) {
	router.AddCommands(bots.WebhookInputText, []bots.Command{
		startCommand(botParams),
		helpRootCommand,
		setBillDueDateCommand,
		groupsCommand,
		settingsCommand,
		ReferrersCommand,
	})
	router.AddCommands(bots.WebhookInputCallbackQuery, []bots.Command{
		onStartCallbackCommand(botParams),
		settleGroupAskForCounterpartyCommand,
		settleGroupCounterpartyChoosenCommand,
		settleGroupCounterpartyConfirmedCommand,
		joinGroupCommand(botParams),
		billCardCommand(botParams),
		setBillCurrencyCommand(botParams),
		groupSettingsSetCurrencyCommand(botParams),
		helpRootCommand,
		groupsCommand,
		groupCommand,
		leaveGroupCommand,
		billMembersCommand,
		inviteToBillCommand,
		setBillDueDateCommand,
		changeBillTotalCommand,
		addBillComment,
		settingsCommand,
		groupSettingsChooseCurrencyCommand,
		groupMembersCommand,
	})
	router.AddCommands(bots.WebhookInputNewChatMembers, []bots.Command{
		newChatMembersCommand,
	})
	router.AddCommands(bots.WebhookInputLeftChatMembers, []bots.Command{
		leftChatCommand,
	})
	router.AddCommands(bots.WebhookInputInlineQuery, []bots.Command{
		inlineQueryCommand,
	})
	router.AddCommands(bots.WebhookInputChosenInlineResult, []bots.Command{
		choosenInlineResultHandler(botParams),
	})
	router.AddCommands(bots.WebhookInputSticker, []bots.Command{
		bots.IgnoreCommand,
	})
}
