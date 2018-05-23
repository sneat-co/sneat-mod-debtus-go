package shared_all

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	"bitbucket.org/asterus/debtstracker-server/gae_app/bot"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

const HELP_COMMAND = "help"

func createHelpRootCommand(params BotParams) bots.Command {
	return bots.Command{
		Code:     HELP_COMMAND,
		Commands: []string{"/help", emoji.HELP_ICON},
		Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
			switch whc.GetBotSettings().Profile {
			case bot.ProfileDebtus:
				return params.HelpCommandAction(whc)
			}
			return helpRootAction(whc, false)
		},
		CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
			_ = whc.ChatEntity()
			q := callbackUrl.Query().Get("q")
			switch q {
			case "":
				m, err = helpRootAction(whc, true)
			case trans.HELP_HOW_TO_CREATE_BILL_Q:
				m, err = helpHowToCreateNewBill(whc)
			}
			m.Format = bots.MessageFormatHTML
			m.IsEdit = true
			return
		},
	}
}

func helpRootAction(whc bots.WebhookContext, isCallback bool) (m bots.MessageFromBot, err error) {
	m.Text = whc.Translate(trans.MESSAGE_TEXT_HELP_ROOT, strings.Replace(whc.GetBotCode(), "Bot", "Group", 1))
	m.Format = bots.MessageFormatHTML
	m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{{
			Text:         whc.Translate(trans.HELP_HOW_TO_CREATE_BILL_Q),
			CallbackData: "help?q=" + url.QueryEscape(trans.HELP_HOW_TO_CREATE_BILL_Q),
		}},
	)
	if isCallback {
		m.IsEdit = true
	}
	return
}

func helpHowToCreateNewBill(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
	var buffer bytes.Buffer
	if err = common.HtmlTemplates.RenderTemplate(whc.Context(), &buffer, whc, trans.HELP_HOW_TO_CREATE_BILL_A, struct{ BotCode string }{whc.GetBotCode()}); err != nil {
		return
	}
	m.Text = fmt.Sprintf("<b>%v</b>", whc.Translate(trans.HELP_HOW_TO_CREATE_BILL_Q)) +
		"\n\n" + buffer.String()
	m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			{
				Text: emoji.CONTACTS_ICON + " Split bill in Telegram Group",
				URL:  "https://t.me/%v?startgroup=new-bill",
			},
		},
		[]tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonSwitchInlineQuery(
				emoji.ROCKET_ICON+" Split bill with Telegram user(s)",
				"",
			),
		},
		[]tgbotapi.InlineKeyboardButton{
			{
				Text:         emoji.CLIPBOARD_ICON + "New bill manually",
				CallbackData: "new-bill-manually",
			},
		},
		[]tgbotapi.InlineKeyboardButton{
			{Text: emoji.RETURN_BACK_ICON + " " + whc.Translate(trans.MESSAGE_TEXT_HELP_BACK_TO_ROOT),
				CallbackData: "help",
			},
		},
	)
	return
}
