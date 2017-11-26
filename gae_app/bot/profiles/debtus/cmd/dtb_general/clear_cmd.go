package dtb_general

import (
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
)

const CLEAR_COMMAND = "clear"

var ClearCommand = bots.Command{
	Code:     CLEAR_COMMAND,
	Commands: trans.Commands(trans.COMMAND_CLEAR),
	//Title:    trans.COMMAND_TEXT_MAIN_MENU_TITLE,
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		log.Warningf(whc.Context(), "User called /clear command (not implemented yet)")
		return MainMenuAction(whc, whc.Translate(trans.MESSAGE_TEXT_NOT_IMPLEMENTED_YET), false)
	},
}
