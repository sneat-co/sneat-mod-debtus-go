package dtb_transfer

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_general"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/sms"
	//"bitbucket.org/asterus/debtstracker-server/gae_app/invites"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/analytics"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/general"
	"context"
	"errors"
	"github.com/bots-go-framework/bots-fw-telegram"
	"github.com/sneat-co/debtstracker-translations/emoji"
	"github.com/strongo/app"
	"github.com/strongo/log"
)

const ASK_PHONE_NUMBER_FOR_RECEIPT_COMMAND = "ask-phone-number-for-receipt"

func cleanPhoneNumber(phoneNumebr string) string {
	phoneNumebr = strings.Replace(phoneNumebr, " ", "", -1)
	phoneNumebr = strings.Replace(phoneNumebr, "(", "", -1)
	phoneNumebr = strings.Replace(phoneNumebr, ")", "", -1)
	return phoneNumebr
}

var AskPhoneNumberForReceiptCommand = botsfw.Command{
	Code: ASK_PHONE_NUMBER_FOR_RECEIPT_COMMAND,
	Action: func(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "AskPhoneNumberForReceiptCommand.Action()")

		input := whc.Input()

		var (
			mt             string
			phoneNumberStr string
			phoneNumber    int64
		)

		contact, isContactMessage := input.(bots.WebhookContactMessage)

		if isContactMessage {
			if contact == nil {
				m = whc.NewMessageByCode(trans.MESSAGE_TEXT_INVALID_PHONE_NUMBER)
				return m, nil
			}
			user, err := facade.User.GetUserByID(c, whc.AppUserIntID())
			if err != nil {
				return m, err
			}
			if user.FirstName == contact.FirstName() && user.LastName == contact.LastName() {
				phoneNumberStr = cleanPhoneNumber(contact.PhoneNumber())
				if phoneNumber, err = strconv.ParseInt(phoneNumberStr, 10, 64); err != nil {
					log.Warningf(c, "Failed to parse contact's phone number: [%v]", phoneNumberStr)
					err = nil
				} else if user.PhoneNumber == 0 {
					err = dtdal.DB.RunInTransaction(c, func(c context.Context) error {
						user, err := facade.User.GetUserByID(c, whc.AppUserIntID())
						if err != nil {
							return err
						}
						if user.PhoneNumber == 0 {
							user.PhoneNumber = phoneNumber
							user.PhoneNumberConfirmed = true
							return facade.User.SaveUser(c, user)
						}
						return nil
					}, nil)
					if err != nil {
						log.Errorf(c, errors.Wrap(err, "Failed to update user with phone number").Error())
						err = nil
					}
				}

				return whc.NewMessage(trans.MESSAGE_TEXT_YOU_CAN_SEND_RECEIPT_TO_YOURSELF_BY_SMS), nil
			}
			mt = contact.PhoneNumber()
		} else {
			mt = whc.Input().(bots.WebhookTextMessage).Text()
		}

		if twilioTestNumber, ok := common.TwilioTestNumbers[mt]; ok {
			log.Debugf(c, "Using predefined test number [%v]: %v", mt, twilioTestNumber)
			phoneNumberStr = twilioTestNumber
		} else {
			phoneNumberStr = cleanPhoneNumber(mt)
		}

		if phoneNumber, err = strconv.ParseInt(phoneNumberStr, 10, 64); err != nil {
			m = whc.NewMessageByCode(trans.MESSAGE_TEXT_INVALID_PHONE_NUMBER)
			return m, nil
		}

		chatEntity := whc.ChatEntity()

		awaitingUrl, err := url.Parse(chatEntity.GetAwaitingReplyTo())
		if err != nil {
			return m, errors.WithMessage(err, "Failed to parse chat state as URL")
		}

		if transferID, err := strconv.ParseInt(awaitingUrl.Query().Get(WIZARD_PARAM_TRANSFER), 10, 64); err != nil {
			return m, errors.WithMessage(err, fmt.Sprintf("Failed to parse transferID: %v", awaitingUrl))
		} else {
			transfer, err := facade.Transfers.GetTransferByID(c, transferID)
			if err != nil {
				return m, errors.WithMessage(err, "Failed to get transfer by ID")
			}
			counterparty, err := facade.GetContactByID(c, transfer.Counterparty().ContactID)
			if err != nil {
				return m, errors.WithMessage(err, "Failed to get contact by ID")
			}
			phoneContact := models.PhoneContact{PhoneNumber: phoneNumber, PhoneNumberConfirmed: false}

			return sendReceiptBySms(whc, phoneContact, transfer, counterparty)
		}
	},
}

const SMS_STATUS_MESSAGE_ID_PARAM_NAME = "SmsStatusMessageId"
const SMS_STATUS_MESSAGE_UPDATES_COUNT_PARAM_NAME = "SmsStatusUpdatesCount"

func sendReceiptBySms(whc botsfw.WebhookContext, phoneContact models.PhoneContact, transfer models.Transfer, counterparty models.Contact) (m botsfw.MessageFromBot, err error) {
	c := whc.Context()

	if transfer.TransferEntity == nil {
		if transfer, err = facade.Transfers.GetTransferByID(c, transfer.ID); err != nil {
			return m, err
		}
	}

	whc.ChatEntity() //TODO: Workaround to make whc.GetAppUser() working
	appUser, err := whc.GetAppUser()
	user := appUser.(*models.AppUserEntity)
	if err != nil {
		return
	}

	var (
		smsText   string
		receiptID int64
		//inviteCode string
	)

	receipt := models.NewReceiptEntity(whc.AppUserIntID(), transfer.ID, transfer.Counterparty().UserID, whc.Locale().Code5, "sms", strconv.FormatInt(phoneContact.PhoneNumber, 10), general.CreatedOn{
		CreatedOnPlatform: whc.BotPlatform().ID(),
		CreatedOnID:       whc.GetBotCode(),
	})
	if receiptID, err = dtdal.Receipt.CreateReceipt(c, &receipt); err != nil {
		return m, err
	}

	receiptUrl := common.GetReceiptUrl(receiptID, common.GetWebsiteHost(receipt.CreatedOnID))

	if counterparty.CounterpartyUserID == 0 {
		//related := fmt.Sprintf("%v=%v", models.TransferKind, transferID)
		//inviteKey, invite, err := invites.CreatePersonalInvite(whc, whc.AppUserIntID(), invites.InviteBySms, strconv.FormatInt(phoneContact.PhoneNumber, 10), whc.BotPlatform().ID(), whc.GetBotCode(), related)
		//if err != nil {
		//	log.Errorf(c, "Failed to create invite: %v", err)
		//	return m, err
		//}
		//inviteCode = inviteKey.StringID()
	} else {
		panic("Not implemented, need to call common.GetReceiptUrlForUser(...)")
	}

	// You've got $10 from Jack
	// You've given $10 to Jack

	switch transfer.Direction() {
	case models.TransferDirectionUser2Counterparty:
		smsText = fmt.Sprintf(whc.Translate(trans.SMS_RECEIPT_YOU_GOT), transfer.GetAmount(), user.FullName())
	case models.TransferDirectionCounterparty2User:
		smsText = fmt.Sprintf(whc.Translate(trans.SMS_RECEIPT_YOU_GAVE), transfer.GetAmount(), user.FullName())
	default:
		return m, errors.New("Unknown direction: " + string(transfer.Direction()))
	}
	smsText += "\n\n" + whc.Translate(trans.SMS_CLICK_TO_CONFIRM_OR_DECLINE, receiptUrl)

	chatEntity := whc.ChatEntity()

	var (
		smsStatusMessageID int
		//smsStatusMessageUpdatesCount int
	)

	var createSmsStatusMessage = func() error {
		var msgSmsStatus botsfw.MessageFromBot
		mt := whc.Translate(trans.MESSAGE_TEXT_SMS_QUEUING_FOR_SENDING, phoneContact.PhoneNumberAsString())
		//log.Debugf(c, "whc.InputTypes(): %v, botsfw.WebhookInputCallbackQuery: %v, MessageID: %v", whc.InputTypes(), botsfw.WebhookInputCallbackQuery, whc.InputCallbackQuery().GetMessage().IntID())
		if whc.InputType() == botsfw.WebhookInputCallbackQuery {
			//log.Debugf(c, "editMessage.MessageID: %v", editMessage.MessageID)
			if msgSmsStatus, err = whc.NewEditMessage(mt, botsfw.MessageFormatHTML); err != nil {
				return err
			}
		} else {
			msgSmsStatus = whc.NewMessage(mt)
		}
		smsStatusMsg, err := whc.Responder().SendMessage(c, msgSmsStatus, botsfw.BotAPISendMessageOverHTTPS)
		if err != nil {
			return err
		}
		smsStatusMessageID = smsStatusMsg.TelegramMessage.(tgbotapi.Message).MessageID
		chatEntity.AddWizardParam(SMS_STATUS_MESSAGE_ID_PARAM_NAME, strconv.Itoa(smsStatusMessageID))
		return nil
	}

	if err = createSmsStatusMessage(); err != nil {
		return m, err
	}
	//if smsStatusMessageID, err = strconv.Atoi(chatEntity.GetWizardParam(SMS_STATUS_MESSAGE_ID_PARAM_NAME)); err != nil {
	//	if err = createSmsStatusMessage(); err != nil {
	//		return m, err
	//	}
	//}
	//if smsStatusMessageUpdatesCount, err = strconv.Atoi(chatEntity.GetWizardParam(SMS_STATUS_MESSAGE_UPDATES_COUNT_PARAM_NAME)); err == nil {
	//	if smsStatusMessageUpdatesCount > 2 {
	//		if err = createSmsStatusMessage(); err != nil {
	//			return m, err
	//		}
	//		chatEntity.AddWizardParam(SMS_STATUS_MESSAGE_UPDATES_COUNT_PARAM_NAME, "1")
	//	} else {
	//		chatEntity.AddWizardParam(SMS_STATUS_MESSAGE_UPDATES_COUNT_PARAM_NAME, strconv.Itoa(smsStatusMessageUpdatesCount + 1))
	//	}
	//} else {
	//	chatEntity.AddWizardParam(SMS_STATUS_MESSAGE_UPDATES_COUNT_PARAM_NAME, "1")
	//}

	tgChatID, err := strconv.ParseInt(whc.MustBotChatID(), 10, 64)

	if err != nil {
		return m, errors.WithMessage(err, "Failed to parse whc.BotChatID() to int")
	}

	if lastTwilioSmsese, err := dtdal.Twilio.GetLastTwilioSmsesForUser(c, whc.AppUserIntID(), phoneContact.PhoneNumberAsString(), 1); err != nil {
		err = errors.Wrap(err, "Failed to check latest SMS records")
		return m, err
	} else if len(lastTwilioSmsese) > 0 {
		smsRecord := lastTwilioSmsese[0]
		if smsRecord.To == phoneContact.PhoneNumberAsString() && (smsRecord.Status == "delivered" || smsRecord.Status == "queued") {
			// TODO: Do smarter check for limit
			m.Text = emoji.ERROR_ICON + " " + fmt.Sprintf("Exceeded limit for sending SMS to same number: %v", phoneContact.PhoneNumberAsString())
			log.Warningf(c, m.Text)
			return m, err
		}
	}
	// TODO: Create SMS record before sending to ensure we don't spam user in case of bug after the API call.

	isTestSender, smsResponse, twilioException, err := sms.SendSms(whc.Context(), whc.GetBotSettings().Env == strongo.EnvProduction, phoneContact.PhoneNumberAsString(), smsText)
	if err != nil {
		return m, errors.WithMessage(err, "Failed to send SMS")
	}
	//sms := common.Sms{
	//	DtCreated: smsResponse.DateCreated,
	//	DtUpdate: smsResponse.DateUpdate,
	//	DtSent: smsResponse.DateSent,
	//	InviteCode: inviteCode,
	//	To: smsResponse.To,
	//	From: smsResponse.From,
	//	Status: smsResponse.Status,
	//}
	//if smsResponse.Price != nil {
	//	sms.Price = *smsResponse.Price
	//}

	if twilioException != nil {
		twilioExceptionStr, _ := json.Marshal(twilioException)
		log.Errorf(c, "Failed to send SMS via Twilio: %v", string(twilioExceptionStr))
		mt, tryAnotherNumber := sms.TwilioExceptionToMessage(whc, twilioException)
		if tryAnotherNumber {
			log.Infof(c, "Twilio identified invalid phone number, need to try another one.")
			if m, err = whc.NewEditMessage(mt, botsfw.MessageFormatText); err != nil {
				return
			}
			m.EditMessageUID = telegram.NewChatMessageUID(tgChatID, smsStatusMessageID)
			return
		}
		if counterparty.PhoneNumber == phoneContact.PhoneNumber {
			dtdal.DB.RunInTransaction(whc.Context(), func(tc context.Context) error {
				counterparty, err := facade.GetContactByID(tc, transfer.Counterparty().ContactID)
				if err != nil {
					return err
				}
				if counterparty.PhoneNumber != phoneContact.PhoneNumber {
					counterparty.PhoneNumber = phoneContact.PhoneNumber
					err = facade.SaveContact(c, counterparty)
				}
				return err
			}, nil)
		}
		if m, err = whc.NewEditMessage(fmt.Sprintf("<b>Exception</b>\n%v\n\n<b>SMS text</b>\n%v", twilioException, smsText), botsfw.MessageFormatHTML); err != nil {
			return
		}
		m.EditMessageUID = telegram.NewChatMessageUID(tgChatID, smsStatusMessageID)
		m.DisableWebPagePreview = true
		dtb_general.SetMainMenuKeyboard(whc, &m)
		return
	}

	smsResponseStr, _ := json.Marshal(smsResponse)
	log.Debugf(c, "Twilio response: %v", string(smsResponseStr))

	analytics.ReceiptSentFromBot(whc, "sms")

	if _, err = dtdal.Twilio.SaveTwilioSms(
		whc.Context(),
		smsResponse,
		transfer,
		phoneContact,
		whc.AppUserIntID(),
		tgChatID,
		smsStatusMessageID,
	); err != nil {
		return
	}

	mt := whc.Translate(trans.MESSAGE_TEXT_SMS_QUEUED_FOR_SENDING, phoneContact.PhoneNumberAsString())

	if isTestSender {
		mt += "\n\n<b>SMS text</b>\n" + smsText
	}

	if m, err = whc.NewEditMessage(mt, botsfw.MessageFormatHTML); err != nil {
		return
	}
	m.EditMessageUID = telegram.NewChatMessageUID(tgChatID, smsStatusMessageID)
	m.DisableWebPagePreview = true

	if _, err := whc.Responder().SendMessage(c, m, botsfw.BotAPISendMessageOverHTTPS); err != nil {
		err = errors.Wrap(err, "Failed to send bot response message over HTTPS")
		return m, err
	}

	return dtb_general.MainMenuCommand.Action(whc)
}
