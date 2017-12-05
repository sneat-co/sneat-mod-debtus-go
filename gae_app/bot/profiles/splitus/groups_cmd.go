package splitus

import (
	"bytes"
	"fmt"
	"net/url"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/log"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/shared_group"
	"github.com/DebtsTracker/translations/emoji"
)

const groupsCommandCode = "groups"

var groupsCommand = bots.Command{
	Code:       groupsCommandCode,
	InputTypes: []bots.WebhookInputType{bots.WebhookInputText, bots.WebhookInputCallbackQuery},
	Commands:   trans.Commands(trans.COMMAND_TEXT_GROUPS, emoji.MAN_AND_WOMAN, "/" + groupsCommandCode),
	Icon: emoji.MAN_AND_WOMAN,
	Title: trans.COMMAND_TEXT_GROUPS,
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		return groupsAction(whc, false, 0)
	},
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		query := callbackUrl.Query()
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

		fmt.Fprint(buf, "\n", whc.Translate(trans.MESSAGE_TEXT_USE_ARROWS_TO_SELECT_GROUP))
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
			shared_group.NewGroupTelegramInlineButton(whc, groupsMessageID),
		},
		[]tgbotapi.InlineKeyboardButton{
			{
				Text:         whc.Translate(trans.COMMAND_TEXT_REFRESH),
				CallbackData: groupsCommandCode + "?do=refresh",
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
