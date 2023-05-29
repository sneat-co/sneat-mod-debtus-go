package splitus

import (
	"fmt"
	"github.com/bots-go-framework/bots-api-telegram/tgbotapi"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/sneat-co/debtstracker-translations/trans"
	"github.com/strongo/i18n"
	"html"
	"net/url"
	"regexp"
	"strings"

	"errors"
	"github.com/bots-go-framework/bots-fw-telegram"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/dtdal"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/models"
	"github.com/strongo/log"
)

var reInlineQueryNewBill = regexp.MustCompile(`^\s*(\d+(?:\.\d*)?)([^\s]*)\s+(.+?)\s*$`)

var inlineQueryCommand = botsfw.Command{
	Code:       "inline-query",
	InputTypes: []botsfw.WebhookInputType{botsfw.WebhookInputInlineQuery},
	Action: func(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error) {
		whc.LogRequest()
		c := whc.Context()
		if tgInput, ok := whc.Input().(telegram.TgWebhookInput); ok {
			update := tgInput.TgUpdate()

			if user, err := whc.AppUserData(); err != nil {
				return m, err
			} else if preferredLocale := user.GetPreferredLocale(); preferredLocale != "" {
				log.Debugf(c, "User has preferring locale")
				_ = whc.SetLocale(preferredLocale)
			} else if tgLang := update.InlineQuery.From.LanguageCode; len(tgLang) >= 2 {
				switch strings.ToLower(tgLang[:2]) {
				case "ru":
					log.Debugf(c, "Telegram client has known language code")
					if err = whc.SetLocale(i18n.LocaleRuRu.Code5); err != nil {
						return m, err
					}
				}
			}
		}
		inlineQuery := whc.Input().(botsfw.WebhookInlineQuery)
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

func inlineEmptyQuery(whc botsfw.WebhookContext, inlineQuery botsfw.WebhookInlineQuery) (m botsfw.MessageFromBot, err error) {
	log.Debugf(whc.Context(), "InlineEmptyQuery()")
	m.BotMessage = telegram.InlineBotMessage(tgbotapi.InlineConfig{
		InlineQueryID:     inlineQuery.GetInlineQueryID(),
		CacheTime:         60,
		SwitchPMText:      "Help: How to use this bot?",
		SwitchPMParameter: "help_inline",
	})
	return
}

func inlineQueryJoinGroup(whc botsfw.WebhookContext, query string) (m botsfw.MessageFromBot, err error) {
	c := whc.Context()

	inlineQuery := whc.Input().(botsfw.WebhookInlineQuery)

	var group models.Group
	if group.ID = query[len(joinGroupCommanCode+"?id="):]; group.ID == "" {
		err = errors.New("Missing group ID")
		return
	}
	if group, err = dtdal.Group.GetGroupByID(c, nil, group.ID); err != nil {
		return
	}

	joinBillInlineResult := tgbotapi.InlineQueryResultArticle{
		ID:          query,
		Type:        "article",
		Title:       "Send invite for joining",
		Description: "group: " + group.Data.Name,
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

	m.BotMessage = telegram.InlineBotMessage(tgbotapi.InlineConfig{
		InlineQueryID: inlineQuery.GetInlineQueryID(),
		CacheTime:     60,
		Results: []interface{}{
			joinBillInlineResult,
		},
	})
	return
}

func inlineQueryNewBill(whc botsfw.WebhookContext, amountNum, amountCurr, billName string) (m botsfw.MessageFromBot, err error) {
	if len(amountCurr) == 3 {
		amountCurr = strings.ToUpper(amountCurr)
	}

	m.Text = fmt.Sprintf("Amount: %v %v, Bill name: %v", amountNum, amountCurr, billName)

	inlineQuery := whc.Input().(botsfw.WebhookInlineQuery)

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

	m.BotMessage = telegram.InlineBotMessage(tgbotapi.InlineConfig{
		InlineQueryID: inlineQuery.GetInlineQueryID(),
		CacheTime:     60,
		Results: []interface{}{
			newBillInlineResult,
		},
	})

	return
}
