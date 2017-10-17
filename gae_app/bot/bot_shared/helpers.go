package bot_shared

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"net/url"
	"github.com/pkg/errors"
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
		err = errors.New("An attempt to get group ID outside of group chat without callback parameter 'group'.")
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