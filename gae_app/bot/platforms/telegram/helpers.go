package telegram

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"github.com/strongo/log"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"golang.org/x/net/context"
	"regexp"
	"strings"
)

func GetTelegramBotApiByBotCode(c context.Context, code string) *tgbotapi.BotAPI {
	if s, ok := _bots.ByCode[code]; ok {
		return tgbotapi.NewBotAPIWithClient(s.Token, dal.HttpClient(c))
	} else {
		return nil
	}
}

var reTelegramStartCommandPrefix = regexp.MustCompile(`/start(@\w+)?\s+`)

func ParseStartCommand(whc bots.WebhookContext) (startParam string, startParams []string) {
	input := whc.Input()

	switch input.(type) {
	case bots.WebhookTextMessage:
		startParam = input.(bots.WebhookTextMessage).Text()
	case bots.WebhookReferralMessage:
		startParam = input.(bots.WebhookReferralMessage).RefData()
	default:
		panic("Unknown input type")
	}
	if strings.HasPrefix(startParam, "/start") && startParam != "/start" {
		if loc := reTelegramStartCommandPrefix.FindStringIndex(startParam); loc != nil && len(loc) > 0 {
			startParam = startParam[loc[1]:]
			var utm_medium, utm_source string
			startParams = strings.Split(startParam, "__")
			for _, p := range startParams {
				switch {
				case strings.HasPrefix(p, "l="):
					code5 := p[len("l="):]
					if len(code5) == 5 {
						whc.SetLocale(code5)
						whc.ChatEntity().SetPreferredLanguage(code5)
					}
				case strings.HasPrefix(p, "utm_m="):
					utm_medium = p[len("utm_m="):]
				case strings.HasPrefix(p, "utm_s="):
					utm_source = p[len("utm_s="):]
				}
			}
			if utm_medium != "" || utm_source != "" { // TODO: Handle analytics
				log.Debugf(whc.Context(), "TODO: utm_medium=%v, utm_source=%v", utm_medium, utm_source)
			}
		} else {
			log.Errorf(whc.Context(), "reTelegramStartCommandPrefix did not match")
		}
		return
	}
	return
}
