package dtb_retention

import (
	"bitbucket.com/debtstracker/gae_app/bot/profiles/debtus/cmd/dtb_general"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-framework/core"
)

var DeleteUserCommand = bots.Command{
	Code:     "delete-user",
	Icon:     emoji.NO_ENTRY_SIGN_ICON,
	Commands: []string{"/deleteuser"},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		err = bots.SetAccessGranted(whc, false)
		if err != nil {
			m = whc.NewMessageByCode(trans.MESSAGE_TEXT_FAILED_TO_DELETE_USER, err)
			dtb_general.SetMainMenuKeyboard(whc, &m)
			return m, nil
		} else {
			m = whc.NewMessageByCode(trans.MESSAGE_TEXT_USER_DELETED)
			return m, nil
		}
	},
}
