package facade

import (
	"strconv"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"github.com/pkg/errors"
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/log"
	"github.com/strongo/nds"
	"context"
	"google.golang.org/appengine/datastore"
)

func GetLocale(c context.Context, botID string, tgChatIntID, userID int64) (locale strongo.Locale, err error) {
	botChatKey := datastore.NewKey(c, telegram.ChatKind, bots.NewChatID(botID, strconv.FormatInt(tgChatIntID, 10)), 0, nil)
	var tgChatEntity telegram.TgChatEntityBase
	if err = nds.Get(c, botChatKey, &tgChatEntity); err != nil {
		log.Debugf(c, "Failed to get TgChat entity by string ID=%v: %v", botChatKey.StringID(), err) // TODO: Replace with error once load by int ID removed
		if err == datastore.ErrNoSuchEntity {
			if err = nds.Get(c, datastore.NewKey(c, telegram.ChatKind, "", tgChatIntID, nil), &tgChatEntity); err != nil { // TODO: Remove this load by int ID
				log.Errorf(c, "Failed to get TgChat entity by int ID=%v: %v", tgChatIntID, err)
				return
			}
		} else {
			return
		}
	}
	tgChatPreferredLanguage := tgChatEntity.PreferredLanguage
	if tgChatPreferredLanguage == "" {
		if userID == 0 && tgChatEntity.AppUserIntID != 0 {
			userID = tgChatEntity.AppUserIntID
		}
		if userID != 0 {
			user, err := dal.User.GetUserByID(c, userID)
			if err != nil {
				log.Errorf(c, errors.Wrapf(err, "Failed to get user by ID=%v", userID).Error())
				return locale, err
			}
			tgChatPreferredLanguage = user.PreferredLanguage
		}
		if tgChatPreferredLanguage == "" {
			tgChatPreferredLanguage = strongo.LOCALE_EN_US
			log.Warningf(c, "tgChat.PreferredLanguage == '' && user.PreferredLanguage == '', set to %v", strongo.LOCALE_EN_US)
		}
	}
	locale = strongo.LocalesByCode5[tgChatPreferredLanguage]
	return
}
