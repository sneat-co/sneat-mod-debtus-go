package bot_shared

import (
	"github.com/strongo/bots-framework/core"
	"net/url"
	"github.com/strongo/bots-api-telegram"
	"fmt"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/trans"
	"golang.org/x/net/context"
	"bitbucket.com/debtstracker/gae_app/debtstracker/dal"
	"bitbucket.com/debtstracker/gae_app/debtstracker/facade"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/app/log"
)

const JOIN_GROUP_COMMAND = "join-group"

func joinGroupCommand(params BotParams) bots.Command {
	return GroupCallbackCommand(JOIN_GROUP_COMMAND,
		func(whc bots.WebhookContext, callbackURL *url.URL, group models.Group) (m bots.MessageFromBot, err error) {
			c := whc.Context()

			if group, err = GetGroup(whc); err != nil {
				return
			}

			userID := whc.AppUserIntID()
			var appUser models.AppUser
			if group.UserIsMember(userID) {
				if appUser, err = dal.User.GetUserByID(c, userID); err != nil {
					return
				}
				whc.LogRequest()
				callbackAnswer := tgbotapi.NewCallback("", whc.Translate(trans.ALERT_TEXT_YOU_ARE_ALREADY_MEMBER_OF_THE_GROUP))
				callbackAnswer.ShowAlert = true
				m.BotMessage = telegram_bot.CallbackAnswer(callbackAnswer)
			} else {
				err = dal.DB.RunInTransaction(c, func(c context.Context) error {
					if appUser, err = dal.User.GetUserByID(c, userID); err != nil {
						return err
					}
					_, changed, memberIndex, member, members := group.AddOrGetMember(userID, 0, appUser.FullName())
					tgUserID := int64(whc.Input().GetSender().GetID().(int))
					if member.TgUserID == 0 {
						member.TgUserID = tgUserID
						changed = true
					} else {

						if tgUserID != member.TgUserID {
							log.Errorf(c, "tgUserID:%d != member.TgUserID:%d", tgUserID, member.TgUserID)
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
						if err = dal.Group.SaveGroup(c, group); err != nil {
							return err
						}
					} else {
						log.Debugf(c, "Group member not changed")
					}
					if userChanged := appUser.AddGroup(group, whc.GetBotCode()); userChanged {
						if err = dal.User.SaveUser(c, appUser); err != nil {
							return err
						}
					}
					if len(members) > 1 {
						groupUsersCount := 0
						for _, m := range members {
							if m.UserID != 0 {
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
				}, dal.CrossGroupTransaction)

				if m, err := params.ShowGroupMembers(whc, group, true); err != nil {
					return m, err
				} else if _, err = whc.Responder().SendMessage(c, m, bots.BotApiSendMessageOverHTTPS); err != nil {
					return m, err
				}

				m.Text = whc.Translate(trans.MESSAGE_TEXT_USER_JOINED_GROUP, fmt.Sprintf(`<a href="tg://user?id=%v">%v</a>`, whc.MustBotChatID(), appUser.FullName()))
			}

			m.Format = bots.MessageFormatHTML
			return
		},
	)
}
