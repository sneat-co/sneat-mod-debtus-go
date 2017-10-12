package bot_shared

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
)

func GetGroup(whc bots.WebhookContext) (group models.Group, err error) {

	if group.ID, err = GetUserGroupID(whc); err != nil {
		return
	} else if group.ID != "" {
		return dal.Group.GetGroupByID(whc.Context(), group.ID)
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