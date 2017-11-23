package bot_shared

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/strongo/log"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"net/url"
	"strconv"
)

const GROUP_COMMAND = "group"

var groupCommand = bots.NewCallbackCommand(GROUP_COMMAND,
	func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "groupCommand.CallbackAction()")

		var user bots.BotAppUser
		if user, err = whc.GetAppUser(); err != nil {
			return
		}
		appUserEntity := user.(*models.AppUserEntity) // TODO: Create shortcut function

		groups := appUserEntity.ActiveGroups()

		if len(groups) == 0 {
			return groupsAction(whc, true, 0)
		}

		query := callbackUrl.Query()

		id := query.Get("id")

		var (
			i     int
			group models.UserGroupJson
		)
		switch id {
		case "first":
			i = 0
		case "last":
			i = len(groups) - 1
		default:
			group.ID = id
			for j, g := range groups {
				if g.ID == group.ID {
					i = j
				}
			}
		}
		group = groups[i]

		do := query.Get("do")
		switch do {
		case "leave":
			if _, _, err = facade.Group.LeaveGroup(c, group.ID, strconv.FormatInt(whc.AppUserIntID(), 10)); err != nil {
				return
			}
			return groupsAction(whc, true, 0)
		}

		m.Text = fmt.Sprintf("<b>Group #%d</b>: %v", i+1, group.Name)
		m.IsEdit = true
		m.Format = bots.MessageFormatHTML
		tgKeyboard := tgbotapi.NewInlineKeyboardMarkup(groupsNavButtons(groups, group.ID))
		tgKeyboard.InlineKeyboard = append(tgKeyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{
			{
				Text:         "Leave group",
				CallbackData: CallbackLink.ToGroup(groups[len(groups)-1].ID, true) + "&do=leave",
			},
		})
		tgKeyboard.InlineKeyboard = append(tgKeyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{NewGroupTelegramInlineButton(whc, 0)})
		m.Keyboard = tgKeyboard

		return
	},
)

func groupsNavButtons(groups []models.UserGroupJson, currentGroupID string) []tgbotapi.InlineKeyboardButton {
	var currentGroupIndex = -1
	if currentGroupID != "" {

		for i, group := range groups {
			if group.ID == currentGroupID {
				currentGroupIndex = i
				break
			}
		}
	}
	buttons := []tgbotapi.InlineKeyboardButton{}
	if len(groups) > 0 || currentGroupIndex < 0 {
		switch currentGroupIndex {
		case -1:
			buttons = append(buttons, tgbotapi.InlineKeyboardButton{
				Text:         "⬅️",
				CallbackData: CallbackLink.ToGroup(groups[len(groups)-1].ID, true),
			})
		case 0:
			buttons = append(buttons, tgbotapi.InlineKeyboardButton{
				Text:         "⬅️",
				CallbackData: GROUPS_COMMAND + "?edit=1",
			})
		default:
			buttons = append(buttons, tgbotapi.InlineKeyboardButton{
				Text:         "⬅️",
				CallbackData: CallbackLink.ToGroup(groups[currentGroupIndex-1].ID, true),
			})
		}
	}
	if currentGroupID != "" && len(groups) != 1 {
		buttons = append(buttons, tgbotapi.InlineKeyboardButton{
			Text:         "📇",
			CallbackData: GROUPS_COMMAND + "?edit=1",
		})

	}
	if len(groups) > 0 || currentGroupIndex < 0 {
		switch currentGroupIndex {
		case -1:
			buttons = append(buttons, tgbotapi.InlineKeyboardButton{
				Text:         "➡️",
				CallbackData: CallbackLink.ToGroup(groups[0].ID, true),
			})
		case len(groups) - 1:
			buttons = append(buttons, tgbotapi.InlineKeyboardButton{
				Text:         "➡️",
				CallbackData: GROUPS_COMMAND + "?edit=1",
			})
		default:
			buttons = append(buttons, tgbotapi.InlineKeyboardButton{
				Text:         "➡️",
				CallbackData: CallbackLink.ToGroup(groups[currentGroupIndex+1].ID, true),
			})
		}
	}
	return buttons
}
