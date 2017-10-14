package bot_shared

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"net/url"
	"github.com/strongo/app/log"
)

const SET_BILL_CURRENCY_COMMAND = "set-bill-currency"

func setBillCurrencyCommand(params BotParams) bots.Command {
	return transactionalCallbackCommand(BillCallbackCommand(SET_BILL_CURRENCY_COMMAND,
		func(whc bots.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m bots.MessageFromBot, err error) {
			c := whc.Context()
			log.Debugf(c, "setBillCurrencyCommand.CallbackAction()")
			query := callbackUrl.Query()
			currencyCode := models.Currency(query.Get("currency"))
			if bill.Currency != currencyCode {
				previousCurrency := bill.Currency
				bill.Currency = currencyCode
				if err = dal.Bill.SaveBill(c, bill); err != nil {
					return
				}

				if bill.UserGroupID() != "" {
					var group models.Group
					if group, err = dal.Group.GetGroupByID(c, bill.UserGroupID()); err != nil {
						return
					}
					diff := bill.GetBalance().BillBalanceDifference(make(models.BillBalanceByMember,0))
					if _, err = group.ApplyBillBalanceDifference(bill.Currency, diff); err != nil {
						return
					}
					if previousCurrency != "" {
						if _, err = group.ApplyBillBalanceDifference(previousCurrency, diff.Reverse()); err != nil {
							return
						}
					}
					if dal.Group.SaveGroup(c, group); err != nil {
						return
					}
				}
			}
			if m.Text, err = GetBillCardMessageText(c, whc.GetBotCode(), whc, bill, true, whc.Translate(trans.MESSAGE_TEXT_BILL_ASK_WHO_PAID)); err != nil {
				return
			}
			m.Format = bots.MessageFormatHTML
			m.Keyboard = params.OnAfterBillCurrencySelected(whc, bill.ID)
			m.IsEdit = true

			return
		},
	), dal.CrossGroupTransaction)
}

func CurrenciesInlineKeyboard(callbackDataPrefix string) *tgbotapi.InlineKeyboardMarkup {
	currencyButton := func(code, flag string) tgbotapi.InlineKeyboardButton {
		btn := tgbotapi.InlineKeyboardButton{CallbackData: callbackDataPrefix + "&currency=" + code}
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

	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			usdRow,
			eurRow,
			rubRow,
			exUSSR,
			eurRow2,
			asiaRow,
		},
	}
}
