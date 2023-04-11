package shared_all

import (
	"fmt"
	"net/url"
	"strings"

	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_general"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bytes"
	"context"
	"errors"
	"github.com/strongo/app"
	"github.com/strongo/log"
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

func createOnboardingAskLocaleCommand(botParams BotParams) botsfw.Command {
	return botsfw.Command{
		Code:       onboardingAskLocaleCommandCode,
		ExactMatch: trans.ChooseLocaleIcon,
		Action: func(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error) {
			return onboardingAskLocaleAction(whc, "", botParams)
		},
	}
}

func onboardingAskLocaleAction(whc botsfw.WebhookContext, messagePrefix string, botParams BotParams) (m botsfw.MessageFromBot, err error) {
	chatEntity := whc.ChatEntity()

	if chatEntity.IsAwaitingReplyTo(onboardingAskLocaleCommandCode) {
		messageText := whc.Input().(botsfw.WebhookTextMessage).Text()
		for _, locale := range trans.SupportedLocales {
			if locale.TitleWithIcon() == messageText {
				return setPreferredLanguageAction(whc, locale.Code5, "onboarding", botParams)
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

var askPreferredLocaleFromSettingsCallback = botsfw.Command{
	Code: SettingsLocaleListCallbackPath,
	CallbackAction: func(whc botsfw.WebhookContext, _ *url.URL) (m botsfw.MessageFromBot, err error) {
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
		if m, err = whc.NewEditMessage(whc.Translate(trans.MESSAGE_TEXT_CHOOSE_UI_LANGUAGE), botsfw.MessageFormatHTML); err != nil {
			return
		}
		m.Keyboard = keyboard
		return m, err
	},
}

func setLocaleCallbackCommand(botParams BotParams) botsfw.Command {
	return botsfw.Command{
		Code: SettingsLocaleSetCallbackPath,
		CallbackAction: func(whc botsfw.WebhookContext, callbackUrl *url.URL) (m botsfw.MessageFromBot, err error) {
			return setPreferredLanguageAction(whc, callbackUrl.Query().Get("code5"), callbackUrl.Query().Get("mode"), botParams)
		},
	}
}

func setPreferredLanguageAction(whc botsfw.WebhookContext, code5, mode string, botParams BotParams) (m botsfw.MessageFromBot, err error) {
	c := whc.Context()
	log.Debugf(c, "setPreferredLanguageAction(code5=%v, mode=%v)", code5, mode)
	appUser, err := whc.GetAppUser()
	if err != nil {
		log.Errorf(c, ": %v", err)
		return m, errors.WithMessage(err, "failed to load userEntity")
	}
	userEntity, ok := appUser.(*models.AppUserEntity)
	if !ok {
		return m, fmt.Errorf("expected *models.AppUser, got: %T", appUser)
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

				if err = dtdal.DB.RunInTransaction(c, func(c context.Context) (err error) {
					var user models.AppUser
					if user, err = facade.User.GetUserByID(c, whc.AppUserIntID()); err != nil {
						return
					}
					if err = user.SetPreferredLocale(locale.Code5); err != nil {
						err = errors.WithMessage(err, "failed to set preferred locale for user")
					}
					chatEntity.SetPreferredLanguage(locale.Code5)
					chatEntity.SetAwaitingReplyTo("")
					if err = whc.SaveBotChat(c, whc.GetBotCode(), whc.MustBotChatID(), chatEntity); err != nil {
						return
					}
					return facade.User.SaveUser(c, user)
				}, db.CrossGroupTransaction); err != nil {
					return
				}
				localeChanged = true
				selectedLocale = locale
				if whc.GetBotSettings().Env == strongo.EnvProduction {
					ga := whc.GA()
					gaEvent := ga.GaEventWithLabel("settings", "locale-changed", strings.ToLower(locale.Code5))
					if gaErr := ga.Queue(gaEvent); gaErr != nil {
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
	} else {
		selectedLocale = strongo.GetLocaleByCode5(chatEntity.GetPreferredLanguage())
	}
	//if localeChanged {

	switch mode {
	case "onboarding":
		log.Debugf(c, "whc.Locale().Code5: %v", whc.Locale().Code5)
		m = whc.NewMessageByCode(trans.MESSAGE_TEXT_YOUR_SELECTED_PREFERRED_LANGUAGE, selectedLocale.NativeTitle)
		botParams.SetMainMenu(whc, &m)
		if _, err = whc.Responder().SendMessage(c, m, botsfw.BotAPISendMessageOverHTTPS); err != nil {
			log.Errorf(c, "Failed to notify userEntity about selected language: %v", err)
			// Not critical, lets continue
		}
		return aboutDrawAction(whc, nil)
	case "settings":
		if localeChanged {
			m, err = dtb_general.MainMenuAction(whc, whc.Translate(trans.MESSAGE_TEXT_LOCALE_CHANGED, selectedLocale.TitleWithIcon()), false)
			if _, err = whc.Responder().SendMessage(c, m, botsfw.BotAPISendMessageOverHTTPS); err != nil {
				return m, err
			}
			return SettingsMainAction(whc)
			//if _, err = whc.Responder().SendMessage(c, m, botsfw.BotAPISendMessageOverHTTPS); err != nil {
			//	return m, err
			//}
			//return dtb_general.MainMenuAction(whc, )
		} else {
			return SettingsMainAction(whc)
		}
	default:
		panic(fmt.Sprintf("Unknown mode: %v", mode))
	}
}

const (
	moreAboutDrawCommandCode = "more-about-draw"
	joinDrawCommandCode      = "join-draw"
)

var aboutDrawCommand = botsfw.Command{
	Commands: []string{"/draw"},
	Code:     moreAboutDrawCommandCode,
	Action: func(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error) {
		return aboutDrawAction(whc, nil)
	},
	CallbackAction: aboutDrawAction,
}

var joinDrawCommand = botsfw.Command{
	Code:           joinDrawCommandCode,
	CallbackAction: aboutDrawAction,
}

func aboutDrawAction(whc botsfw.WebhookContext, callbackUrl *url.URL) (m botsfw.MessageFromBot, err error) {
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
	m.Format = botsfw.MessageFormatHTML
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
			if _, err = whc.Responder().SendMessage(c, m, botsfw.BotAPISendMessageOverHTTPS); err != nil {
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

//func LanguageOptions(whc botsfw.WebhookContext, mainMenu bool) tgbotapi.ReplyKeyboardMarkup {
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
