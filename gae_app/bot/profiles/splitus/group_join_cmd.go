package splitus

import (
	"fmt"
	"net/url"
	"strconv"

	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/shared_group"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/log"
)

const joinGroupCommanCode = "join-group"

var joinGroupCommand = shared_group.GroupCallbackCommand(joinGroupCommanCode,
	func(whc bots.WebhookContext, callbackUrl *url.URL, group models.Group) (m bots.MessageFromBot, err error) {
		c := whc.Context()

		userID := whc.AppUserStrID()
		var appUser models.AppUser
		if group.UserIsMember(userID) {
			if appUser, err = dtdal.User.GetUserByStrID(c, userID); err != nil {
				return
			}
			whc.LogRequest()
			callbackAnswer := tgbotapi.NewCallback("", whc.Translate(trans.ALERT_TEXT_YOU_ARE_ALREADY_MEMBER_OF_THE_GROUP))
			callbackAnswer.ShowAlert = true
			m.BotMessage = telegram.CallbackAnswer(callbackAnswer)
		} else {
			err = dtdal.DB.RunInTransaction(c, func(c context.Context) error {
				if appUser, err = dtdal.User.GetUserByStrID(c, userID); err != nil {
					return err
				}
				_, changed, memberIndex, member, members := group.AddOrGetMember(userID, "", appUser.FullName())
				tgUserID := strconv.FormatInt(int64(whc.Input().GetSender().GetID().(int)), 10)
				if member.TgUserID == "" {
					member.TgUserID = tgUserID
					changed = true
				} else {
					if tgUserID != member.TgUserID {
						log.Errorf(c, "tgUserID:%v != member.TgUserID:%v", tgUserID, member.TgUserID)
					}
				}
				switch group.GetSplitMode() {
				case models.SplitModeEqually:
					var shares int
					if group.MembersCount > 0 {
						shares = group.GetGroupMembers()[0].Shares
					} else {
						shares = 1
					}
					if member.Shares != shares {
						member.Shares = shares
						changed = true
					}
				case models.SplitModeShare:
					if member.Shares != 0 {
						member.Shares = 0
						changed = true
					}
				}
				if changed {
					members[memberIndex] = member
					group.SetGroupMembers(members)
					if err = dtdal.Group.SaveGroup(c, group); err != nil {
						return err
					}
				} else {
					log.Debugf(c, "Group member not changed")
				}
				if userChanged := appUser.AddGroup(group, whc.GetBotCode()); userChanged {
					if err = facade.User.SaveUser(c, appUser); err != nil {
						return err
					}
				}
				if len(members) > 1 {
					groupUsersCount := 0
					for _, m := range members {
						if m.UserID != "" {
							groupUsersCount += 1
						}
					}
					if groupUsersCount > 1 {
						if err = facade.Group.DelayUpdateGroupUsers(c, group.ID); err != nil {
							return err
						}
					}
				}
				return err
			}, dtdal.CrossGroupTransaction)

			if m, err := showGroupMembers(whc, group, true); err != nil {
				return m, err
			} else if _, err = whc.Responder().SendMessage(c, m, bots.BotAPISendMessageOverHTTPS); err != nil {
				return m, err
			}

			m.Text = whc.Translate(trans.MESSAGE_TEXT_USER_JOINED_GROUP, fmt.Sprintf(`<a href="tg://user?id=%v">%v</a>`, whc.MustBotChatID(), appUser.FullName()))
		}

		m.Format = bots.MessageFormatHTML
		return
	},
)
