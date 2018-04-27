package shared_all

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/platforms/tgbots"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bytes"
	"fmt"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
	"context"
	"net/url"
	"strings"
)

func StartBotLink(botID, command string, params ...string) string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "https://t.me/%v?start=%v", botID, command)
	for _, p := range params {
		buf.WriteString("__")
		buf.WriteString(p)
	}
	return buf.String()
}

func createStartCommand(botParams BotParams) bots.Command {
	return bots.Command{
		Code:       "start",
		Commands:   []string{"/start"},
		InputTypes: []bots.WebhookInputType{bots.WebhookInputInlineQuery},
		Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
			whc.LogRequest()
			c := whc.Context()
			text := whc.Input().(bots.WebhookTextMessage).Text()
			log.Debugf(c, "createStartCommand.Action() => text: "+text)

			startParam, startParams := tgbots.ParseStartCommand(whc)

			if whc.IsInGroup() {
				return botParams.StartInGroupAction(whc)
			} else {
				chatEntity := whc.ChatEntity()
				chatEntity.SetAwaitingReplyTo("")

				switch {
				case startParam == "help_inline":
					return startInlineHelp(whc)
				case strings.HasPrefix(startParam, "login-"):
					loginID, err := common.DecodeID(startParam[len("login-"):])
					if err != nil {
						return m, err
					}
					return startLoginGac(whc, loginID)
					//case strings.HasPrefix(textToMatchNoStart, JOIN_BILL_COMMAND):
					//	return JoinBillCommand.Action(whc)
				case strings.HasPrefix(startParam, "refbytguser-") && startParam != "refbytguser-YOUR_CHANNEL":
					facade.Referer.AddTelegramReferrer(c, whc.AppUserIntID(), strings.TrimPrefix(startParam, "refbytguser-"), whc.GetBotCode())
				}
				return startInBotAction(whc, startParams, botParams)
			}
		},
	}
}

func startLoginGac(whc bots.WebhookContext, loginID int64) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	var loginPin models.LoginPin
	if loginPin, err = facade.AuthFacade.AssignPinCode(c, loginID, whc.AppUserIntID()); err != nil {
		return
	}
	return whc.NewMessageByCode(trans.MESSAGE_TEXT_LOGIN_CODE, models.LoginCodeToString(loginPin.Code)), nil
}

func startInlineHelp(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
	m = whc.NewMessage("<b>Help: How to use this bot in chats</b>\n\nExplain here how to use bot's inline mode.")
	m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{{Text: "Button 1", URL: "https://debtstracker.io/#btn=1"}},
		[]tgbotapi.InlineKeyboardButton{{Text: "Button 2", URL: "https://debtstracker.io/#btn=2"}},
		//[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonSwitch("Back to chat 1", "1")},
		//[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonSwitch("Back to chat 2", "2")},
		[]tgbotapi.InlineKeyboardButton{{Text: "Button 3", CallbackData: "help-3"}},
		[]tgbotapi.InlineKeyboardButton{{Text: "Button 4", CallbackData: "help-4"}},
		[]tgbotapi.InlineKeyboardButton{{Text: "Button 5", CallbackData: "help-5"}},
	)
	return m, err
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
			CallbackData: onStartCallbackCommandCode + "?lang=" + strongo.LOCALE_EN_US,
		},
		{
			Text:         strongo.LocaleRuRu.TitleWithIcon(),
			CallbackData: onStartCallbackCommandCode + "?lang=" + strongo.LOCALE_RU_RU,
		},
	},
)

const onStartCallbackCommandCode = "on-start-callback"

func onStartCallbackCommand(params BotParams) bots.Command {
	return bots.NewCallbackCommand(onStartCallbackCommandCode,
		func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
			lang := callbackUrl.Query().Get("lang")
			c := whc.Context()
			log.Debugf(c, "Locale: "+lang)

			whc.ChatEntity().SetPreferredLanguage(lang)

			if err = dal.DB.RunInTransaction(c, func(c context.Context) error {
				if user, err := dal.User.GetUserByID(c, whc.AppUserIntID()); err != nil {
					return err
				} else if err = user.SetPreferredLocale(lang); err != nil {
					return err
				} else if err = dal.User.SaveUser(c, user); err != nil {
					return err
				}
				return nil
			}, nil); err != nil {
				return
			}

			if err = whc.SetLocale(lang); err != nil {
				return
			}

			//if whc.IsInGroup() {
			//	var group models.Group
			//	if group, err = GetGroup(whc, callbackUrl); err != nil {
			//		return
			//	}
			//	return onStartCallbackInGroup(whc, group, params)
			//} else {
			//	return onStartCallbackInBot(whc, params)
			//}
			return
		},
	)
}
