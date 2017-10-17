package bot_shared

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bytes"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app/log"
	"github.com/strongo/bots-framework/core"
	"strings"
)

func startInBotAction(whc bots.WebhookContext, startParams []string, botParams BotParams) (m bots.MessageFromBot, err error) {
	log.Debugf(whc.Context(), "startInBotAction() => startParams: %v", startParams)
	if len(startParams) > 0 {
		switch {
		case strings.HasPrefix(startParams[0], "bill-"):
			return startBillAction(whc, startParams[0], botParams)
		case strings.HasPrefix(startParams[0], "how-to"):
			return howToCommand.Action(whc)
		case startParams[0] == SETTLE_GROUP_ASK_FOR_COUNTERPARTY_COMMAND:
			return settleGroupStartAction(whc, startParams[1:])
		}
	}

	return startInBotWelcomeAction(whc, botParams)
}

func startInBotWelcomeAction(whc bots.WebhookContext, botParams BotParams) (m bots.MessageFromBot, err error) {
	var user *models.AppUserEntity
	if user, err = GetUser(whc); err != nil {
		return
	}

	buf := new(bytes.Buffer)

	buf.WriteString(whc.Translate(trans.MESSAGE_TEXT_HI_USERNAME, user.FirstName))
	buf.WriteString(" ")

	botParams.WelcomeText(whc, buf)

	buf.WriteString("\n\n")
	buf.WriteString(whc.Translate(trans.MESSAGE_TEXT_ASK_LANG))
	m.Text = buf.String()

	m.Format = bots.MessageFormatHTML
	m.Keyboard = LangKeyboard
	return
}

func onStartCallbackInBot(whc bots.WebhookContext, params BotParams) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	log.Debugf(c, "onStartCallbackInBot()")

	if m, err = params.InBotWelcomeMessage(whc); err != nil {
		return
	}

	return
}


