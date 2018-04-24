package debtus

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_admin"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_fbm"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_general"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_invite"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_retention"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_settings"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_transfer"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/shared_all"
	"github.com/strongo/bots-framework/core"
)

var botParams = shared_all.BotParams{
	//GetGroupBillCardInlineKeyboard:   getGroupBillCardInlineKeyboard,
	//GetPrivateBillCardInlineKeyboard: getPrivateBillCardInlineKeyboard,
	//DelayUpdateBillCardOnUserJoin:    delayUpdateBillCardOnUserJoin,
	//OnAfterBillCurrencySelected:      getWhoPaidInlineKeyboard,
	//ShowGroupMembers:                 showGroupMembers,
	HelpCommandAction: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		return dtb_general.HelpCommandAction(whc, true)
	},
	//InGroupWelcomeMessage: func(whc bots.WebhookContext, group models.Group) (m bots.MessageFromBot, err error) {
	//	m, err = shared_all.GroupSettingsAction(whc, group, false)
	//	if err != nil {
	//		return
	//	}
	//	if _, err = whc.Responder().SendMessage(whc.Context(), m, bots.BotApiSendMessageOverHTTPS); err != nil {
	//		return
	//	}
	//
	//	return whc.NewEditMessage(whc.Translate(trans.MESSAGE_TEXT_HI)+
	//		"\n\n"+ whc.Translate(trans.SPLITUS_TEXT_HI_IN_GROUP)+
	//		"\n\n"+ whc.Translate(trans.SPLITUS_TEXT_ABOUT_ME_AND_CO),
	//		bots.MessageFormatHTML)
	//},
	InBotWelcomeMessage: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		m.Text = "Hi there"
		m.Format = bots.MessageFormatHTML
		//m.IsEdit = true
		return
	},
	//
	//
	//
	StartInBotAction: dtb_settings.StartInBotAction,
	SetMainMenu:      dtb_general.SetMainMenuKeyboard,
}

func init() {
	shared_all.AddSharedRoutes(Router, botParams)
}

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
	dtb_settings.SettingsCommand,
	dtb_settings.LoginPinCommand,
	//dtb_settings.OnboardingTellAboutInviteCodeCommand, // We need it as otherwise do not handle replies. Consider incorporate to StartCommand?
	dtb_settings.FixBalanceCommand,
	dtb_settings.ContactsListCommand,
	//
	//dtb_settings.AskCurrencySettingsCommand,
	//
	dtb_general.Login2WebCommand,
	dtb_general.MainMenuCommand,
	dtb_general.ClearCommand,
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
	dtb_retention.DeleteUserCommand,
	//
	dtb_invite.InviteCommand,
	dtb_transfer.AskEmailForReceiptCommand,       // TODO: Should it be in dtb_transfer?
	dtb_transfer.AskPhoneNumberForReceiptCommand, // TODO: Should it be in dtb_transfer?
	dtb_invite.CreateMassInviteCommand,
	//
}

var callbackCommands = []bots.Command{
	dtb_general.MainMenuCommand,
	dtb_general.PleaseWaitCommand,
	//dtb_invite.InviteCommand,
	//
	dtb_settings.SettingsCommand,
	dtb_settings.ContactsListCommand,
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
	shared_all.AddReferrerCommand,
}

var Router = bots.NewWebhookRouter(
	map[bots.WebhookInputType][]bots.Command{
		bots.WebhookInputText:          textAndContactCommands,
		bots.WebhookInputContact:       textAndContactCommands,
		bots.WebhookInputCallbackQuery: callbackCommands,
		//
		bots.WebhookInputInlineQuery: {
			InlineQueryCommand,
		},
		bots.WebhookInputChosenInlineResult: {
			dtb_invite.ChosenInlineResultCommand,
		},
		bots.WebhookInputNewChatMembers: {
			newChatMembersCommand,
		},
	},
	func() string { return "Please report any errors to @DebtsTrackerGroup" },
)
