package dtb_general

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"fmt"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
)

const DELETE_ALL_COMMAND = "delete-all"

var DeleteAllCommand = bots.Command{
	Code:     DELETE_ALL_COMMAND,
	Icon:     emoji.MAIN_MENU_ICON,
	Commands: []string{"/deleteall"},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		botSettings := whc.GetBotSettings()
		if botSettings.Env != strongo.EnvLocal && botSettings.Env != strongo.EnvDevTest {
			return whc.NewMessage(fmt.Sprintf("This command supported just in development, got botSettings.Env: %v", botSettings.Env)), nil
		} else if botSettings.Env == strongo.EnvProduction {
			return whc.NewMessage("This command supported production environment"), nil
		}

		// We create success message ahead of actual operation as keyboard creation will fail once user deleted.
		m = whc.NewMessage("Deleted all records")
		SetMainMenuKeyboard(whc, &m)

		var chatID string
		if chatID, err = whc.BotChatID(); err != nil {
			return
		}

		if err = dal.Admin.DeleteAll(whc.Context(), botSettings.Code, chatID); err != nil {
			return
		}

		return
	},
}
