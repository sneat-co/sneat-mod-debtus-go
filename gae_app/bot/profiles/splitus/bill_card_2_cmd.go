package splitus

import (
	"bytes"
	"fmt"
	"github.com/crediterra/money"
	"net/url"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/app"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/decimal"
	"github.com/strongo/log"
)

const billCardCommandCode = "bill-card"

var billCardCommand = bots.Command{
	Code: billCardCommandCode,
	CallbackAction: billCallbackAction(func(whc bots.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		if m.Text, err = getBillCardMessageText(c, whc.GetBotCode(), whc, bill, false, ""); err != nil {
			return
		}
		m.Format = bots.MessageFormatHTML
		m.Keyboard = getGroupBillCardInlineKeyboard(whc, bill)
		return
	}),
}

func startBillAction(whc bots.WebhookContext, billParam string) (m bots.MessageFromBot, err error) {
	var bill models.Bill
	if bill.ID = billParam[len("bill-"):]; bill.ID == "" {
		return m, errors.New("Invalid bill parameter")
	}
	if bill, err = facade.GetBillByID(whc.Context(), bill.ID); err != nil {
		return
	}
	return ShowBillCard(whc, false, bill, "")
}

func billCardCallbackCommandData(billID string) string {
	return billCallbackCommandData(billCardCommandCode, billID)
}

const billMembersCommandCode = "bill-members"

func billCallbackCommandData(command string, billID string) string {
	return command + "?bill=" + billID
}

var billMembersCommand = billCallbackCommand(billMembersCommandCode, nil,
	func(whc bots.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m bots.MessageFromBot, err error) {
		var buffer bytes.Buffer
		if err = writeBillCardTitle(whc.Context(), bill, whc.GetBotCode(), &buffer, whc); err != nil {
			return
		}
		buffer.WriteString("\n\n")
		writeBillMembersList(whc.Context(), &buffer, whc, bill, "")
		m.Text = buffer.String()
		m.Format = bots.MessageFormatHTML

		m.Keyboard = &tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{
					{
						Text:         whc.Translate(trans.BUTTON_TEXT_JOIN),
						CallbackData: billCallbackCommandData(joinBillCommandCode, bill.ID),
					},
				},
				{
					{
						Text:         whc.Translate(trans.COMMAND_TEXT_INVITE_MEMBER),
						CallbackData: billCallbackCommandData(INVITE_BILL_MEMBER_COMMAND, bill.ID),
					},
				},
				{
					{
						Text:         whc.Translate(emoji.RETURN_BACK_ICON),
						CallbackData: billCardCallbackCommandData(bill.ID),
					},
				},
			},
		}
		return
	},
)

func writeBillMembersList(
	c context.Context,
	buffer *bytes.Buffer,
	translator strongo.SingleLocaleTranslator,
	bill models.Bill,
	selectedMemberID string,
) {
	billCurrency := money.Currency(bill.Currency)
	type MemberRowParams struct {
		N          int
		MemberName string
		Percent    decimal.Decimal64p2
		Owes       money.Amount
		Paid       money.Amount
	}
	billMembers := bill.GetBillMembers()

	totalShares := 0

	for _, member := range billMembers {
		totalShares += member.Shares
	}

	for i, member := range bill.GetBillMembers() {
		templateParams := MemberRowParams{
			N:          i + 1,
			MemberName: member.Name,
			Owes:       money.NewAmount(billCurrency, member.Owes),
			Paid:       money.NewAmount(billCurrency, member.Paid),
		}
		if totalShares == 0 {
			templateParams.Percent = decimal.Decimal64p2(1 * 100 / len(billMembers))
		} else {
			templateParams.Percent = decimal.Decimal64p2(member.Shares * 100 * 100 / totalShares)
		}

		var (
			templateName string
			err          error
		)
		if member.Paid == bill.AmountTotal {
			buffer.WriteString("<b>")
		}
		if err = common.HtmlTemplates.RenderTemplate(c, buffer, translator, trans.MESSAGE_TEXT_BILL_CARD_MEMBER_TITLE, templateParams); err != nil {
			log.Errorf(c, "Failed to render template")
			return
		}
		if member.Paid == bill.AmountTotal {
			buffer.WriteString("</b>")
		}

		if selectedMemberID == "" {
			switch {
			case member.Owes > 0 && member.Paid > 0:
				templateName = trans.MESSAGE_TEXT_BILL_CARD_MEMBERS_ROW_PART_PAID
			case member.Owes > 0:
				templateName = trans.MESSAGE_TEXT_BILL_CARD_MEMBERS_ROW_OWES
			case member.Paid > 0:
				templateName = trans.MESSAGE_TEXT_BILL_CARD_MEMBERS_ROW_PAID
			default:
				templateName = trans.MESSAGE_TEXT_BILL_CARD_MEMBERS_ROW
			}
		} else {
			templateName = trans.MESSAGE_TEXT_BILL_CARD_MEMBERS_ROW
		}

		log.Debugf(c, "Will render template")
		buffer.WriteString(" ")
		if err = common.HtmlTemplates.RenderTemplate(c, buffer, translator, templateName, templateParams); err != nil {
			log.Errorf(c, "Failed to render template")
			return
		}
		buffer.WriteString("\n\n")
	}
}

const INVITE_BILL_MEMBER_COMMAND = "invite2bill"

const INLINE_COMMAND_JOIN = "join"

var inviteToBillCommand = billCallbackCommand(INVITE_BILL_MEMBER_COMMAND, nil,
	func(whc bots.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m bots.MessageFromBot, err error) {
		m.Keyboard = &tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{
					tgbotapi.NewInlineKeyboardButtonSwitchInlineQuery(
						"via Telegram",
						INLINE_COMMAND_JOIN+"?bill="+bill.ID,
					),
				},
				{
					{
						Text:         whc.Translate(emoji.RETURN_BACK_ICON),
						CallbackData: billCardCallbackCommandData(bill.ID),
					},
				},
			},
		}
		return
	},
)

func ShowBillCard(whc bots.WebhookContext, isEdit bool, bill models.Bill, footer string) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	m = whc.NewMessage("")
	m.IsEdit = isEdit
	if m.Text, err = getBillCardMessageText(c, whc.GetBotCode(), whc, bill, true, footer); err != nil {
		return
	}
	if whc.IsInGroup() || whc.Chat() == nil {
		m.Keyboard = getGroupBillCardInlineKeyboard(whc, bill)
	} else {
		m.Keyboard = getPrivateBillCardInlineKeyboard(whc, whc.GetBotCode(), bill)
	}
	return
}

func writeBillCardTitle(c context.Context, bill models.Bill, botID string, buffer *bytes.Buffer, translator strongo.SingleLocaleTranslator) error {
	var amount interface{}
	if bill.Currency == "" {
		amount = bill.AmountTotal
	} else {
		amount = bill.TotalAmount()
	}
	titleWithLink := fmt.Sprintf(`<a href="https://t.me/%v?start=bill-%v">%v</a>`, botID, bill.ID, bill.Name)
	log.Debugf(c, "titleWithLink: %v", titleWithLink)
	header := translator.Translate(trans.MESSAGE_TEXT_BILL_CARD_HEADER, amount, titleWithLink)
	log.Debugf(c, "header: %v", header)
	if _, err := buffer.WriteString(header); err != nil {
		log.Errorf(c, "Failed to write bill header")
		return err
	}
	return nil
}

func getBillCardMessageText(c context.Context, botID string, translator strongo.SingleLocaleTranslator, bill models.Bill, showMembers bool, footer string) (string, error) {
	log.Debugf(c, "getBillCardMessageText() => bill.BillEntity: %v", bill.BillEntity)

	var buffer bytes.Buffer
	log.Debugf(c, "Will write bill header...")

	if err := writeBillCardTitle(c, bill, botID, &buffer, translator); err != nil {
		return "", err
	}
	//buffer.WriteString("\n" + strings.Repeat("―", 15))

	buffer.WriteString("\n" + translator.Translate(trans.MT_TEXT_MEMBERS_COUNT, bill.MembersCount))

	if showMembers {
		//buffer.WriteString("\n")
		//buffer.WriteString(translator.Translate(trans.MESSAGE_TEXT_SPLIT_LABEL_WITH_VALUE, translator.Translate(string(bill.SplitMode))))
		//if bill.Status != models.BillStatusOutstanding {
		//	buffer.WriteString(", " + translator.Translate(trans.MESSAGE_TEXT_STATUS, bill.Status))
		//}
		//buffer.WriteString(fmt.Sprintf("\n\n<b>%v</b> (%d)\n\n", translator.Translate(trans.MESSAGE_TEXT_MEMBERS_TITLE), bill.MembersCount))
		buffer.WriteString("\n\n")
		writeBillMembersList(c, &buffer, translator, bill, "")
	}

	if footer != "" {
		if !showMembers || bill.MembersCount == 0 {
			buffer.WriteString("\n\n")
		}
		buffer.WriteString(footer)
	}
	log.Debugf(c, "getBillCardMessageText() completed")
	return buffer.String(), nil
}
