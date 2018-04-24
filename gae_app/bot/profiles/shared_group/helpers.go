package shared_group

import (
	"net/url"

	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/shared_all"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bytes"
	"fmt"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/db"
	"github.com/strongo/log"
	"context"
	"strconv"
)

func GetGroup(whc bots.WebhookContext, callbackUrl *url.URL) (group models.Group, err error) {
	if callbackUrl != nil {
		group.ID = callbackUrl.Query().Get("group")
	}
	if group.ID == "" {
		if group.ID, err = GetUserGroupID(whc); err != nil {
			return
		}
	}

	if group.ID != "" {
		return dal.Group.GetGroupByID(whc.Context(), group.ID)
	}

	if !whc.IsInGroup() {
		if callbackUrl != nil {
			err = errors.New("An attempt to get group ID outside of group chat without callback parameter 'group'.")
		}
		return
	}

	tgChat := whc.Input().(telegram_bot.TelegramWebhookInput).TgUpdate().Chat()
	var tgChatEntity *models.DtTelegramChatEntity
	if tgChatEntity, err = getTgChatEntity(whc); err != nil {
		return
	}
	return createGroupFromTelegram(whc, tgChatEntity, tgChat) // TODO: No need to pass tgChatEntity - need to be updated in transaction
}

func GetUserGroupID(whc bots.WebhookContext) (groupID string, err error) {
	var tgChatEntity *models.DtTelegramChatEntity
	if tgChatEntity, err = getTgChatEntity(whc); err != nil || tgChatEntity == nil {
		return
	}
	if groupID = tgChatEntity.UserGroupID; groupID != "" {
		return
	}
	return
}

func createGroupFromTelegram(whc bots.WebhookContext, chatEntity *models.DtTelegramChatEntity, tgChat *tgbotapi.Chat) (group models.Group, err error) {
	c := whc.Context()
	log.Debugf(c, "createGroupFromTelegram()")
	var user *models.AppUserEntity
	if user, err = shared_all.GetUser(whc); err != nil {
		return
	}
	var chatInviteLink string

	if tgChat.IsSuperGroup() { // See: https://core.telegram.org/bots/api#exportchatinvitelink
		// TODO: Do this in delayed task - Lets try to get chat  invite link
		msg := bots.MessageFromBot{BotMessage: telegram_bot.ExportChatInviteLink{}}
		if tgResponse, err := whc.Responder().SendMessage(c, msg, bots.BotApiSendMessageOverHTTPS); err != nil {
			log.Debugf(c, "Not able to export chat invite link: %v", err)
		} else {
			chatInviteLink = string(tgResponse.TelegramMessage.(tgbotapi.APIResponse).Result)
			log.Debugf(c, "exportInviteLink response: %v", chatInviteLink)
		}
	}

	userID := whc.AppUserStrID()
	groupEntity := models.GroupEntity{
		CreatorUserID: userID,
		Name:          tgChat.Title,
	}
	groupEntity.SetTelegramGroups([]models.GroupTgChatJson{
		{
			ChatID:         tgChat.ID,
			Title:          tgChat.Title,
			ChatInviteLink: chatInviteLink,
		},
	})

	hasTgGroupEntity := false
	beforeGroupInsert := func(c context.Context, groupEntity *models.GroupEntity) (group models.Group, err error) {
		log.Debugf(c, "beforeGroupInsert()")
		var tgGroup models.TgGroup
		if tgGroup, err = dal.TgGroup.GetTgGroupByID(c, tgChat.ID); err != nil {
			if db.IsNotFound(err) {
				err = nil
			} else {
				return
			}
		}
		if tgGroup.TgGroupEntity != nil && tgGroup.UserGroupID != "" {
			hasTgGroupEntity = true
			return dal.Group.GetGroupByID(c, tgGroup.UserGroupID)
		}
		_, _, idx, member, members := groupEntity.AddOrGetMember(userID, "", user.FullName())
		member.TgUserID = strconv.FormatInt(int64(whc.Input().GetSender().GetID().(int)), 10)
		members[idx] = member
		groupEntity.SetGroupMembers(members)
		return
	}

	afterGroupInsert := func(c context.Context, group models.Group, user models.AppUser) (err error) {
		log.Debugf(c, "afterGroupInsert()")
		if !hasTgGroupEntity {
			if err = dal.TgGroup.SaveTgGroup(c, models.TgGroup{
				IntegerID: db.NewIntID(tgChat.ID),
				TgGroupEntity: &models.TgGroupEntity{
					UserGroupID: group.ID,
				},
			}); err != nil {
				return
			}
		}

		_ = user.AddGroup(group, whc.GetBotCode())
		chatEntity.UserGroupID = group.ID // TODO: !!! has to be updated in transaction!!!
		if err = whc.SaveBotChat(c, whc.GetBotCode(), whc.MustBotChatID(), chatEntity); err != nil {
			return
		}
		return
	}

	if group, _, err = facade.Group.CreateGroup(c, &groupEntity, whc.GetBotCode(), beforeGroupInsert, afterGroupInsert); err != nil {
		return
	}
	return
}

func getTgChatEntity(whc bots.WebhookContext) (tgChatEntity *models.DtTelegramChatEntity, err error) {
	chatEntity := whc.ChatEntity()
	if chatEntity == nil {
		whc.LogRequest()
		log.Debugf(whc.Context(), "can't get group as chatEntity == nil")
		return
	}
	var ok bool
	if tgChatEntity, ok = chatEntity.(*models.DtTelegramChatEntity); !ok {
		log.Debugf(whc.Context(), "whc.ChatEntity() is not TelegramChatEntityBase")
		return
	}
	return tgChatEntity, nil
}

func NewGroupTelegramInlineButton(whc bots.WebhookContext, groupsMessageID int) tgbotapi.InlineKeyboardButton {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "https://t.me/%v?startgroup=utm_s=%v__utm_m=%v__l=%v", whc.GetBotCode(), whc.GetBotCode(), "tgbot", whc.Locale().Code5)
	if groupsMessageID != 0 {
		buf.WriteString("__grpsMsgID=")
		buf.WriteString(strconv.Itoa(groupsMessageID))
	}
	return tgbotapi.InlineKeyboardButton{
		Text: whc.CommandText(trans.COMMAND_TEXT_ADD_GROUP, ""),
		URL:  buf.String(),
	}
}
