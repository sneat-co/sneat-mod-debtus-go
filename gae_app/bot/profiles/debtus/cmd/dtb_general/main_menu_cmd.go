package dtb_general

import (
	"github.com/DebtsTracker/translations/emoji"
	"fmt"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app/log"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/fbm"
	"net/url"
)

const MAIN_MENU_COMMAND = "main-menu"

var MainMenuCommand = bots.Command{
	Code:     MAIN_MENU_COMMAND,
	Icon:     emoji.MAIN_MENU_ICON,
	Commands: trans.Commands(trans.COMMAND_MENU, emoji.MAIN_MENU_ICON),
	Title:    trans.COMMAND_TEXT_MAIN_MENU_TITLE,
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		return MainMenuAction(whc, "", true)
	},
	CallbackAction: func(whc bots.WebhookContext, callbackURL *url.URL) (m bots.MessageFromBot, err error) {
		return MainMenuAction(whc, "", true)
	},
}

func MainMenuAction(whc bots.WebhookContext, messageText string, showHint bool) (bots.MessageFromBot, error) {
	if messageText == "" {
		if whc.BotPlatform().Id() != fbm_bot.FbmPlatformID {
			if showHint {
				messageText = fmt.Sprintf("%v\n%v", whc.Translate(trans.MESSAGE_TEXT_WHATS_NEXT), whc.Translate(trans.MESSAGE_TEXT_WHATS_NEXT_HINT))
			} else {
				messageText = whc.Translate(trans.MESSAGE_TEXT_WHATS_NEXT)
			}
		}
	}
	log.Infof(whc.Context(), "MainMenuCommand.Action()")
	whc.ChatEntity().SetAwaitingReplyTo("")
	m := whc.NewMessage(messageText)
	SetMainMenuKeyboard(whc, &m)
	return m, nil
}
