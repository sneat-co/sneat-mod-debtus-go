package dtb_transfer

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
)

const BALANCE_COMMAND = "balance"

var BalanceCallbackCommand = bots.NewCallbackCommand(BALANCE_COMMAND, balanceCallbackAction)

var BalanceCommand = bots.Command{ //TODO: Write unit tests!
	Code:     BALANCE_COMMAND,
	Title:    trans.COMMAND_TEXT_BALANCE,
	Icon:     emoji.BALANCE_ICON,
	Commands: trans.Commands(trans.COMMAND_BALANCE),
	Action:   balanceAction,
}

func balanceCallbackAction(whc bots.WebhookContext, _ *url.URL) (m bots.MessageFromBot, err error) {
	return balanceAction(whc)
}

func balanceAction(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
	c := whc.Context()

	log.Debugf(c, "BalanceCommand.Action()")

	var user models.AppUser

	if user, err = dal.User.GetUserByID(c, whc.AppUserIntID()); err != nil {
		return
	}

	var buffer bytes.Buffer
	if user.BalanceCount == 0 {
		if _, err = buffer.WriteString(whc.Translate(trans.MESSAGE_TEXT_BALANCE_IS_ZERO)); err != nil {
			return
		}
	} else {
		balanceMessageBuilder := NewBalanceMessageBuilder(whc)
		contacts := user.Contacts()
		if len(contacts) == 0 {
			return m, fmt.Errorf("Integrity issue: User{ID=%v} has non zero balance and no contacts.", whc.AppUserIntID())
		}
		buffer.WriteString(fmt.Sprintf("<b>%v</b>", whc.Translate(trans.MESSAGE_TEXT_BALANCE_HEADER)) + common.HORIZONTAL_LINE)
		linker := common.NewLinkerFromWhc(whc)
		buffer.WriteString(balanceMessageBuilder.ByCounterparty(c, linker, contacts))

		var thereAreFewDebtsForSingleCurrency = func() bool {
			//TODO: Duplicate call to Balance() - consider move inside BalanceMessageBuilder
			//log.Debugf(c, "thereAreFewDebtsForSingleCurrency()")
			var currencies []models.Currency
			for _, counterparty := range contacts {
				//log.Debugf(c, "counterparty: %v", counterparty)
				for currency, _ := range counterparty.Balance() {
					//log.Debugf(c, "currency: %v", currency)
					for _, curr := range currencies {
						//log.Debugf(c, "curr: %v; curr == currency: %v", curr, curr == currency)
						if curr == currency {
							return true
						}
					}
					currencies = append(currencies, currency)
				}
			}
			//log.Debugf(c, "thereAreFewDebtsForSingleCurrency: %v", currencies)
			return false
		}

		if len(contacts) > 1 && thereAreFewDebtsForSingleCurrency() {
			userBalanceWithInterest := user.BalanceWithInterest(c, time.Now())
			buffer.WriteString("\n" + strings.Repeat("─", 16) + "\n" + balanceMessageBuilder.ByCurrency(true, userBalanceWithInterest))
		}

		//if len(contacts) > 0 {
		//	//for i, counterparty := range contacts {
		//	//	telegramKeyboard = append(telegramKeyboard, []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(counterparty.FullName(), fmt.Sprintf("transfer-history?counterparty=%v", counterpartyKeys[i].IntID()))})
		//	//}
		//	telegramKeyboard = append(telegramKeyboard, []tgbotapi.InlineKeyboardButton{
		//		tgbotapi.NewInlineKeyboardButtonData("<", fmt.Sprintf("balance?counterparty=%v", counterpartyKeys[len(counterpartyKeys)-1].IntID())),
		//		tgbotapi.NewInlineKeyboardButtonData(">", fmt.Sprintf("balance?counterparty=%v", counterpartyKeys[0].IntID())),
		//	})
		//}
	}
	buffer.WriteString(common.HORIZONTAL_LINE)
	//buffer.WriteString(dtb_general.AdSlot(whc, "balance"))
	const THUMB_UP = "👍"
	buffer.WriteString(THUMB_UP + " " + whc.Translate(trans.MESSAGE_TEXT_PLEASE_HELP_MAKE_IT_BETTER))
	if whc.InputType() == bots.WebhookInputCallbackQuery {
		if m, err = whc.NewEditMessage(buffer.String(), bots.MessageFormatHTML); err != nil {
			return
		}
	} else {
		m = whc.NewMessage(buffer.String())
		m.Format = bots.MessageFormatHTML
	}

	m.DisableWebPagePreview = true

	if user.HasDueTransfers {
		m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
			[]tgbotapi.InlineKeyboardButton{
				{
					Text:         whc.Translate(trans.COMMAND_TEXT_DUE_RETURNS),
					CallbackData: DUE_RETURNS_COMMAND,
				},
			},
			[]tgbotapi.InlineKeyboardButton{
				{
					Text:         whc.Translate(trans.COMMAND_TEXT_INVITE_FIREND),
					CallbackData: "invite",
				},
			},
		)
	}

	//err = whc.Responder().SendMessage(c, m, bots.BotApiSendMessageOverHTTPS)
	return m, err
	//SetMainMenuKeyboard(whc, &m) - Bad idea! Need to cleanup AwaitingReplyTo
}
