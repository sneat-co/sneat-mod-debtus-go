package bot_shared

import (
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
)

func GetGroup(whc bots.WebhookContext) (group models.Group, err error) {
	var tgChatEntity *models.DtTelegramChatEntity
	if tgChatEntity, err = getTgChatEntity(whc); err != nil {
		return
	}
	if tgChatEntity.UserGroupID != "" {
		return dal.Group.GetGroupByID(whc.Context(), tgChatEntity.UserGroupID)
	}
	tgChat := whc.Input().(telegram_bot.TelegramWebhookInput).TgUpdate().Chat()
	return createGroupFromTelegram(whc, tgChatEntity, tgChat)
}

