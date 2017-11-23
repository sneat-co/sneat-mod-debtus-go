package dtb_general

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/platforms/viber"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/log"
	"github.com/strongo/bots-api-fbm"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-api-viber/viberinterface"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/fbm"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/bots-framework/platforms/viber"
)

const INVITES_SHOT_COMMAND = emoji.PRESENT_ICON

// This commands are required for main menu because of circular references
var _lendCommand = bots.Command{Code: "lend", Title: trans.COMMAND_TEXT_GAVE, Icon: emoji.GIVE_ICON}
var _borrowCommand = bots.Command{Code: "borrow", Title: trans.COMMAND_TEXT_GOT, Icon: emoji.TAKE_ICON}
var _returnCommand = bots.Command{Code: "return", Title: trans.COMMAND_TEXT_RETURN, Icon: emoji.RETURN_BACK_ICON}

func MainMenuKeyboardOnReceiptAck(whc bots.WebhookContext) *tgbotapi.ReplyKeyboardMarkup {
	return mainMenuTelegramKeyboard(whc, getMainMenuParams(whc, true))
}

type mainMenuParams struct {
	showBalanceAndHistory bool
	showReturn            bool
}

func getMainMenuParams(whc bots.WebhookContext, onReceiptAck bool) (params mainMenuParams) {
	var (
		user      *models.AppUserEntity
		isAppUser bool
	)

	c := whc.Context()
	if userEntity, err := whc.GetAppUser(); err != nil {
		log.Errorf(c, "Failed to get user: %v", err)
	} else if user, isAppUser = userEntity.(*models.AppUserEntity); !isAppUser {
		log.Errorf(c, "Failed to caset user to *models.AppUser: %T", userEntity)
	} else if onReceiptAck || !user.Balance().IsZero() {
		params.showReturn = true
	}
	params.showBalanceAndHistory = onReceiptAck || (user != nil && user.CountOfTransfers > 0)
	return
}

func mainMenuTelegramKeyboard(whc bots.WebhookContext, params mainMenuParams) *tgbotapi.ReplyKeyboardMarkup {
	firstRow := []string{
		_lendCommand.DefaultTitle(whc),
		_borrowCommand.DefaultTitle(whc),
	}

	if params.showReturn {
		firstRow = append(firstRow, _returnCommand.DefaultTitle(whc))
	}

	buttonRows := make([][]string, 0, 3)
	buttonRows = append(buttonRows, firstRow)

	if params.showBalanceAndHistory {
		buttonRows = append(buttonRows, []string{
			whc.CommandText(trans.COMMAND_TEXT_BALANCE, emoji.BALANCE_ICON),
			whc.CommandText(trans.COMMAND_TEXT_HISTORY, emoji.HISTORY_ICON),
		})
	}

	buttonRows = append(buttonRows, []string{
		whc.CommandText(trans.COMMAND_TEXT_SETTING, emoji.SETTINGS_ICON),
		whc.CommandText(trans.COMMAND_TEXT_FEEDBACK, emoji.BULB_ICON),
		whc.CommandText(trans.COMMAND_TEXT_HELP, emoji.HELP_ICON),
	})

	return tgbotapi.NewReplyKeyboardUsingStrings(buttonRows)
}
func SetMainMenuKeyboard(whc bots.WebhookContext, m *bots.MessageFromBot) {
	params := getMainMenuParams(whc, true)
	switch whc.BotPlatform().Id() {
	case telegram_bot.TelegramPlatformID:
		m.Keyboard = mainMenuTelegramKeyboard(whc, params)
	case viber_bot.ViberPlatformID:
		m.Keyboard = mainMenuViberKeyboard(whc, params)
	case fbm_bot.FbmPlatformID:
		if m.Text != "" {
			panic("FBM does not support message text and attachments in the same request.")
		}
		m.FbmAttachment = mainMenuFbmAttachment(whc, params)
	default:
		panic("Unsupported platform id=" + whc.BotPlatform().Id())
	}
}

func mainMenuFbmAttachment(whc bots.WebhookContext, params mainMenuParams) *fbm_api.RequestAttachment {
	attachment := &fbm_api.RequestAttachment{
		Type: fbm_api.RequestAttachmentTypeTemplate,
		Payload: fbm_api.NewListTemplate(
			fbm_api.TopElementStyleCompact,
			fbm_api.NewRequestElementWithDefaultAction(
				"DebtsTracker.io",
				"Tracks personal debts (auto-reminders to your debtors)",
				fbm_api.NewDefaultActionWithWebUrl(fbm_api.RequestWebUrlAction{MessengerExtensions: true, Url: "https://debtstracker-dev1.appspot.com/app/?page=debts&lang=ru"}),
				fbm_api.NewRequestWebUrlButtonWithRatio(emoji.CURRENCY_EXCAHNGE_ICON+" Record new debt", "https://debtstracker-dev1.appspot.com/app/?page=new-debt&lang=ru", "full"),
			),
			fbm_api.NewRequestElementWithDefaultAction(
				"Current balance",
				"You owe $100",
				fbm_api.NewDefaultActionWithWebUrl(fbm_api.RequestWebUrlAction{MessengerExtensions: true, Url: "https://debtstracker-dev1.appspot.com/app/?page=debts&lang=ru"}),
				fbm_api.NewRequestWebUrlButtonWithRatio(emoji.BALANCE_ICON+" Record return", "https://debtstracker-dev1.appspot.com/app/?page=return&lang=ru", "full"),
			),
			fbm_api.NewRequestElementWithDefaultAction(
				"History",
				"Last transfer: $100 to Jack Smith",
				fbm_api.NewDefaultActionWithWebUrl(fbm_api.RequestWebUrlAction{MessengerExtensions: true, Url: "https://debtstracker-dev1.appspot.com/app/?page=history&lang=ru"}),
				fbm_api.NewRequestWebUrlButtonWithRatio(emoji.HISTORY_ICON+" View full history", "https://debtstracker-dev1.appspot.com/app/?page=history&lang=ru", "full"),
			),
			fbm_api.NewRequestElementWithDefaultAction(
				"Settings",
				"You can change language, notification preferences, etc.",
				fbm_api.NewDefaultActionWithWebUrl(fbm_api.RequestWebUrlAction{MessengerExtensions: true, Url: "https://debtstracker-dev1.appspot.com/app/?page=debts&lang=ru"}),
				fbm_api.NewRequestWebUrlButtonWithRatio(emoji.SETTINGS_ICON+" Edit my preferences", "https://debtstracker-dev1.appspot.com/app/?page=settings&lang=ru", "full"),
			),
		),
	}
	log.Debugf(whc.Context(), "First element: %v", attachment.Payload.RequestAttachmentListTemplate.Elements[0])
	return attachment
}

const (
	UTM_CAMPAIGN_BOT_MAIN_MENU = "bot-main-menu"
)

func mainMenuViberKeyboard(whc bots.WebhookContext, params mainMenuParams) *viberinterface.Keyboard {
	var buttons []viberinterface.Button
	lendingText := _lendCommand.DefaultTitle(whc)
	borrowText := _borrowCommand.DefaultTitle(whc)
	const (
		maxColumns = 6
		in3columns = maxColumns / 3
		in2columns = maxColumns / 2
	)
	if params.showReturn {
		returnText := _returnCommand.DefaultTitle(whc)
		buttons = []viberinterface.Button{
			{
				Columns:    in3columns,
				BgColor:    viber.ButtonBgColor,
				Text:       lendingText,
				ActionType: viberinterface.ActionTypeOpenUrl,
				ActionBody: common.GetNewDebtPageUrl(whc, models.TransferDirectionUser2Counterparty, UTM_CAMPAIGN_BOT_MAIN_MENU),
			},
			{
				Columns:    in3columns,
				BgColor:    viber.ButtonBgColor,
				Text:       borrowText,
				ActionType: viberinterface.ActionTypeOpenUrl,
				ActionBody: common.GetNewDebtPageUrl(whc, models.TransferDirectionCounterparty2User, UTM_CAMPAIGN_BOT_MAIN_MENU),
			},
			{Columns: in3columns, ActionBody: returnText, Text: returnText, BgColor: viber.ButtonBgColor},
		}
	} else {
		buttons = []viberinterface.Button{
			{Columns: in2columns, ActionBody: lendingText, Text: lendingText, BgColor: viber.ButtonBgColor},
			{Columns: in2columns, ActionBody: borrowText, Text: borrowText, BgColor: viber.ButtonBgColor},
		}
	}
	if params.showBalanceAndHistory {
		userID := whc.AppUserIntID()
		locale := whc.Locale()
		balanceUrl := common.GetBalanceUrlForUser(userID, locale, whc.BotPlatform().Id(), whc.GetBotCode())
		historyUrl := common.GetHistoryUrlForUser(userID, locale, whc.BotPlatform().Id(), whc.GetBotCode())
		buttons = append(buttons, []viberinterface.Button{
			{Columns: in2columns, ActionType: "open-url", ActionBody: balanceUrl, Text: whc.CommandText(trans.COMMAND_TEXT_BALANCE, emoji.BALANCE_ICON), BgColor: viber.ButtonBgColor},
			{Columns: in2columns, ActionType: "open-url", ActionBody: historyUrl, Text: whc.CommandText(trans.COMMAND_TEXT_HISTORY, emoji.HISTORY_ICON), BgColor: viber.ButtonBgColor},
		}...)
	}
	{ // Last row
		settings := whc.CommandText(trans.COMMAND_TEXT_SETTING, emoji.SETTINGS_ICON)
		rate := whc.CommandText(trans.COMMAND_TEXT_HIGH_FIVE, emoji.STAR_ICON)
		help := whc.CommandText(trans.COMMAND_TEXT_HELP, emoji.HELP_ICON)
		buttons = append(buttons, []viberinterface.Button{
			{Columns: in3columns, ActionBody: settings, Text: settings, BgColor: viber.ButtonBgColor},
			{Columns: in3columns, ActionBody: rate, Text: rate, BgColor: viber.ButtonBgColor},
			{Columns: in3columns, ActionBody: help, Text: help, BgColor: viber.ButtonBgColor},
		}...)
	}

	return viberinterface.NewKeyboard(viber.KeyboardBgColor, false, buttons...)
}
