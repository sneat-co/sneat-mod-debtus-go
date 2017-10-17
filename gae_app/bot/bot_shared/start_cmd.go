package bot_shared

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/platforms/telegram"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/app"
	"github.com/strongo/app/db"
	"github.com/strongo/app/log"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"golang.org/x/net/context"
	"net/url"
	"strconv"
	"fmt"
	"bytes"
)


func StartBotLink(botID, command string, params... string) string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "https://t.me/%v?start=%v", botID, command)
	for _, p := range params {
		buf.WriteString("__")
		buf.WriteString(p)
	}
	return buf.String()
}

func startCommand(botParams BotParams) bots.Command {
	return bots.Command{
		Code:       "start",
		Commands:   []string{"/start"},
		InputTypes: []bots.WebhookInputType{bots.WebhookInputInlineQuery},
		Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
			whc.LogRequest()
			c := whc.Context()
			text := whc.Input().(bots.WebhookTextMessage).Text()
			log.Debugf(c, "startCommand.Action() => text: "+text)

			_, startParams := telegram.ParseStartCommand(whc)

			if whc.IsInGroup() {
				return startInGroupAction(whc)
			} else {
				return startInBotAction(whc, startParams, botParams)
			}
		},
	}
}

func GetUser(whc bots.WebhookContext) (userEntity *models.AppUserEntity, err error) { // TODO: Make library and use across app
	var botAppUser bots.BotAppUser
	if botAppUser, err = whc.GetAppUser(); err != nil {
		return
	}
	userEntity = botAppUser.(*models.AppUserEntity)
	return
}

var LangKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	[]tgbotapi.InlineKeyboardButton{
		{
			Text:         strongo.LocaleEnUS.TitleWithIcon(),
			CallbackData: ON_START_CALLBACK_COMMAND + "?lang=" + strongo.LOCALE_EN_US,
		},
		{
			Text:         strongo.LocaleRuRu.TitleWithIcon(),
			CallbackData: ON_START_CALLBACK_COMMAND + "?lang=" + strongo.LOCALE_RU_RU,
		},
	},
)

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

const ON_START_CALLBACK_COMMAND = "on-start-callback"

func onStartCallbackCommand(params BotParams) bots.Command {
	return bots.NewCallbackCommand(ON_START_CALLBACK_COMMAND,
		func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
			lang := callbackUrl.Query().Get("lang")
			c := whc.Context()
			log.Debugf(c, "Locale: "+lang)

			whc.ChatEntity().SetPreferredLanguage(lang)

			if user, err := whc.GetAppUser(); err != nil {
				return m, err
			} else if err = user.SetPreferredLocale(lang); err != nil {
				return m, err
			} else if err = whc.SaveAppUser(whc.AppUserIntID(), user); err != nil {
				return m, err
			}

			if err = whc.SetLocale(lang); err != nil {
				return
			}

			if whc.IsInGroup() {
				return onStartCallbackInGroup(whc, params)
			} else {
				return onStartCallbackInBot(whc, params)
			}
		},
	)
}

func createGroupFromTelegram(whc bots.WebhookContext, chatEntity *models.DtTelegramChatEntity, tgChat *tgbotapi.Chat) (group models.Group, err error) {
	c := whc.Context()
	log.Debugf(c, "createGroupFromTelegram()")
	var user *models.AppUserEntity
	if user, err = GetUser(whc); err != nil {
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
				ID: tgChat.ID,
				TgGroupEntity: &models.TgGroupEntity{
					UserGroupID: group.ID,
				},
			}); err != nil {
				return
			}
		}

		_ = user.AddGroup(group, whc.GetBotCode())
		chatEntity.UserGroupID = group.ID  // TODO: !!! has to be updated in transaction!!!
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

const HOW_TO_COMMAND = "how-to"

var howToCommand = bots.Command{
	Code: HOW_TO_COMMAND,
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		m.Text = "<b>How To</b> - not implemented yet"
		return
	},
}
