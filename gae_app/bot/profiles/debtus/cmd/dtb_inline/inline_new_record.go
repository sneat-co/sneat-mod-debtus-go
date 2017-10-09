package dtb_inline

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/app/log"
	"regexp"
	"strings"
	"net/url"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/decimal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"html"
	"fmt"
	"github.com/strongo/bots-framework/platforms/telegram"
)

var ReInlineQueryAmount = regexp.MustCompile(`^\s*(\d+(?:\.\d*)?)\s*((?:\b|\B).+?)?\s*$`)

func InlineNewRecord(whc bots.WebhookContext, amountMatches []string) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	log.Debugf(c, "InlineNewRecord()")

	inlineQuery := whc.Input().(bots.WebhookInlineQuery)
	var (
		amountValue    decimal.Decimal64p2
		amountCurrency models.Currency
	)
	if amountValue, err = decimal.ParseDecimal64p2(strings.TrimRight(amountMatches[1], ".")); err != nil {
		return
	}
	currencyCode := strings.TrimRight(amountMatches[2], ".,;()[]{} ")
	log.Debugf(c, "currencyCode: %v", currencyCode)
	if currencyCode != "" {
		if len(currencyCode) > 20 {
			currencyCode = currencyCode[:20]
		}
		ccLow := strings.ToLower(currencyCode)
		if ccLow == models.RUR_SIGN || ccLow == "р" || ccLow == "руб" || ccLow == "рубля" || ccLow == "рублей" || ccLow == "rub" || ccLow == "rubles" || ccLow == "ruble" || ccLow == "rubley" {
			amountCurrency = models.CURRENCY_RUB
		} else if ccLow == "eur" || ccLow == "euro" || ccLow == models.EUR_SIGN {
			amountCurrency = models.CURRENCY_EUR
		} else if ccLow == "гривна" || ccLow == "гривен" || ccLow == "г" || ccLow == models.UAH_SIGN {
			amountCurrency = models.CURRENCY_UAH
		} else if ccLow == "тенге" || ccLow == "теңге"  || ccLow == "т" || ccLow == models.KZT_SIGN {
			amountCurrency = models.CURRENCY_KZT
		} else {
			amountCurrency = models.Currency(currencyCode)
		}
	} else {
		amountCurrency = models.CURRENCY_USD
	}

	amountText := html.EscapeString(models.NewAmount(amountCurrency, amountValue).String())

	newBillCallbackData := fmt.Sprintf("new-bill?v=%v&c=%v", amountMatches[1], url.QueryEscape(string(amountCurrency)))
	m.BotMessage = telegram_bot.InlineBotMessage(tgbotapi.InlineConfig{
		InlineQueryID: inlineQuery.GetInlineQueryID(),
		Results: []interface{}{
			tgbotapi.InlineQueryResultArticle{
				ID:          "SplitBill_" + whc.Locale().Code5,
				Type:        "article",
				Title:       "🛒 " + whc.Translate(trans.ARTICLE_TITLE_SPLIT_BILL),
				Description: whc.Translate(trans.ARTICLE_SUBTITLE_SPLIT_BILL, amountText),
				InputMessageContent: tgbotapi.InputTextMessageContent{
					Text:                  whc.Translate(trans.MESSAGE_TEXT_BILL_HEADER, amountText),
					ParseMode:             "HTML",
					DisableWebPagePreview: true,
				},
				ReplyMarkup: &tgbotapi.InlineKeyboardMarkup{
					InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
						{
							{Text: whc.Translate(trans.COMMAND_TEXT_I_PAID), CallbackData: newBillCallbackData + "&i=paid"},
							{Text: whc.Translate(trans.COMMAND_TEXT_I_OWE), CallbackData: newBillCallbackData + "&i=owe"},
						},
					},
				},
			},
			tgbotapi.InlineQueryResultArticle{
				ID:          "NewDebt_" + whc.Locale().Code5,
				Type:        "article",
				Title:       "💵 " + whc.Translate(trans.ARTICLE_NEW_DEBT_TITLE),
				Description: whc.Translate(trans.ARTICLE_NEW_DEBT_SUBTITLE, amountText),
				InputMessageContent: tgbotapi.InputTextMessageContent{
					Text:                  whc.Translate(trans.MESSAGE_TEXT_NEW_DEBT_HEADER, amountText),
					ParseMode:             "HTML",
					DisableWebPagePreview: true,
				},
				ReplyMarkup: &tgbotapi.InlineKeyboardMarkup{
					InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
						{
							{Text: whc.Translate(trans.COMMAND_TEXT_I_OWE), CallbackData: "i-owed?debt=new"},
							{Text: whc.Translate(trans.COMMAND_TEXT_OWED_TO_ME), CallbackData: "owed2me?debt=new"},
						},
					},
				},
			},
		},
	})
	return m, err
}
