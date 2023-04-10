package splitus

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/shared_all"
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/shared_group"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"fmt"
	"github.com/crediterra/money"
	"github.com/strongo/log"
	"net/url"
)

const CURRENCY_PARAM_NAME = "currency"

func currenciesInlineKeyboard(callbackDataPrefix string, more ...[]tgbotapi.InlineKeyboardButton) *tgbotapi.InlineKeyboardMarkup {
	currencyButton := func(code, flag string) tgbotapi.InlineKeyboardButton {
		btn := tgbotapi.InlineKeyboardButton{CallbackData: callbackDataPrefix + "&" + CURRENCY_PARAM_NAME + "=" + code}
		if flag == "" {
			btn.Text = code
		} else {
			btn.Text = flag + " " + code
		}
		return btn
	}

	usdRow := []tgbotapi.InlineKeyboardButton{
		currencyButton("USD", "🇺🇸"),
		currencyButton("AUD", "🇦🇺"),
		currencyButton("CAD", "🇨🇦"),
		currencyButton("GBP", "🇬🇧"),
	}

	eurRow := []tgbotapi.InlineKeyboardButton{
		currencyButton("EUR", "🇪🇺"),
		currencyButton("CHF", "🇨🇭"),
		currencyButton("NOK", "🇳🇴"),
		currencyButton("SEK", "🇸🇪"),
	}

	eurRow2 := []tgbotapi.InlineKeyboardButton{
		currencyButton("BGN", "🇧🇬"),
		currencyButton("HUF", "🇭🇺"),
		currencyButton("PLN", "🇵🇱"),
		currencyButton("RON", "🇷🇴"),
	}

	rubRow := []tgbotapi.InlineKeyboardButton{
		currencyButton("RUB", "🇷🇺"),
		currencyButton("BYN", "🇧🇾"),
		currencyButton("UAH", "🇺🇦"),
		currencyButton("MDL", "🇲🇩"),
	}

	exUSSR := []tgbotapi.InlineKeyboardButton{
		currencyButton("KGS", "🇰🇬"),
		currencyButton("KZT", "🇰🇿"),
		currencyButton("TJS", "🇹🇯"),
		currencyButton("UZS", "🇺🇿"),
	}

	asiaRow := []tgbotapi.InlineKeyboardButton{
		currencyButton("CNY", "🇨🇳"),
		currencyButton("JPY", "🇯🇵"),
		currencyButton("IDR", "🇮🇩"),
		currencyButton("KRW", "🇰🇷"),
		//currencyButton("VND", "🇻🇳"),
	}

	keyboard := append([][]tgbotapi.InlineKeyboardButton{
		usdRow,
		eurRow,
		rubRow,
		exUSSR,
		eurRow2,
		asiaRow,
	}, more...)

	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}
}

const (
	GroupSettingsChooseCurrencyCommandCode = "grp-stngs-chs-ccy"
	GroupSettingsSetCurrencyCommandCode    = "grp-stngs-set-ccy"
)

var groupSettingsChooseCurrencyCommand = shared_group.GroupCallbackCommand(GroupSettingsChooseCurrencyCommandCode,
	func(whc botsfw.WebhookContext, callbackUrl *url.URL, group models.Group) (m botsfw.MessageFromBot, err error) {
		m.IsEdit = true
		m.Text = whc.Translate(trans.MESSAGE_TEXT_ASK_PRIMARY_CURRENCY)
		m.Keyboard = currenciesInlineKeyboard(
			GroupSettingsSetCurrencyCommandCode+"?group="+group.ID,
			[]tgbotapi.InlineKeyboardButton{
				{
					Text: whc.Translate(trans.BT_OTHER_CURRENCY),
					URL:  fmt.Sprintf("https://t.me/%v?start=", whc.GetBotCode()) + GroupSettingsChooseCurrencyCommandCode,
				},
			},
		)
		return
	},
)

func groupSettingsSetCurrencyCommand(params shared_all.BotParams) botsfw.Command {
	return botsfw.Command{
		Code: GroupSettingsSetCurrencyCommandCode,
		CallbackAction: shared_group.NewGroupCallbackAction(func(whc botsfw.WebhookContext, callbackUrl *url.URL, group models.Group) (m botsfw.MessageFromBot, err error) {
			currency := money.Currency(callbackUrl.Query().Get(CURRENCY_PARAM_NAME))
			if group.DefaultCurrency != currency {
				c := whc.Context()
				if err := dtdal.DB.RunInTransaction(c, func(c context.Context) (err error) {
					if group, err = dtdal.Group.GetGroupByID(c, group.ID); err != nil {
						return
					}
					if group.DefaultCurrency != currency {
						group.DefaultCurrency = currency
						if err = dtdal.Group.SaveGroup(c, group); err != nil {
							return
						}
					}
					return
				}, db.SingleGroupTransaction); err != nil {
					log.Errorf(whc.Context(), "failed to change group default currency: %v", err)
				} else {
					log.Debugf(c, "Default currency for group %v updated to: %v", group.ID, currency)
				}
			}
			if callbackUrl.Query().Get("start") == "y" {
				return onStartCallbackInGroup(whc, group)
			} else {
				return GroupSettingsAction(whc, group, true)
			}
		}),
	}
}

func onStartCallbackInGroup(whc botsfw.WebhookContext, group models.Group) (m botsfw.MessageFromBot, err error) {
	// This links Telegram ChatID and ChatInstance
	panic("not implemeted")
	// if twhc, ok := whc.(*telegram.tgWebhookContext); ok {
	// 	if err = twhc.CreateOrUpdateTgChatInstance(); err != nil {
	// 		return
	// 	}
	// }
	// return inGroupWelcomeMessage(whc, group)
}

func inGroupWelcomeMessage(whc botsfw.WebhookContext, group models.Group) (m botsfw.MessageFromBot, err error) {
	m, err = GroupSettingsAction(whc, group, false)
	if err != nil {
		return
	}
	if _, err = whc.Responder().SendMessage(whc.Context(), m, botsfw.BotAPISendMessageOverHTTPS); err != nil {
		return
	}

	return whc.NewEditMessage(whc.Translate(trans.MESSAGE_TEXT_HI)+
		"\n\n"+whc.Translate(trans.SPLITUS_TEXT_HI_IN_GROUP)+
		"\n\n"+whc.Translate(trans.SPLITUS_TEXT_ABOUT_ME_AND_CO),
		bots.MessageFormatHTML)
}
