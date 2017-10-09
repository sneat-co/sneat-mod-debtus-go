package splitus

import (
	"github.com/strongo/bots-framework/core"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/bot_shared"
	"github.com/strongo/app"
	"bytes"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-api-telegram"
	"github.com/DebtsTracker/translations/emoji"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
)

var botParams = bot_shared.BotParams{
	GetGroupBillCardInlineKeyboard:   getGroupBillCardInlineKeyboard,
	GetPrivateBillCardInlineKeyboard: getPrivateBillCardInlineKeyboard,
	DelayUpdateBillCardOnUserJoin:    delayUpdateBillCardOnUserJoin,
	OnAfterBillCurrencySelected:      getWhoPaidInlineKeyboard,
	ShowGroupMembers:                 showGroupMembers,
	WelcomeText: func(translator strongo.SingleLocaleTranslator, buf *bytes.Buffer) {
		buf.WriteString(translator.Translate(trans.SPLITUS_TEXT_HI))
		buf.WriteString("\n\n")
		buf.WriteString(translator.Translate(trans.SPLITUS_TEXT_ABOUT_ME_AND_CO))
	},
	InGroupWelcomeMessage: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		return whc.NewEditMessage(whc.Translate(trans.MESSAGE_TEXT_HI)+
			"\n\n"+ whc.Translate(trans.SPLITUS_TEXT_HI_IN_GROUP)+
			"\n\n"+ whc.Translate(trans.SPLITUS_TEXT_ABOUT_ME_AND_CO),
			bots.MessageFormatHTML)
	},
	InBotWelcomeMessage: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		var user *models.AppUserEntity
		if user, err = bot_shared.GetUser(whc); err != nil {
			return
		}
		m.Text = whc.Translate(
			trans.MESSAGE_TEXT_HI_USERNAME, user.FirstName) + " " + whc.Translate(trans.SPLITUS_TEXT_HI) +
			"\n\n" + whc.Translate(trans.SPLITUS_TEXT_ABOUT_ME_AND_CO) +
			"\n\n" + whc.Translate(trans.SPLITUS_TG_COMMANDS)
		m.Format = bots.MessageFormatHTML
		m.IsEdit = true

		m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
			[]tgbotapi.InlineKeyboardButton{
				//{
				//	Text:         emoji.CLIPBOARD_ICON + " Bills",
				//	CallbackData: "bills",
				//},
				tgbotapi.NewInlineKeyboardButtonSwitchInlineQuery(
					whc.CommandText(trans.COMMAND_TEXT_NEW_BILL, emoji.MEMO_ICON),
					"",
				),
			},
			[]tgbotapi.InlineKeyboardButton{
				//{
				//	Text:         emoji.CONTACTS_ICON + " Groups",
				//	CallbackData: "groups",
				//},
				bot_shared.NewGroupTelegramInlineButton(whc, 0),
			},
		)
		return
	},
}

var Router bots.WebhooksRouter = bots.NewWebhookRouter(
	map[bots.WebhookInputType][]bots.Command{
		bots.WebhookInputText: {
			bot_shared.EditedBillCardHookCommand,
			billsCommand,
		},
		bots.WebhookInputCallbackQuery: {
			bot_shared.JoinBillCommand(botParams),
			bot_shared.CloseBillCommand(botParams),
			bot_shared.EditBillCommand(botParams),
			bot_shared.NewBillCommand(botParams),
			groupMembersCommand,
			billsCommand,
			billSplitCommand,
			billSplitModesListCommand,
			finalizeBillCommand,
			billChangeSplitModeCommand,
			changeBillPayerCommand,
			groupSplitCommand,
		},
	},
	func() string { return "Please report any errors to @SplitusGroup" },
)

func init() {
	bot_shared.AddSharedRoutes(Router, botParams)
}

func getWhoPaidInlineKeyboard(translator strongo.SingleLocaleTranslator, billID string) *tgbotapi.InlineKeyboardMarkup {
	callbackDataPrefix := bot_shared.BillCallbackCommandData(bot_shared.JOIN_BILL_COMMAND, billID)
	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{{Text: "✋ " + translator.Translate(trans.BUTTON_TEXT_I_PAID_FOR_THE_BILL), CallbackData: callbackDataPrefix + "&i=paid"}},
			{{Text: "🙏 " + translator.Translate(trans.BUTTON_TEXT_I_OWE_FOR_THE_BILL), CallbackData: callbackDataPrefix + "&i=owe"}},
			{{Text: "🚫 " + translator.Translate(trans.BUTTON_TEXT_I_DO_NOT_SHARE_THIS_BILL), CallbackData: bot_shared.BillCallbackCommandData(bot_shared.LEAVE_BILL_COMMAND, billID)}},
		},
	}
}

