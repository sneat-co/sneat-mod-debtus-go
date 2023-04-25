package admin

import (
	"context"
	"fmt"
	"github.com/bots-go-framework/bots-api-telegram/tgbotapi"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/models"
	"google.golang.org/appengine/urlfetch"
)

func SendFeedbackToAdmins(c context.Context, botToken string, feedback models.Feedback) (err error) {
	bot := tgbotapi.NewBotAPIWithClient(botToken, urlfetch.Client(c))
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
