package dtb_general

import (
	"fmt"
	"net/url"

	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/fbm"
	"github.com/strongo/log"
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
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		return MainMenuAction(whc, "", true)
	},
}

func MainMenuAction(whc bots.WebhookContext, messageText string, showHint bool) (bots.MessageFromBot, error) {
	if messageText == "" {
		if whc.BotPlatform().ID() != fbm.PlatformID {
			if showHint {
				messageText = fmt.Sprintf("%v\n\n%v", whc.Translate(trans.MESSAGE_TEXT_WHATS_NEXT), whc.Translate(trans.MESSAGE_TEXT_DEBTUS_COMMANDS))
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
