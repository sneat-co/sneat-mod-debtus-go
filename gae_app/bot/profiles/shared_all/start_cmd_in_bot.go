package shared_all

import (
	"bytes"
	"strings"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"errors"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
)

var ErrUnknownStartParam = errors.New("unknown start parameter")

func startInBotAction(whc bots.WebhookContext, startParams []string, botParams BotParams) (m bots.MessageFromBot, err error) {
	log.Debugf(whc.Context(), "startInBotAction() => startParams: %v", startParams)
	if m, err = botParams.StartInBotAction(whc, startParams); err == nil || err != ErrUnknownStartParam {
		return
	}
	if err == ErrUnknownStartParam {
		if whc.ChatEntity().GetPreferredLanguage() == "" {
			return onboardingAskLocaleAction(whc, whc.Translate(trans.MESSAGE_TEXT_HI)+"\n\n", botParams)
		}
	}
	err = nil
	if len(startParams) > 0 {
		switch {
		case strings.HasPrefix(startParams[0], "how-to"):
			return howToCommand.Action(whc)
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

	buf.WriteString(whc.Translate(trans.SPLITUS_TEXT_HI))
	buf.WriteString("\n\n")
	buf.WriteString(whc.Translate(trans.SPLITUS_TEXT_ABOUT_ME_AND_CO))

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
