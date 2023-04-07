package splitus

import (
	"fmt"
	"github.com/crediterra/money"
	"net/url"
	"strconv"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bytes"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/decimal"
	"github.com/strongo/log"
)

const groupCommandCode = "group"

var groupCommand = bots.NewCallbackCommand(groupCommandCode,
	func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		// we can't use GroupCallbackCommand as we have parameter id=[first|last|<id>]
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
			i             int
			userGroupJson models.UserGroupJson
		)
		switch id {
		case "first":
			i = 0
		case "last":
			i = len(groups) - 1
		default:
			userGroupJson.ID = id
			for j, g := range groups {
				if g.ID == userGroupJson.ID {
					i = j
				}
			}
		}
		userGroupJson = groups[i]

		do := query.Get("do")
		switch do {
		case "leave":
			if _, _, err = facade.Group.LeaveGroup(c, userGroupJson.ID, strconv.FormatInt(whc.AppUserIntID(), 10)); err != nil {
				if err == facade.ErrAttemptToLeaveUnsettledGroup {
					err = nil
					m.BotMessage = telegram.CallbackAnswer(tgbotapi.AnswerCallbackQueryConfig{Text: "Please settle group debts before leaving it."})
				}
				return
			}
			return groupsAction(whc, true, 0)
		}

		var group models.Group

		if group, err = dtdal.Group.GetGroupByID(c, userGroupJson.ID); err != nil {
			return
		}

		buf := new(bytes.Buffer)

		fmt.Fprintf(buf, "<b>Group #%d</b>: %v", i+1, userGroupJson.Name)
		var groupMemberJson models.GroupMemberJson
		if groupMemberJson, err = group.GetGroupMemberByUserID(strconv.FormatInt(whc.AppUserIntID(), 10)); err != nil {
			return
		}
		writeBalanceSide := func(title string, sign decimal.Decimal64p2, b money.Balance) {
			if len(b) > 0 {
				fmt.Fprintf(buf, "\n<b>%v</b>: ", title)
				if len(b) == 1 {
					for currency, value := range b {
						fmt.Fprintf(buf, "%v %v", sign*value, currency)
					}
				} else {
					for currency, value := range b {
						fmt.Fprintf(buf, "\n%v %v", sign*value, currency)
					}
				}
			}
		}
		writeBalanceSide("Owed to me", +1, groupMemberJson.Balance.OnlyPositive())
		writeBalanceSide("I owe", -1, groupMemberJson.Balance.OnlyNegative())
		fmt.Fprintf(buf, "\n<b>Members</b>: %v", group.MembersCount)

		m.Text = buf.String()

		m.IsEdit = true
		m.Format = bots.MessageFormatHTML
		tgKeyboard := tgbotapi.NewInlineKeyboardMarkup(groupsNavButtons(whc, groups, userGroupJson.ID))
		tgKeyboard.InlineKeyboard = append(tgKeyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{
			{
				Text:         whc.Translate("Leave group"),
				CallbackData: CallbackLink.ToGroup(groups[len(groups)-1].ID, true) + "&do=leave",
			},
		})
		m.Keyboard = tgKeyboard

		return
	},
)

func groupsNavButtons(translator strongo.SingleLocaleTranslator, groups []models.UserGroupJson, currentGroupID string) []tgbotapi.InlineKeyboardButton {
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
				CallbackData: groupsCommandCode + "?edit=1",
			})
		default:
			buttons = append(buttons, tgbotapi.InlineKeyboardButton{
				Text:         "⬅️",
				CallbackData: CallbackLink.ToGroup(groups[currentGroupIndex-1].ID, true),
			})
		}
	}
	if currentGroupID != "" {
		buttons = append(buttons, tgbotapi.InlineKeyboardButton{
			Text:         translator.Translate(trans.COMMAND_TEXT_GROUPS),
			CallbackData: groupsCommandCode + "?edit=1",
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
				CallbackData: groupsCommandCode + "?edit=1",
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
