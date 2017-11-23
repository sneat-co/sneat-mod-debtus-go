package dtb_invite

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_general"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bitbucket.com/asterus/debtstracker-server/gae_app/invites"
	"fmt"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/log"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"net/url"
	"strings"
)

var AskInviteAddressTelegramCommand = AskInviteAddress(string(models.InviteByTelegram), emoji.ROCKET_ICON, trans.COMMAND_TEXT_INVITE_BY_TELEGRAM, trans.MESSAGE_TEXT_INVITE_BY_TELEGRAM, trans.MESSAGE_TEXT_NO_CONTACT_RECEIVED)
var AskInviteAddressEmailCommand = AskInviteAddress(string(models.InviteByEmail), emoji.EMAIL_ICON, trans.COMMAND_TEXT_SEND_BY_EMAIL, trans.MESSAGE_TEXT_INVITE_BY_EMAIL, trans.MESSAGE_TEXT_INVALID_EMAIL)
var AskInviteAddressSmsCommand = AskInviteAddress(string(models.InviteBySms), emoji.PHONE_ICON, trans.COMMAND_TEXT_SEND_BY_SMS, trans.MESSAGE_TEXT_INVITE_BY_SMS, trans.MESSAGE_TEXT_INVALID_PHONE_NUMBER)

func AskInviteAddress(channel, icon, commandText, messageCode, invalidMessageCode string) bots.Command {
	code := fmt.Sprintf("ask-%v-address-for-invite", channel)
	return bots.Command{
		Code:  code,
		Icon:  icon,
		Title: commandText,
		Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
			chatEntity := whc.ChatEntity()

			if chatEntity.IsAwaitingReplyTo(code) {
				email := strings.TrimSpace(whc.Input().(bots.WebhookTextMessage).Text())
				isValid := channel == string(models.InviteByEmail) && strings.Contains(email, "@") && strings.Contains(email, ".")
				if isValid {
					invite, err := dal.Invite.CreatePersonalInvite(whc, whc.AppUserIntID(), models.InviteByEmail, email, whc.BotPlatform().Id(), whc.GetBotCode(), "counterparty=?")
					if err != nil {
						log.Errorf(whc.Context(), "Failed to call invites.CreateInvite()")
						return m, err
					}
					var emailID string
					emailID, err = invites.SendInviteByEmail(
						whc.ExecutionContext(),
						whc.GetSender().GetFirstName(),
						"alex@debtstracker.io",
						"Stranger",
						invite.ID,
						whc.GetBotCode(),
						common.UtmSourceFromContext(whc),
					)
					if err != nil {
						return m, err
					}
					m = whc.NewMessageByCode(trans.MESSAGE_TEXT_INVITE_CREATED, emailID)
				} else {
					m = whc.NewMessageByCode(invalidMessageCode)
					m.Keyboard = tgbotapi.NewReplyKeyboardUsingStrings([][]string{
						{whc.Translate(trans.COMMAND_TEXT_MISTYPE_WILL_TRY_AGAIN)},
						{whc.Translate(trans.COMMAND_TEXT_OTHER_WAYS_TO_SEND_INVITE)},
						{dtb_general.MainMenuCommand.DefaultTitle(whc)},
					})
				}
			} else {
				m = whc.NewMessageByCode(messageCode)
				chatEntity.PushStepToAwaitingReplyTo(code)
			}
			return m, nil
		},
	}
}

var AskInviteAddressCallbackCommand = bots.Command{
	Code: "invite",
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		q := callbackUrl.Query()
		echoSelection := func(mt string) error {
			if m, err = whc.NewEditMessage(whc.Translate(trans.MESSAGE_TEXT_ABOUT_INVITES)+"\n\n"+mt, bots.MessageFormatHTML); err != nil {
				return err
			}
			_, err := whc.Responder().SendMessage(whc.Context(), m, bots.BotApiSendMessageOverHTTPS)
			return errors.Wrap(err, "Failed to edit callback message")
		}
		_ = whc.ChatEntity() // To switch locale
		switch q.Get("by") {
		case string(models.InviteByEmail):
			if err = echoSelection(whc.Translate(trans.MESSAGE_TEXT_YOU_SELECTED_INVITE_BY_EMAIL)); err != nil {
				return
			}
			return AskInviteAddressEmailCommand.Action(whc)
		case string(models.InviteBySms):
			if err = echoSelection(whc.Translate(trans.MESSAGE_TEXT_YOU_SELECTED_INVITE_BY_SMS)); err != nil {
				return
			}
			return AskInviteAddressSmsCommand.Action(whc)
		case "":
			log.Warningf(whc.Context(), "AskInviteAddressCallbackCommand: got request to create invite without specifying a channel - not implemented yet. Need to ask a channel first. Check how it works if message forwarded to secret chat.")
			m.Text = whc.Translate(trans.MESSAGE_TEXT_NOT_IMPLEMENTED_YET)
			return
		default:
			err = fmt.Errorf("unknown invite channel: %v", q.Get("by"))
			return
		}
	},
}
