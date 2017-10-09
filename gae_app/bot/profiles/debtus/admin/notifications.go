package admin

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bitbucket.com/asterus/debtstracker-server/gae_app/secret"
	"fmt"
	"github.com/strongo/bots-api-telegram"
	"golang.org/x/net/context"
	"google.golang.org/appengine/urlfetch"
)

func SendFeedbackToAdmins(c context.Context, feedback models.Feedback) (err error) {
	bot := tgbotapi.NewBotAPIWithClient(secret.AdminBotToken, urlfetch.Client(c))
	text := fmt.Sprintf("%v user #%d @%v (rate=%v):\n%v", feedback.CreatedOnPlatform, feedback.UserID, feedback.CreatedOnID, feedback.Rate, feedback.Text)
	message := tgbotapi.NewMessageToChannel("-1001128307094", text)
	message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			{Text: "Reply to feedback", URL: fmt.Sprintf("https://debtstracker.io/app/#/reply-to-feedback/%d", feedback.ID)},
		},
	)
	_, err = bot.Send(message)
	return
}
