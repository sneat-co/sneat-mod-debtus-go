package dtb_transfer

import (
	"fmt"
	"strconv"
	"strings"

	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_general"
	dtb_common "bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/debtus/dtb_common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"errors"
	"github.com/strongo/log"
)

const NEW_COUNTERPARTY_COMMAND = "new-counterparty"

func NewCounterpartyCommand(nextCommand botsfw.Command) botsfw.Command {
	return botsfw.Command{
		Code:    NEW_COUNTERPARTY_COMMAND,
		Title:   trans.COMMAND_TEXT_NEW_COUNTERPARTY,
		Replies: []botsfw.Command{nextCommand},
		Action: func(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error) {
			c := whc.Context()

			chatEntity := whc.ChatEntity()
			if chatEntity.IsAwaitingReplyTo(NEW_COUNTERPARTY_COMMAND) {
				var user models.AppUser
				if user, err = dtb_common.GetUser(whc); err != nil {
					return
				}

				input := whc.Input()
				input.LogRequest()

				var contact models.Contact

				var (
					contactDetails  models.ContactDetails
					existingContact bool
				)

				switch input.(type) {
				case botsfw.WebhookTextMessage:
					webhookMessage := input.(botsfw.WebhookTextMessage)
					mt := strings.TrimSpace(webhookMessage.Text())
					if mt == "." {
						return dtb_general.MainMenuAction(whc, "", false)
					}
					if mt == "" {
						return m, errors.New("failed to get userContactJson details: mt is empty && inputMessage == nil")
					}
					if _, err = strconv.ParseFloat(mt, 64); err == nil {
						// User entered a number
						return whc.NewMessageByCode(trans.MESSAGE_TEXT_CONTACT_NAME_IS_NUMBER), nil
					}
					contactDetails = models.ContactDetails{
						Username: mt,
					}
				case botsfw.WebhookContactMessage:
					contactMessage := input.(botsfw.WebhookContactMessage)
					if contactMessage == nil {
						return m, errors.New("failed to get WebhookContactMessage: contactMessage == nil")
					}

					contactDetails = models.ContactDetails{
						FirstName: contactMessage.FirstName(),
						LastName:  contactMessage.LastName(),
						//Username: username,
					}
					phoneStr := contactMessage.PhoneNumber()
					if phoneNum, err := strconv.ParseInt(phoneStr, 10, 64); err != nil {
						log.Warningf(c, "Failed to parse phone string to int (%v)", phoneStr)
					} else {
						contactDetails.PhoneContact = models.PhoneContact{
							PhoneNumber:          phoneNum,
							PhoneNumberConfirmed: true,
						}
					}

					switch input.InputType() {
					case botsfw.WebhookInputContact:
						contactDetails.TelegramUserID = int64(contactMessage.UserID().(int)) // TODO: check we are on Telegram
						if contactDetails.TelegramUserID != 0 {
							for _, userContactJson := range user.Contacts() {
								if userContactJson.TgUserID == contactDetails.TelegramUserID {
									log.Debugf(c, "Matched contact my TelegramUserID=%d", contactDetails.TelegramUserID)
									existingContact = true
									contact.ID = userContactJson.ID
								}
							}
						}
					}
				default:
					err = fmt.Errorf("unknown input, expected text or contact message, got: %T", input)
					return
				}

				if !existingContact {
					var user models.AppUser
					if user, err = facade.User.GetUserByID(c, whc.AppUserIntID()); err != nil {
						return
					}

					contactFullName := contactDetails.FullName()

					for _, userContact := range user.Contacts() {
						if userContact.Name == contactFullName {
							m.Text = whc.Translate(trans.MESSAGE_TEXT_ALREADY_HAS_CONTACT_WITH_SUCH_NAME)
							return
						}
					}
				}

				if !existingContact {
					if contact, user, err = facade.CreateContact(c, whc.AppUserIntID(), contactDetails); err != nil {
						return m, err
					}
					ga := whc.GA()
					ga.Queue(ga.GaEventWithLabel(
						"contacts",
						"contact-created",
						fmt.Sprintf("user-%v", whc.AppUserIntID()),
					))
					if contact.PhoneNumber != 0 && contact.PhoneNumberConfirmed {
						ga.Queue(ga.GaEventWithLabel(
							"contacts",
							"contact-details-added",
							"phone-number",
						))
					}
				}
				if contact.ID == 0 {
					panic("contact.ID == 0")
				}
				chatEntity.AddWizardParam(WIZARD_PARAM_COUNTERPARTY, strconv.FormatInt(contact.ID, 10))
				return nextCommand.Action(whc)
				//m = whc.NewMessageByCode(fmt.Sprintf("Contact Created: %v", counterpartyKey))
			} else {
				m = whc.NewMessageByCode(trans.MESSAGE_TEXT_ASK_NEW_COUNTERPARTY_NAME)
				m.Format = botsfw.MessageFormatHTML
				m.Keyboard = tgbotapi.NewHideKeyboard(true)
				chatEntity.PushStepToAwaitingReplyTo(NEW_COUNTERPARTY_COMMAND)
			}
			return m, err
		},
	}
}
