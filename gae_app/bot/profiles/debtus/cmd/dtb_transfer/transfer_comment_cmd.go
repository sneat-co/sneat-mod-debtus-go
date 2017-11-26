package dtb_transfer

import (
	//"github.com/DebtsTracker/translations/emoji"
	//"fmt"
	"fmt"
	"net/url"
	"strings"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
	"golang.org/x/net/html"
)

const (
	//TRANSFER_WIZARD_PARAM_NOTE    = "note"
	TRANSFER_WIZARD_PARAM_COMMENT = "comment"
)

//const (
//	ADD_NOTE_COMMAND    = "add-note"
//	ADD_COMMENT_COMMAND = "add-comment"
//)
//
//func createTransferAddNoteOrCommentCommand(code string, anotherCommand *bots.Command, nextCommand bots.Command) bots.Command {
//	var icon, title string
//	switch code {
//	case ADD_NOTE_COMMAND:
//		icon = emoji.MEMO_ICON
//		title = trans.COMMAND_TEXT_ADD_NOTE_TO_TRANSFER
//	case ADD_COMMENT_COMMAND:
//		icon = emoji.NEWSPAPER_ICON
//		title = trans.COMMAND_TEXT_ADD_COMMENT_TO_TRANSFER
//	}
//
//	return bots.Command{
//		Code:  code,
//		Icon:  icon,
//		Title: title,
//		Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
//
//			log.Debugf(c, "createTransferAddNoteOrCommentCommand().Action(), code=%v", code)
//			if code != ADD_NOTE_COMMAND && code != ADD_COMMENT_COMMAND {
//				panic(fmt.Sprintf("Unknown code: %v", code))
//			}
//			chatEntity := whc.ChatEntity()
//			if chatEntity.IsAwaitingReplyTo(code) {
//				switch code {
//				case ADD_NOTE_COMMAND:
//					chatEntity.AddWizardParam(TRANSFER_WIZARD_PARAM_NOTE, whc.Input().(bots.WebhookTextMessage).Text())
//					if chatEntity.GetWizardParam(TRANSFER_WIZARD_PARAM_COMMENT) != "" {
//						return nextCommand.Action(whc)
//					}
//					m = whc.NewMessage(whc.Translate(trans.MESSAGE_TEXT_TRANSFER_NOTE_ADDED_ASK_FOR_COMMENT))
//					m.Keyboard = tgbotapi.NewReplyKeyboard(
//						[]tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton(anotherCommand.DefaultTitle(whc))},
//						[]tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton(whc.Translate(trans.COMMAND_TEXT_NO_COMMENT_FOR_TRANSFER))},
//					)
//				case ADD_COMMENT_COMMAND:
//					chatEntity.AddWizardParam(TRANSFER_WIZARD_PARAM_COMMENT, whc.Input().(bots.WebhookTextMessage).Text())
//					if chatEntity.GetWizardParam(TRANSFER_WIZARD_PARAM_NOTE) != "" {
//						return nextCommand.Action(whc)
//					}
//					m = whc.NewMessage(whc.Translate(trans.MESSAGE_TEXT_TRANSFER_COMMENT_ADDED_ASK_FOR_NOTE))
//					m.Keyboard = tgbotapi.NewReplyKeyboard(
//						[]tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton(anotherCommand.DefaultTitle(whc))},
//						[]tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton(whc.Translate(trans.COMMAND_TEXT_NO_NOTE_FOR_TRANSFER))},
//					)
//				default:
//					panic(fmt.Sprintf("Unknown code: %v", code))
//				}
//				chatEntity.PopStepsFromAwaitingReplyUpToSpecificParent(ASK_NOTE_OR_COMMENT_FOR_TRANSFER_COMMAND)
//			} else {
//				chatEntity.PushStepToAwaitingReplyTo(code)
//				switch code {
//				case ADD_NOTE_COMMAND:
//					m = whc.NewMessage(whc.Translate(trans.MESSAGE_TEXT_TRANSFER_ASK_FOR_NOTE))
//				case ADD_COMMENT_COMMAND:
//					m = whc.NewMessage(whc.Translate(trans.MESSAGE_TEXT_TRANSFER_ASK_FOR_COMMENT))
//				default:
//					panic(fmt.Sprintf("Unknown code: %v", code))
//				}
//				m.Keyboard = tgbotapi.NewHideKeyboard(true)
//			}
//			m.Format = bots.MessageFormatHTML
//			return m, err
//		},
//	}
//}

func createTransferAskNoteOrCommentCommand(code string, nextCommand bots.Command) bots.Command {
	var addNoteCommand bots.Command
	var addCommentCommand bots.Command

	//addNoteCommand = createTransferAddNoteOrCommentCommand(ADD_NOTE_COMMAND, &addCommentCommand, nextCommand)
	//addCommentCommand = createTransferAddNoteOrCommentCommand(ADD_COMMENT_COMMAND, &addNoteCommand, nextCommand)

	return bots.Command{
		Code: code,
		Replies: []bots.Command{
			addNoteCommand,
			addCommentCommand,
		},
		Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
			c := whc.Context()
			log.Infof(c, "createTransferAskNoteOrCommentCommand().Action()")
			chatEntity := whc.ChatEntity()
			//noOptionSelected := false
			if chatEntity.IsAwaitingReplyTo(code) {
				if m, err = interestAction(whc, nextCommand.Action); err != nil || m.Text != "" {
					return
				}
				mt := whc.Input().(bots.WebhookTextMessage).Text()
				switch mt {
				//case whc.Translate(trans.COMMAND_TEXT_ADD_NOTE_TO_TRANSFER):
				//	return addNoteCommand.Action(whc)
				//case whc.Translate(trans.COMMAND_TEXT_ADD_COMMENT_TO_TRANSFER):
				//	return addCommentCommand.Action(whc)
				//case whc.Translate(trans.COMMAND_TEXT_NO_COMMENT_OR_NOTE_FOR_TRANSFER):
				//	return nextCommand.Action(whc)
				case whc.Translate(trans.COMMAND_TEXT_NO_COMMENT_FOR_TRANSFER):
					return nextCommand.Action(whc)
					//case whc.Translate(trans.COMMAND_TEXT_NO_NOTE_FOR_TRANSFER):
					//	return nextCommand.Action(whc)
				default:
					chatEntity.AddWizardParam(TRANSFER_WIZARD_PARAM_COMMENT, mt)
					return nextCommand.Action(whc)
					//noOptionSelected = true
				}
			} else {
				chatEntity.PushStepToAwaitingReplyTo(code)
			}

			m = whc.NewMessage(whc.Translate(trans.MESSAGE_TEXT_TRANSFER_ASK_FOR_INTEREST_SHORT))
			m.Format = bots.MessageFormatHTML
			m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
				[]tgbotapi.InlineKeyboardButton{
					tgbotapi.NewInlineKeyboardButtonData(whc.Translate(trans.COMMAND_TEXT_MORE_ABOUT_INTEREST_COMMAND), ASK_FOR_INTEREST_AND_COMMENT_COMMAND),
				},
			)
			if _, err = whc.Responder().SendMessage(c, m, bots.BotApiSendMessageOverHTTPS); err != nil {
				return
			}

			var transferWizard TransferWizard
			if transferWizard, err = NewTransferWizard(whc); err != nil {
				return
			}
			counterpartyID := transferWizard.CounterpartyID(c)
			if counterpartyID == 0 {
				return m, errors.New("transferWizard.CounterpartyID() == 0")
			}
			counterparty, err := dal.Contact.GetContactByID(whc.Context(), counterpartyID)
			m.Text = strings.TrimLeft(fmt.Sprintf("%v\n(<i>%v</i>)",
				whc.Translate(trans.MESSAGE_TEXT_TRANSFER_ASK_FOR_COMMENT_ONLY),
				whc.Translate(trans.MESSAGE_TEXT_VISIBLE_TO_YOU_AND_COUNTERPARTY, html.EscapeString(counterparty.FullName()))),
				"\n ",
			)

			m.Keyboard = tgbotapi.NewReplyKeyboard([]tgbotapi.KeyboardButton{{Text: whc.Translate(trans.COMMAND_TEXT_NO_COMMENT_FOR_TRANSFER)}})
			m.Format = bots.MessageFormatHTML
			return
		},
	}
}

const ASK_FOR_INTEREST_AND_COMMENT_COMMAND = "ask-for-interest-and-comment-long"

var AskForInterestAndCommentCallbackCommand = bots.Command{
	Code: ASK_FOR_INTEREST_AND_COMMENT_COMMAND,
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		m.Text = whc.Translate(trans.MESSAGE_TEXT_TRANSFER_ASK_FOR_INTEREST_LONG)
		m.Format = bots.MessageFormatHTML
		m.IsEdit = true
		return
	},
}

const ASK_NOTE_OR_COMMENT_FOR_TRANSFER_COMMAND = "ask-note-or-comment"

var TransferFromUserAskNoteOrCommentCommand = createTransferAskNoteOrCommentCommand(
	ASK_NOTE_OR_COMMENT_FOR_TRANSFER_COMMAND,
	BorrowingWizardCompletedCommand,
)

var TransferToUserAskNoteOrCommentCommand = createTransferAskNoteOrCommentCommand(
	ASK_NOTE_OR_COMMENT_FOR_TRANSFER_COMMAND,
	LendingWizardCompletedCommand,
)
