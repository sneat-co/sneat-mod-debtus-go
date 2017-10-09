package bot_shared

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bytes"
	"github.com/strongo/app"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"golang.org/x/net/context"
)

type BotParams struct {
	GetGroupBillCardInlineKeyboard   func(translator strongo.SingleLocaleTranslator, bill models.Bill) *tgbotapi.InlineKeyboardMarkup
	GetPrivateBillCardInlineKeyboard func(translator strongo.SingleLocaleTranslator, botCode string, bill models.Bill) *tgbotapi.InlineKeyboardMarkup
	OnAfterBillCurrencySelected      func(translator strongo.SingleLocaleTranslator, billID string) *tgbotapi.InlineKeyboardMarkup
	DelayUpdateBillCardOnUserJoin    func(c context.Context, billID string, message string) error
	ShowGroupMembers                 func(whc bots.WebhookContext, group models.Group, isEdit bool) (m bots.MessageFromBot, err error)
	WelcomeText                      func(translator strongo.SingleLocaleTranslator, buf *bytes.Buffer)
	InGroupWelcomeMessage            func(whc bots.WebhookContext) (m bots.MessageFromBot, err error)
	InBotWelcomeMessage              func(whc bots.WebhookContext) (m bots.MessageFromBot, err error)
}
