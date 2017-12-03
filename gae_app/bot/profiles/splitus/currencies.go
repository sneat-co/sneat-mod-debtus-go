package splitus

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"net/url"
	"fmt"
	"github.com/strongo/db"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/trans"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/shared_group"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/shared_all"
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
	func(whc bots.WebhookContext, callbackUrl *url.URL, group models.Group) (m bots.MessageFromBot, err error) {
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

func groupSettingsSetCurrencyCommand(params shared_all.BotParams) bots.Command {
	return bots.Command{
		Code: GroupSettingsSetCurrencyCommandCode,
		CallbackAction: shared_all.TransactionalCallbackAction(db.CrossGroupTransaction, // TODO: Should be single group transaction, but a chat entity is loaded as well
			shared_group.NewGroupCallbackAction(func(whc bots.WebhookContext, callbackUrl *url.URL, group models.Group) (m bots.MessageFromBot, err error) {
				currency := models.Currency(callbackUrl.Query().Get(CURRENCY_PARAM_NAME))
				if group.DefaultCurrency != currency {
					group.DefaultCurrency = currency
					if err = dal.Group.SaveGroup(whc.Context(), group); err != nil {
						return
					}
				}
				if callbackUrl.Query().Get("start") == "y" {
					panic(`return params.InGroupWelcomeMessage(whc, group)`)
				} else {
					return GroupSettingsAction(whc, group, true)
				}
			})),
	}
}

