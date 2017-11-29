package debtus

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/bot_shared"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_admin"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_fbm"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_general"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_invite"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_retention"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_settings"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_splitbill"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_transfer"
	"github.com/strongo/bots-framework/core"
)

var textAndContactCommands = []bots.Command{ // TODO: Check for Action || CallbackAction and register accordingly.
	//OnboardingAskInviteChannelCommand, // We need it as otherwise do not handle replies.
	//SetPreferredLanguageCommand,
	//OnboardingAskInviteCodeCommand,
	//OnboardingCheckInviteCommand,
	//
	dtb_general.FeedbackCommand,
	dtb_general.FeedbackTextCommand,
	dtb_general.DeleteAllCommand,
	dtb_general.BetaCommand,
	//
	dtb_admin.AdminCommand,
	//
	dtb_settings.StartCommand,
	dtb_settings.LoginPinCommand,
	dtb_settings.OnboardingAskLocaleCommand,
	dtb_settings.OnboardingTellAboutInviteCodeCommand, // We need it as otherwise do not handle replies. Consider incorporate to StartCommand?
	//
	dtb_general.Login2WebCommand,
	dtb_general.MainMenuCommand,
	dtb_general.ClearCommand,
	dtb_general.HelpCommand,
	dtb_general.AdsCommand,
	//
	dtb_transfer.StartLendingWizardCommand,
	dtb_transfer.StartBorrowingWizardCommand,
	dtb_transfer.StartReturnWizardCommand,
	dtb_transfer.BalanceCommand,
	dtb_transfer.HistoryCommand,
	dtb_transfer.CancelTransferWizardCommand,
	dtb_transfer.ParseTransferCommand,
	dtb_transfer.AskHowMuchHaveBeenReturnedCommand,
	dtb_transfer.SetNextReminderDateCallbackCommand,
	//
	dtb_settings.SettingsCommand,
	dtb_settings.FixBalanceCommand,
	dtb_settings.ContactsListCommand,
	dtb_settings.AboutDrawCommand,
	//dtb_settings.AskCurrencySettingsCommand,
	//
	dtb_retention.DeleteUserCommand,
	//
	dtb_invite.InviteCommand,
	dtb_transfer.AskEmailForReceiptCommand,       // TODO: Should it be in dtb_transfer?
	dtb_transfer.AskPhoneNumberForReceiptCommand, // TODO: Should it be in dtb_transfer?
	dtb_invite.CreateMassInviteCommand,
	//
	bot_shared.ReferrersCommand,
}

var callbackCommands = []bots.Command{
	dtb_general.MainMenuCommand,
	dtb_general.PleaseWaitCommand,
	//dtb_invite.InviteCommand,
	//
	dtb_settings.AskPreferredLocaleFromSettingsCallback,
	dtb_settings.SetLocaleCallbackCommand,
	dtb_settings.BackToSettingsCallbackCommand,
	dtb_settings.ContactsListCommand,
	dtb_settings.AboutDrawCommand,
	dtb_settings.JoinDrawCommand,
	//
	dtb_fbm.FbmGetStartedCommand, // TODO: Move command to other package?
	dtb_fbm.FbmMainMenuCommand,
	dtb_fbm.FbmDebtsCommand,
	dtb_fbm.FbmBillsCommand,
	dtb_fbm.FbmSettingsCommand,
	//
	dtb_invite.CreateMassInviteCommand,
	dtb_invite.AskInviteAddressCallbackCommand,
	//
	dtb_transfer.CreateReceiptIfNoInlineNotificationCommand,
	dtb_transfer.SendReceiptCallbackCommand,
	//dtb_transfer.AcknowledgeReceiptCommand,
	dtb_transfer.ViewReceiptInTelegramCallbackCommand,
	dtb_transfer.ChangeReceiptAnnouncementLangCommand,
	dtb_transfer.ViewReceiptCallbackCommand,
	dtb_transfer.AcknowledgeReceiptCallbackCommand,
	dtb_transfer.TransferHistoryCallbackCommand,
	dtb_transfer.AskForInterestAndCommentCallbackCommand,
	dtb_transfer.BalanceCallbackCommand,
	dtb_transfer.DueReturnsCallbackCommand,
	dtb_transfer.ReturnCallbackCommand,
	dtb_transfer.EnableReminderAgainCallbackCommand,
	dtb_transfer.SetNextReminderDateCallbackCommand,
	//dtb_transfer.CounterpartyNoTelegramCommand,
	dtb_transfer.RemindAgainCallbackCommand,
	//dtb_general.FeedbackCallbackCommand,
	dtb_general.FeedbackCommand,
	dtb_general.CanYouRateCommand,
	dtb_general.FeedbackTextCommand,
	bot_shared.AddReferrerCommand,
}

var Router bots.WebhooksRouter = bots.NewWebhookRouter(
	map[bots.WebhookInputType][]bots.Command{
		bots.WebhookInputText:          textAndContactCommands,
		bots.WebhookInputContact:       textAndContactCommands,
		bots.WebhookInputCallbackQuery: callbackCommands,
		//
		bots.WebhookInputReferral: {
			dtb_settings.StartCommand,
		},
		bots.WebhookInputSticker: {
			bots.IgnoreCommand,
		},
		bots.WebhookInputConversationStarted: {
			dtb_settings.StartCommand,
		},
		bots.WebhookInputInlineQuery: {
			InlineQueryCommand,
		},
		bots.WebhookInputChosenInlineResult: {
			dtb_invite.ChosenInlineResultCommand,
		},
		bots.WebhookInputNewChatMembers: {
			dtb_splitbill.NewChatMembersCommand,
		},
	},
	func() string { return "Please report any errors to @DebtsTrackerGroup" },
)
