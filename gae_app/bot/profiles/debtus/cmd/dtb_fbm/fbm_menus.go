package dtb_fbm

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/platforms/fbm"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_transfer"
	"fmt"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app"
	"github.com/strongo/bots-api-fbm"
	"github.com/strongo/bots-framework/core"
)

func aboutCard(whc bots.WebhookContext) fbm_api.RequestElement {
	baseUrl := fbmAppBaseUrl(whc)
	return fbm_api.NewRequestElementWithDefaultAction(
		"More...",
		"What can I do for you?",
		newDefaultUrlAction(baseUrl, ""),
		newUrlButton(emoji.HELP_ICON, "Help", baseUrl, "#help"),
		newUrlButton(emoji.CONTACTS_ICON, "Contacts", baseUrl, "#contacts"),
		newUrlButton(emoji.HISTORY_ICON, "History", baseUrl, "#history"),
	)
}

func linkAccountsCard(whc bots.WebhookContext) fbm_api.RequestElement {
	baseUrl := fbmAppBaseUrl(whc)
	return fbm_api.NewRequestElementWithDefaultAction(
		"link Accounts",
		"to to...",
		newDefaultUrlAction(baseUrl, ""),
		newUrlButton(emoji.ROCKET_ICON, "Telegram", "t.me/", ""),
	)
}

func mainMenuCard(whc bots.WebhookContext) fbm_api.RequestElement {
	baseUrl := fbmAppBaseUrl(whc)
	return fbm_api.NewRequestElementWithDefaultAction(
		"Welcome",
		"This is an app to split bills and track debt records.",
		newDefaultUrlAction(baseUrl, ""),
		newPostbackButton(emoji.MEMO_ICON, "Debts", FbmDebtsCommand.Code),
		newPostbackButton(emoji.BILLS_ICON, "Bills", FbmBillsCommand.Code),
		newPostbackButton(emoji.SETTINGS_ICON, "Settings", "fbm-settings"),
	)
}

//func mainMenuCard(whc bots.WebhookContext) fbm_api.RequestAttachmentPayload {
//	//baseUrl := fbmAppBaseUrl(whc)
//	return &fbm_api.NewListTemplate(
//		fbm_api.TopElementStyleCompact,
//		fbm_api.NewRequestElementWithDefaultAction(
//			emoji.MEMO_ICON + EM_SPACE + "Debts",
//			"Track your debts",
//			fbm_api.RequestDefaultAction{
//				Type: fbm_api.
//				newPostbackButton(emoji.MEMO_ICON, "Debts", FbmDebtsCommand.Code)
//			},
//		),
//	)
//}

func askLanguageCard(whc bots.WebhookContext) fbm_api.RequestAttachmentPayload {
	fbm_api.NewButtonTemplate(
		"",
	)
	requestElement := fbm_api.RequestElement{
		Title:    whc.Translate(trans.MESSAGE_TEXT_HI),
		Subtitle: "Please choose your language:",
	}
	for _, lang := range []strongo.Locale{strongo.LocaleEnUS, strongo.LocaleRuRu} {
		requestElement.Buttons = append(requestElement.Buttons, newPostbackButton(lang.FlagIcon, lang.NativeTitle, "fbm-set-lang?code5="+lang.Code5))
	}
	requestElement.Buttons = append(requestElement.Buttons, newUrlButton("", "More...", fbmAppBaseUrl(whc), "#set-locale"))
	return fbm_api.NewGenericTemplate(requestElement)
}

func welcomeCard(whc bots.WebhookContext) fbm_api.RequestElement {
	baseUrl := fbmAppBaseUrl(whc)
	return fbm_api.NewRequestElementWithDefaultAction(
		"Welcome!",
		"Have you ever used DebtsTracker.io app/bot outside of FB Messenger before?",
		newDefaultUrlAction(baseUrl, ""),
		newPostbackButton(emoji.MEMO_ICON, "Have not used", "fbm-debts"),
		newPostbackButton(emoji.ROCKET_ICON, "Used @ https://debtstracker.io/", "fbm-bills"),
		newPostbackButton(emoji.ROBOT_ICON, "Used @ Telegram", "fbm-settings"),
	)
}

func debtsCard(whc bots.WebhookContext) fbm_api.RequestElement {
	baseUrl := fbmAppBaseUrl(whc)
	requestElement := fbm_api.NewRequestElementWithDefaultAction(
		"Debts",
		"Tracks personal debts (auto-reminders to your debtors)",
		newDefaultUrlAction(baseUrl, "#debts"),
		newPostbackButton(emoji.MEMO_ICON, whc.Translate("New record"), "new-debt-or-return"),
		newPostbackButton(emoji.CLIPBOARD_ICON, whc.Translate(trans.COMMAND_TEXT_BALANCE), dtb_transfer.BALANCE_COMMAND),
		newPostbackButton(emoji.HISTORY_ICON, whc.Translate(trans.COMMAND_TEXT_HISTORY), dtb_transfer.HISTORY_COMMAND),
	)
	//requestElement.ImageUrl = ""
	return requestElement
}

func billsCard(whc bots.WebhookContext) fbm_api.RequestElement {
	baseUrl := fbmAppBaseUrl(whc)
	return fbm_api.NewRequestElementWithDefaultAction(
		"Bills",
		"Split regular or single bills and get paid back",
		newDefaultUrlAction(baseUrl, "#bills"),
		newUrlButton(emoji.DIVIDE_ICON, "Split bill", baseUrl, "#split-bill"),
		newUrlButton(emoji.BILLS_ICON, "Outstanding bills", baseUrl, "#bills"),
		newUrlButton(emoji.CALENDAR_ICON, "Recurring bills", baseUrl, "#bills"),
	)
}

func settingsCard(whc bots.WebhookContext) fbm_api.RequestElement {
	baseUrl := fbmAppBaseUrl(whc)
	return fbm_api.NewRequestElementWithDefaultAction(
		"Settings",
		"Adjust settings",
		newDefaultUrlAction(baseUrl, "#bills"),
		newUrlButton(emoji.BILLS_ICON, "Bills", baseUrl, "#bills"),
		newUrlButton(emoji.MEMO_ICON, "Split bill", baseUrl, "#split-bill"),
	)
}

func fbmAppBaseUrl(whc bots.WebhookContext) string {
	fbApp, host, err := fbm.GetFbAppAndHost(whc.Request())
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("https://%v/app/#fbm%v", host, fbApp.AppId)
}

func newDefaultUrlAction(baseUrl, hash string) fbm_api.RequestDefaultAction {
	return fbm_api.NewDefaultActionWithWebUrl(
		fbm_api.RequestWebUrlAction{
			MessengerExtensions: true,
			Url:                 baseUrl + hash,
		},
	)
}

func newUrlButton(icon, title, baseUrl, hash string) fbm_api.RequestButton {
	if icon != "" {
		title = icon + EM_SPACE + title
	}
	button := fbm_api.NewRequestWebUrlButtonWithRatio(
		title,
		baseUrl+hash,
		fbm_api.WebviewHeightRatioFull,
	)
	button.MessengerExtensions = true
	return button
}

func newPostbackButton(icon, title, payload string) fbm_api.RequestButton {
	if icon != "" {
		title = icon + EM_SPACE + title
	}
	button := fbm_api.NewRequestPostbackButton(
		title,
		payload,
	)
	return button
}
