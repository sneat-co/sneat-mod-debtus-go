package dtb_settings

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_general"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/app"
	"github.com/strongo/app/log"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/measurement-protocol"
	"net/url"
	"strings"
)

const (
	SETTINGS_LOCALE_LIST_CALLBACK_PATH = "settings/locale/list"
	SETTINGS_LOCALE_SET_CALLBACK_PATH  = "settings/locale/set"
)

const ONBOARDING_ASK_LOCALE_COMMAND = "onboarding-ask-locale"

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
		{Text: strongo.LocaleFaIr.TitleWithIcon()},
	},
)

var OnboardingAskLocaleCommand = bots.Command{
	Code:       ONBOARDING_ASK_LOCALE_COMMAND,
	ExactMatch: trans.ChooseLocaleIcon,
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		return onboardingAskLocaleAction(whc, "")
	},
}

func onboardingAskLocaleAction(whc bots.WebhookContext, messagePrefix string) (m bots.MessageFromBot, err error) {
	chatEntity := whc.ChatEntity()

	if chatEntity.IsAwaitingReplyTo(ONBOARDING_ASK_LOCALE_COMMAND) {
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
		chatEntity.SetAwaitingReplyTo(ONBOARDING_ASK_LOCALE_COMMAND)
		m = whc.NewMessageByCode(trans.MESSAGE_TEXT_ONBOARDING_ASK_TO_CHOOSE_LANGUAGE, whc.GetSender().GetFirstName())
		//localesReplyKeyboard.OneTimeKeyboard = true
		m.Keyboard = localesReplyKeyboard
	}
	return
}

var AskPreferredLocaleFromSettingsCallback = bots.Command{
	Code: SETTINGS_LOCALE_LIST_CALLBACK_PATH,
	CallbackAction: func(whc bots.WebhookContext, _ *url.URL) (m bots.MessageFromBot, err error) {
		callbackData := fmt.Sprintf("%v?mode=settings&code5=", SETTINGS_LOCALE_SET_CALLBACK_PATH)
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
			{Text: SettingsCommand.DefaultTitle(whc), CallbackData: SETTINGS_CALLBACK_PATH},
		})
		if m, err = whc.NewEditMessage(whc.Translate(trans.MESSAGE_TEXT_CHOOSE_UI_LANGUAGE), bots.MessageFormatHTML); err != nil {
			return
		}
		m.Keyboard = keyboard
		return m, err
	},
}

var SetLocaleCallbackCommand = bots.Command{
	Code: SETTINGS_LOCALE_SET_CALLBACK_PATH,
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		return setPreferredLanguageCommand(whc, callbackUrl.Query().Get("code5"), callbackUrl.Query().Get("mode"))
	},
}

func setPreferredLanguageCommand(whc bots.WebhookContext, code5, mode string) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	log.Debugf(c, "setPreferredLanguageCommand(code5=%v, mode=%v)", code5, mode)
	userEntity, err := whc.GetAppUser()
	if err != nil {
		log.Errorf(c, ": %v", err)
		return m, errors.Wrap(err, "Failed to load user")
	}
	user, ok := userEntity.(*models.AppUserEntity)
	if !ok {
		return m, errors.New(fmt.Sprintf("Expected *models.AppUser, got: %T", userEntity))
	}

	var (
		localeChanged  bool
		selectedLocale strongo.Locale
	)

	chatEntity := whc.ChatEntity()
	log.Debugf(c, "user.PreferredLanguage: %v, chatEntity.GetPreferredLanguage(): %v, code5: %v", user.PreferredLanguage, chatEntity.GetPreferredLanguage(), code5)
	if user.PreferredLanguage != code5 || chatEntity.GetPreferredLanguage() != code5 {
		log.Debugf(c, "PreferredLanguage will be updated for user & chat entities.")
		for _, locale := range trans.SupportedLocalesByCode5 {
			if locale.Code5 == code5 {
				whc.SetLocale(locale.Code5)
				err := user.SetPreferredLocale(locale.Code5)
				if err != nil {
					return m, errors.Wrap(err, "Failed to set preferred locale for user")
				}
				chatEntity.SetPreferredLanguage(locale.Code5)
				//err = whc.SaveBotChat(whc.BotChatID(), chatEntity) // TODO: Should be run in transaction
				if err != nil {
					return m, errors.Wrap(err, "Failed to save chat entity to datastore")
				}

				err = whc.SaveAppUser(whc.AppUserIntID(), user)
				if err != nil {
					return m, errors.Wrap(err, "Failed to save user to datastore")
				}
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
		if _, err = whc.Responder().SendMessage(c, m, bots.BotApiSendMessageOverHTTPS); err != nil {
			log.Errorf(c, "Failed to notify user about selected language: %v", err)
			// Not critical, lets continue
		}
		isAccessGranted := true //bots.IsAccessGranted(whc)
		if isAccessGranted {
			log.Debugf(c, "IsAccessGranted(): %v", isAccessGranted)
			// TODO: Reply to callback as well
			return dtb_general.MainMenuAction(whc, "", false)
		} else {
			return OnboardingTellAboutInviteCodeAction(whc)
		}
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
			return SettingsAction(whc)
		}
	default:
		panic(fmt.Sprintf("Unknown mode: %v", mode))
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
