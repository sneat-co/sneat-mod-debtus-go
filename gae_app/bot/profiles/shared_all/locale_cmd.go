package shared_all

import (
	"fmt"
	"net/url"
	"strings"

	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_general"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/app"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
	"github.com/strongo/measurement-protocol"
	"golang.org/x/net/context"
	"bytes"
)

const (
	SettingsLocaleListCallbackPath = "settings/locale/list"
	SettingsLocaleSetCallbackPath  = "settings/locale/set"
)

const onboardingAskLocaleCommandCode = "onboarding-ask-locale"

var localesReplyKeyboard = tgbotapi.NewReplyKeyboard(
	[]tgbotapi.KeyboardButton{
		{Text: strongo.LocaleEnUS.TitleWithIcon()},
		{Text: strongo.LocaleRuRu.TitleWithIcon()},
	},
	[]tgbotapi.KeyboardButton{
		{Text: strongo.LocaleEsEs.TitleWithIcon()},
		{Text: strongo.LocaleItIt.TitleWithIcon()},
	},
	[]tgbotapi.KeyboardButton{
		{Text: strongo.LocaleDeDe.TitleWithIcon()},
		{Text: strongo.LocaleFaIr.TitleWithIcon()},
	},
)

var onboardingAskLocaleCommand = bots.Command{
	Code:       onboardingAskLocaleCommandCode,
	ExactMatch: trans.ChooseLocaleIcon,
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		return onboardingAskLocaleAction(whc, "")
	},
}

func onboardingAskLocaleAction(whc bots.WebhookContext, messagePrefix string) (m bots.MessageFromBot, err error) {
	chatEntity := whc.ChatEntity()

	if chatEntity.IsAwaitingReplyTo(onboardingAskLocaleCommandCode) {
		messageText := whc.Input().(bots.WebhookTextMessage).Text()
		for _, locale := range trans.SupportedLocales {
			if locale.TitleWithIcon() == messageText {
				return setPreferredLanguageCommand(whc, locale.Code5, "onboarding")
			}
		}
		m = whc.NewMessageByCode(trans.MESSAGE_TEXT_UNKNOWN_LANGUAGE)
		//localesReplyKeyboard.OneTimeKeyboard = true
		m.Keyboard = localesReplyKeyboard
	} else {
		m.Text = messagePrefix + m.Text
		chatEntity.SetAwaitingReplyTo(onboardingAskLocaleCommandCode)
		m = whc.NewMessageByCode(trans.MESSAGE_TEXT_ONBOARDING_ASK_TO_CHOOSE_LANGUAGE, whc.GetSender().GetFirstName())
		//localesReplyKeyboard.OneTimeKeyboard = true
		m.Keyboard = localesReplyKeyboard
	}
	return
}

var askPreferredLocaleFromSettingsCallback = bots.Command{
	Code: SettingsLocaleListCallbackPath,
	CallbackAction: func(whc bots.WebhookContext, _ *url.URL) (m bots.MessageFromBot, err error) {
		callbackData := fmt.Sprintf("%v?mode=settings&code5=", SettingsLocaleSetCallbackPath)
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			[]tgbotapi.InlineKeyboardButton{
				{Text: strongo.LocaleEnUS.TitleWithIcon(), CallbackData: callbackData + strongo.LocaleEnUS.Code5},
				{Text: strongo.LocaleRuRu.TitleWithIcon(), CallbackData: callbackData + strongo.LocaleRuRu.Code5},
			},
			[]tgbotapi.InlineKeyboardButton{
				{Text: strongo.LocaleEsEs.TitleWithIcon(), CallbackData: callbackData + strongo.LocaleEsEs.Code5},
				{Text: strongo.LocaleItIt.TitleWithIcon(), CallbackData: callbackData + strongo.LocaleItIt.Code5},
			},
			[]tgbotapi.InlineKeyboardButton{
				{Text: strongo.LocaleDeDe.TitleWithIcon(), CallbackData: callbackData + strongo.LocaleDeDe.Code5},
				{Text: strongo.LocaleFaIr.TitleWithIcon(), CallbackData: callbackData + strongo.LocaleFaIr.Code5},
			},
		) //dtb_general.LanguageOptions(whc, false)
		log.Debugf(whc.Context(), "AskPreferredLanguage(): locale: %v", whc.Locale().Code5)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{
			{Text: whc.Translate(trans.COMMAND_TEXT_SETTING), CallbackData: SettingsCommandCode},
		})
		if m, err = whc.NewEditMessage(whc.Translate(trans.MESSAGE_TEXT_CHOOSE_UI_LANGUAGE), bots.MessageFormatHTML); err != nil {
			return
		}
		m.Keyboard = keyboard
		return m, err
	},
}

func setLocaleCallbackCommand(params BotParams) bots.Command {
	return bots.Command{
		Code: SettingsLocaleSetCallbackPath,
		CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
			return setPreferredLanguageCommand(whc, callbackUrl.Query().Get("code5"), callbackUrl.Query().Get("mode"))
		},
	}
}

func setPreferredLanguageCommand(whc bots.WebhookContext, code5, mode string) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	log.Debugf(c, "setPreferredLanguageCommand(code5=%v, mode=%v)", code5, mode)
	appUser, err := whc.GetAppUser()
	if err != nil {
		log.Errorf(c, ": %v", err)
		return m, errors.Wrap(err, "Failed to load userEntity")
	}
	userEntity, ok := appUser.(*models.AppUserEntity)
	if !ok {
		return m, fmt.Errorf("Expected *models.AppUser, got: %T", appUser)
	}

	var (
		localeChanged  bool
		selectedLocale strongo.Locale
	)

	chatEntity := whc.ChatEntity()
	log.Debugf(c, "userEntity.PreferredLanguage: %v, chatEntity.GetPreferredLanguage(): %v, code5: %v", userEntity.PreferredLanguage, chatEntity.GetPreferredLanguage(), code5)
	if userEntity.PreferredLanguage != code5 || chatEntity.GetPreferredLanguage() != code5 {
		log.Debugf(c, "PreferredLanguage will be updated for userEntity & chat entities.")
		for _, locale := range trans.SupportedLocalesByCode5 {
			if locale.Code5 == code5 {
				whc.SetLocale(locale.Code5)

				if err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
					var user models.AppUser
					if user, err = dal.User.GetUserByID(c, whc.AppUserIntID()); err != nil {
						return
					}
					if err = user.SetPreferredLocale(locale.Code5); err != nil {
						err = errors.WithMessage(err, "Failed to set preferred locale for user")
					}
					return dal.User.SaveUser(c, user)
				}, nil); err != nil {
					return
				}
				chatEntity.SetPreferredLanguage(locale.Code5)
				//if err = whc.SaveBotChat(whc.BotChatID(), chatEntity); err != nil { // TODO: Should be run in transaction
				//	return m, errors.Wrap(err, "Failed to save chat entity to datastore")
				//}
				localeChanged = true
				selectedLocale = locale
				if whc.GetBotSettings().Env == strongo.EnvProduction {
					gaEvent := measurement.NewEvent("settings", "locale-changed", whc.GaCommon())
					gaEvent.Label = strings.ToLower(locale.Code5)
					gaErr := whc.GaMeasurement().Queue(gaEvent)
					if gaErr != nil {
						log.Warningf(c, "Failed to log event: %v", gaErr)
					} else {
						log.Infof(c, "GA event queued: %v", gaEvent)
					}
				}
				break
			}
		}
		if !localeChanged {
			log.Errorf(c, "Unknown locale: %v", code5)
		}
	}
	//if localeChanged {

	switch mode {
	case "onboarding":
		log.Debugf(c, "whc.Locale().Code5: %v", whc.Locale().Code5)
		m = whc.NewMessageByCode(trans.MESSAGE_TEXT_YOUR_SELECTED_PREFERRED_LANGUAGE, selectedLocale.NativeTitle)
		dtb_general.SetMainMenuKeyboard(whc, &m)
		if _, err = whc.Responder().SendMessage(c, m, bots.BotApiSendMessageOverHTTPS); err != nil {
			log.Errorf(c, "Failed to notify userEntity about selected language: %v", err)
			// Not critical, lets continue
		}
		return aboutDrawAction(whc, nil)
	case "settings":
		if localeChanged {
			m, err = dtb_general.MainMenuAction(whc, whc.Translate(trans.MESSAGE_TEXT_LOCALE_CHANGED, selectedLocale.TitleWithIcon()), false)
			if _, err = whc.Responder().SendMessage(c, m, bots.BotApiSendMessageOverHTTPS); err != nil {
				return m, err
			}
			return BackToSettingsAction(whc, "")
			//if _, err = whc.Responder().SendMessage(c, m, bots.BotApiSendMessageOverHTTPS); err != nil {
			//	return m, err
			//}
			//return dtb_general.MainMenuAction(whc, )
		} else {
			return BackToSettingsAction(whc, "")
		}
	default:
		panic(fmt.Sprintf("Unknown mode: %v", mode))
	}
}

const (
	moreAboutDrawCommandCode = "more-about-draw"
	joinDrawCommandCode      = "join-draw"
)

var aboutDrawCommand = bots.Command{
	Commands: []string{"/draw"},
	Code:     moreAboutDrawCommandCode,
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		return aboutDrawAction(whc, nil)
	},
	CallbackAction: aboutDrawAction,
}

var joinDrawCommand = bots.Command{
	Code:           joinDrawCommandCode,
	CallbackAction: aboutDrawAction,
}

func aboutDrawAction(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	buf := new(bytes.Buffer)
	sender := whc.GetSender()
	name := sender.GetFirstName()
	if name == "" {
		name = sender.GetUserName()
		if name == "" {
			name = sender.GetLastName()
		}
	}
	buf.WriteString(whc.Translate(trans.MESSAGE_TEXT_ABOUT_DRAW_SHORT, name))
	buf.WriteString("\n\n")
	m.Format = bots.MessageFormatHTML
	if callbackUrl == nil {
		buf.WriteString(whc.Translate(trans.MESSAGE_TEXT_ABOUT_DRAW_CALL_TO_ACTION))
		m.Text = buf.String()
		m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
			[]tgbotapi.InlineKeyboardButton{
				{
					Text:         whc.Translate(trans.COMMAN_TEXT_MORE_ABOUT_DRAW),
					CallbackData: moreAboutDrawCommandCode,
				},
			},
		)
		return
	} else {
		m.IsEdit = true
		buf.WriteString(whc.Translate(trans.MESSAGE_TEXT_ABOUT_DRAW_MORE))
		m.Text = buf.String()
		switch callbackUrl.Path {
		case moreAboutDrawCommandCode:
			m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
				[]tgbotapi.InlineKeyboardButton{
					{
						Text:         whc.Translate(trans.COMMAN_TEXT_I_AM_IN_DRAW),
						CallbackData: joinDrawCommandCode,
					},
				},
			)
			return
		case joinDrawCommandCode:
			if _, err = whc.Responder().SendMessage(c, m, bots.BotApiSendMessageOverHTTPS); err != nil {
				log.Warningf(c, "Failed to edit message: %v", err)
				err = nil // Not critical
			}
			m.IsEdit = false
			m.Text = whc.Translate(trans.MESSAGE_TEXT_JOINED_DRAW)
			return
		default:
			err = fmt.Errorf("unknown callback command: %v", callbackUrl.String())
			return
		}
	}
}

//func LanguageOptions(whc bots.WebhookContext, mainMenu bool) tgbotapi.ReplyKeyboardMarkup {
//
//	buttons := [][]string{}
//	buttons = append(buttons, []string{whc.Locale().TitleWithIcon()})
//	row := []string{"", ""}
//	col := 0
//	whcLocalCode := whc.Locale().Code5
//	for _, locale := range trans.SupportedLocales {
//		log.Debugf(c, "locale: %v, row: %v", locale, row)
//		if locale.Code5 == whcLocalCode {
//			log.Debugf(c, "continue")
//			continue
//		}
//		row[col] = locale.TitleWithIcon()
//		log.Debugf(c, "row: %v", row)
//		if col == 1 {
//			buttons = append(buttons, []string{row[0], row[1]})
//			log.Debugf(c, "col: %v, keyboard: %v", col, buttons)
//			col = 0
//		} else {
//			col += 1
//		}
//	}
//	if mainMenu {
//		buttons = append(buttons, []string{MainMenuCommand.DefaultTitle(whc)})
//	}
//	log.Debugf(c, "keyboard: %v", buttons)
//	return tgbotapi.NewReplyKeyboardUsingStrings(buttons)
//}
//
