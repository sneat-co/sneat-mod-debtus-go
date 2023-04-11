package facade

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"fmt"
	tgstore "github.com/bots-go-framework/bots-fw-telegram/store"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/dal-go/dalgo/dal"
	"github.com/strongo/app"
	"github.com/strongo/log"
	"strconv"
)

func GetLocale(c context.Context, botID string, tgChatIntID, userID int64) (locale strongo.Locale, err error) {
	chatID := botsfw.NewChatID(botID, strconv.FormatInt(tgChatIntID, 10))
	//var tgChatEntity tgstore.ChatEntity
	tgChat := tgstore.NewChat(chatID, new(models.DebtusTelegramChatData))
	var db dal.Database
	if db, err = GetDatabase(c); err != nil {
		return
	}
	if err = db.Get(c, tgChat.Record); err != nil {
		log.Debugf(c, "Failed to get TgChat entity by string ID=%v: %v", tgChat.ID, err) // TODO: Replace with error once load by int ID removed
		if dal.IsNotFound(err) {
			panic("TODO: Remove this load by int ID")
			//if err = nds.Get(c, datastore.NewKey(c, tgstore.TgChatCollection, "", tgChatIntID, nil), &tgChatEntity); err != nil { // TODO: Remove this load by int ID
			//	log.Errorf(c, "Failed to get TgChat entity by int ID=%v: %v", tgChatIntID, err)
			//	return
			//}
		} else {
			return
		}
	}
	tgChatPreferredLanguage := tgChat.Data.BaseChatData().PreferredLanguage
	if tgChatPreferredLanguage == "" {
		if userID == 0 && tgChatEntity.AppUserIntID != 0 {
			userID = tgChatEntity.AppUserIntID
		}
		if userID != 0 {
			user, err := User.GetUserByID(c, tx, userID)
			if err != nil {
				log.Errorf(c, fmt.Errorf("failed to get user by ID=%v: %w", userID, err).Error())
				return locale, err
			}
			tgChatPreferredLanguage = user.Data.PreferredLanguage
		}
		if tgChatPreferredLanguage == "" {
			tgChatPreferredLanguage = strongo.LocaleCodeEnUS
			log.Warningf(c, "tgChat.PreferredLanguage == '' && user.PreferredLanguage == '', set to %v", strongo.LocaleCodeEnUS)
		}
	}
	locale = strongo.LocalesByCode5[tgChatPreferredLanguage]
	return
}
