package splitus

import (
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strings"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/app"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/log"
)

var reInlineQueryNewBill = regexp.MustCompile(`^\s*(\d+(?:\.\d*)?)([^\s]*)\s+(.+?)\s*$`)

var inlineQueryCommand = bots.Command{
	Code:       "inline-query",
	InputTypes: []bots.WebhookInputType{bots.WebhookInputInlineQuery},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		whc.LogRequest()
		c := whc.Context()
		if tgInput, ok := whc.Input().(telegram_bot.TelegramWebhookInput); ok {
			update := tgInput.TgUpdate()

			if user, err := whc.GetAppUser(); err != nil {
				return m, err
			} else if preferredLocale := user.PreferredLocale(); preferredLocale != "" {
				log.Debugf(c, "User has preferring locale")
				whc.SetLocale(preferredLocale)
			} else if tgLang := update.InlineQuery.From.LanguageCode; len(tgLang) >= 2 {
				switch strings.ToLower(tgLang[:2]) {
				case "ru":
					log.Debugf(c, "Telegram client has known language code")
					whc.SetLocale(strongo.LocaleRuRu.Code5)
				}
			}
		}
		inlineQuery := whc.Input().(bots.WebhookInlineQuery)
		query := strings.TrimSpace(inlineQuery.GetQuery())
		log.Debugf(c, "inlineQueryCommand.Action(query=%v)", query)
		switch {
		case query == "":
			return inlineEmptyQuery(whc, inlineQuery)
		case strings.HasPrefix(query, joinGroupCommanCode+"?id="):
			return inlineQueryJoinGroup(whc, query)
		default:
			if reMatches := reInlineQueryNewBill.FindStringSubmatch(query); reMatches != nil {
				return inlineQueryNewBill(whc, reMatches[1], reMatches[2], reMatches[3])
			}
			log.Debugf(c, "Inline query not matched to any action: [%v]", query)
		}

		return
	},
}

func inlineEmptyQuery(whc bots.WebhookContext, inlineQuery bots.WebhookInlineQuery) (m bots.MessageFromBot, err error) {
	log.Debugf(whc.Context(), "InlineEmptyQuery()")
	m.BotMessage = telegram_bot.InlineBotMessage(tgbotapi.InlineConfig{
		InlineQueryID:     inlineQuery.GetInlineQueryID(),
		CacheTime:         60,
		SwitchPMText:      "Help: How to use this bot?",
		SwitchPMParameter: "help_inline",
	})
	return
}

func inlineQueryJoinGroup(whc bots.WebhookContext, query string) (m bots.MessageFromBot, err error) {
	c := whc.Context()

	inlineQuery := whc.Input().(bots.WebhookInlineQuery)

	var group models.Group
	if group.ID = query[len(joinGroupCommanCode+"?id="):]; group.ID == "" {
		err = errors.New("Missing group ID")
		return
	}
	if group, err = dal.Group.GetGroupByID(c, group.ID); err != nil {
		return
	}

	joinBillInlineResult := tgbotapi.InlineQueryResultArticle{
		ID:          query,
		Type:        "article",
		Title:       "Send invite for joining",
		Description: "group: " + group.Name,
		InputMessageContent: tgbotapi.InputTextMessageContent{
			Text:      fmt.Sprintf("I'm inviting you to join <b>bills sharing</b> group @%v.", whc.GetBotCode()),
			ParseMode: "HTML",
		},
		ReplyMarkup: &tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{
					{
						Text:         "Join",
						CallbackData: query,
					},
				},
			},
		},
	}

	m.BotMessage = telegram_bot.InlineBotMessage(tgbotapi.InlineConfig{
		InlineQueryID: inlineQuery.GetInlineQueryID(),
		CacheTime:     60,
		Results: []interface{}{
			joinBillInlineResult,
		},
	})
	return
}

func inlineQueryNewBill(whc bots.WebhookContext, amountNum, amountCurr, billName string) (m bots.MessageFromBot, err error) {
	if len(amountCurr) == 3 {
		amountCurr = strings.ToUpper(amountCurr)
	}

	m.Text = fmt.Sprintf("Amount: %v %v, Bill name: %v", amountNum, amountCurr, billName)

	inlineQuery := whc.Input().(bots.WebhookInlineQuery)

	params := fmt.Sprintf("amount=%v&lang=%v", url.QueryEscape(amountNum+amountCurr), whc.Locale().Code5)

	resultID := "bill?" + params

	newBillInlineResult := tgbotapi.InlineQueryResultArticle{
		ID:          resultID,
		Type:        "article",
		Title:       fmt.Sprintf("%v: %v", whc.Translate(trans.COMMAND_TEXT_NEW_BILL), billName),
		Description: fmt.Sprintf("%v: %v %v", whc.Translate(trans.HTML_AMOUNT), amountNum, amountCurr),
		InputMessageContent: tgbotapi.InputTextMessageContent{
			Text: fmt.Sprintf("<b>%v</b>: %v %v - %v",
				whc.Translate(trans.MESAGE_TEXT_CREATING_BILL),
				html.EscapeString(amountNum),
				html.EscapeString(amountCurr),
				html.EscapeString(billName),
			),
			ParseMode: "HTML",
		},
		ReplyMarkup: &tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{
					{
						Text:         whc.Translate(trans.MESSAGE_TEXT_PLEASE_WAIT),
						CallbackData: "creating-bill?" + params,
					},
				},
			},
		},
	}

	m.BotMessage = telegram_bot.InlineBotMessage(tgbotapi.InlineConfig{
		InlineQueryID: inlineQuery.GetInlineQueryID(),
		CacheTime:     60,
		Results: []interface{}{
			newBillInlineResult,
		},
	})

	return
}
