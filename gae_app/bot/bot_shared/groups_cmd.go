package bot_shared

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bytes"
	"fmt"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app/log"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"net/url"
	"strconv"
)

const GROUPS_COMMAND = "groups"

func NewGroupTelegramInlineButton(whc bots.WebhookContext, groupsMessageID int) tgbotapi.InlineKeyboardButton {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "https://t.me/%v?startgroup=utm_s=%v__utm_m=%v__l=%v", whc.GetBotCode(), whc.GetBotCode(), "tgbot", whc.Locale().Code5)
	if groupsMessageID != 0 {
		buf.WriteString("__grpsMsgID=")
		buf.WriteString(strconv.Itoa(groupsMessageID))
	}
	return tgbotapi.InlineKeyboardButton{
		Text: whc.CommandText(trans.COMMAND_TEXT_ADD_GROUP, ""),
		URL:  buf.String(),
	}
}

var groupsCommand = bots.Command{
	Code:       GROUPS_COMMAND,
	InputTypes: []bots.WebhookInputType{bots.WebhookInputText, bots.WebhookInputCallbackQuery},
	Commands:   []string{"/groups"},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		return groupsAction(whc, false, 0)
	},
	CallbackAction: func(whc bots.WebhookContext, callbackURL *url.URL) (m bots.MessageFromBot, err error) {
		query := callbackURL.Query()
		isRefresh := query.Get("do") == "refresh"
		if m, err = groupsAction(whc, isRefresh || query.Get("edit") == "1", 0); err != nil {
			return
		}
		if isRefresh {
			c := whc.Context()
			log.Debugf(c, "do == 'refresh'")
			if m, err = bot.SendRefreshOrNothingChanged(whc, m); err != nil {
				return
			}
		}
		return
	},
}

func groupsAction(whc bots.WebhookContext, isEdit bool, groupsMessageID int) (m bots.MessageFromBot, err error) {
	if whc.IsInGroup() {
		m.Text = "This command supported just in private chat with @" + whc.GetBotCode()
		return
	}
	c := whc.Context()
	log.Debugf(c, "groupsAction(isEdit=%v, groupsMessageID=%d)", isEdit, groupsMessageID)
	buf := new(bytes.Buffer)

	fmt.Fprintf(buf, "<b>%v</b>\n\n", whc.Translate(trans.MESSAGE_TEXT_YOUR_BILL_SPLITTING_GROUPS))

	var user bots.BotAppUser
	if user, err = whc.GetAppUser(); err != nil {
		return
	}
	appUserEntity := user.(*models.AppUserEntity)

	groups := appUserEntity.ActiveGroups()

	{ // Filter groups known to bot or not linked to bot
		botCode := whc.GetBotCode()
		var j = 0
		for _, g := range groups {
			knownGroup := false
			if len(g.TgBots) == 0 {
				knownGroup = true
			} else {
				for _, tgBot := range g.TgBots {
					if tgBot == botCode {
						knownGroup = true
						break
					}
				}
			}
			if knownGroup {
				groups[j] = g
				j += 1
			}
		}
		groups = groups[:j]
	}

	if len(groups) == 0 {
		buf.WriteString(whc.Translate(trans.MESSAGE_TEXT_NO_GROUPS))
	} else {
		for i, group := range groups {
			fmt.Fprintf(buf, "  %d. %v\n", i+1, group.Name)
		}
	}

	m.Text = buf.String()

	tgKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{},
	)
	if len(groups) > 0 {
		tgKeyboard.InlineKeyboard = append(tgKeyboard.InlineKeyboard, groupsNavButtons(groups, ""))
	}

	if groupsMessageID == 0 {
		if isEdit {
			groupsMessageID = whc.Input().(telegram_bot.TelegramWebhookCallbackQuery).TgUpdate().CallbackQuery.Message.MessageID
		}
	} else {
		m.EditMessageUID = telegram_bot.ChatMessageUID{MessageID: groupsMessageID}
	}

	tgKeyboard.InlineKeyboard = append(tgKeyboard.InlineKeyboard,
		[]tgbotapi.InlineKeyboardButton{
			NewGroupTelegramInlineButton(whc, groupsMessageID),
		},
		[]tgbotapi.InlineKeyboardButton{
			{
				Text:         whc.Translate(trans.COMMAND_TEXT_REFRESH),
				CallbackData: GROUPS_COMMAND + "?do=refresh",
			},
		},
	)

	m.Keyboard = tgKeyboard
	m.IsEdit = isEdit
	m.Format = bots.MessageFormatHTML
	if !isEdit {
		var msg bots.OnMessageSentResponse
		if msg, err = whc.Responder().SendMessage(c, m, bots.BotApiSendMessageOverHTTPS); err != nil {

		}
		return groupsAction(whc, true, msg.TelegramMessage.(tgbotapi.Message).MessageID)
	}
	return
}
