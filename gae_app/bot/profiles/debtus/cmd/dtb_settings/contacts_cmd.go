package dtb_settings

import (
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-framework/core"
	"net/url"
	//"github.com/strongo/bots-api-telegram"
	//"github.com/DebtsTracker/translations/emoji"
	//"github.com/strongo/bots-framework/platforms/telegram"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bytes"
	"fmt"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/strongo/bots-api-telegram"
	"html"
	"strconv"
	"time"
)

const CONTACTS_LIST_COMMAND = "contacts-list"

var ContactsListCommand = bots.Command{
	Code:     CONTACTS_LIST_COMMAND,
	Commands: trans.Commands(CONTACTS_LIST_COMMAND),
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		var user *models.AppUserEntity
		if appUser, err := whc.GetAppUser(); err != nil {
			return m, err
		} else {
			user = appUser.(*models.AppUserEntity)
		}
		var buffer bytes.Buffer
		buffer.WriteString(fmt.Sprintf("<b>%v</b>\n", whc.Translate(trans.COMMAND_TEXT_CONTACTS)))
		linker := common.NewLinkerFromWhc(whc)
		contacts := user.Contacts()
		numFormat := "%0" + strconv.Itoa(len(strconv.Itoa(len(contacts)))) + "d. "
		if len(contacts) == 0 {
			buffer.WriteString(whc.Translate(trans.MESSAGE_TEXT_YOU_HAVE_NO_CONTACTS))
		} else {
			for i, contact := range contacts {
				buffer.WriteString(fmt.Sprintf(numFormat, i+1))
				buffer.WriteString(fmt.Sprintf(`<a href="%v">%v</a>`, linker.UrlToContact(contact.ID), html.EscapeString(contact.Name)))
				if contact.Status != "" && contact.Status != models.STATUS_ACTIVE {
					buffer.WriteString(" (")
					buffer.WriteString(contact.Status)
					buffer.WriteString(")")
				}
				buffer.WriteString("\n")
			}
		}
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			[]tgbotapi.InlineKeyboardButton{
				{
					Text:         whc.CommandText(trans.COMMAND_TEXT_REFRESH, emoji.REFRESH_ICON),
					CallbackData: CONTACTS_LIST_COMMAND + "?do=refresh",
				},
			},
		)
		buffer.WriteString(fmt.Sprintf("\n\nRefreshed on: %v", time.Now()))
		m = whc.NewMessage(buffer.String())
		m.Keyboard = keyboard
		m.IsEdit = whc.InputType() == bots.WebhookInputCallbackQuery
		//if callbackUrl.Query().Get("do") == "refresh" {
		//	if m, err = bot.SendRefreshOrNothingChanged(whc, m); err != nil {
		//		return
		//	}
		//}
		return
	},
}

//const CONTACT_DETAILS_COMMAND = "contact-details"
//
//var ContactDetailsCommand = bots.Command{
//	Code:     CONTACTS_LIST_COMMAND,
//	Commands: trans.Commands(CONTACTS_LIST_COMMAND),
//	CallbackAction: func(whc bots.WebhookContext, _ *url.URL) (m bots.MessageFromBot, err error) {
//		keyboard := tgbotapi.NewInlineKeyboardMarkup(
//			[]tgbotapi.InlineKeyboardButton{
//				{
//					Text:         whc.CommandText(trans.COMMAND_TEXT_LANGUAGE, emoji.EARTH_ICON),
//					CallbackData: SETTINGS_LOCALE_LIST_CALLBACK_PATH,
//				},
//			},
//		)
//		messageText := whc.NewMessageByCode(trans.MESSAGE_TEXT_CONTACT_DETAILS)
//		m.TelegramEditMessageText = telegram_bot.EditMessageOnCallbackQuery(whc.Input().(bots.WebhookCallbackQuery), "HTML", messageText)
//		m.TelegramEditMessageText.ReplyMarkup = keyboard
//		return
//	},
//}
//
//const DELETE_CONTACT_COMMAND = "delete-contact"
//
//var DeleteContactCommand = bots.Command{
//	Code:     DELETE_CONTACT_COMMAND,
//	Commands: trans.Commands(CONTACTS_LIST_COMMAND),
//	CallbackAction: func(whc bots.WebhookContext, _ *url.URL) (m bots.MessageFromBot, err error) {
//
//		return
//	},
//}
//
//const EDIT_CONTACT_NAME_COMMAND = "edit-contact-name"
//
//var EditContactNameCommand = bots.Command{
//	Code:     EDIT_CONTACT_NAME_COMMAND,
//	Commands: trans.Commands(CONTACTS_LIST_COMMAND),
//	CallbackAction: func(whc bots.WebhookContext, _ *url.URL) (m bots.MessageFromBot, err error) {
//
//		return
//	},
//}
