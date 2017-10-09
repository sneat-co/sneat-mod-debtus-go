package bot_shared

import (
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-api-telegram"
	"fmt"
	"bytes"
	"github.com/DebtsTracker/translations/trans"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
	"net/url"
	"github.com/strongo/app"
	"bitbucket.com/debtstracker/gae_app/debtstracker/common"
	"golang.org/x/net/context"
	"github.com/strongo/app/log"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/strongo/decimal"
	"bitbucket.com/debtstracker/gae_app/debtstracker/dal"
	"github.com/pkg/errors"
)

const BILL_CARD_COMMAND = "bill"

func startBillAction(whc bots.WebhookContext, billParam string, botParams BotParams) (m bots.MessageFromBot, err error) {
	var bill models.Bill
	if bill.ID = billParam[len("bill-"):]; bill.ID == "" {
		return m, errors.New("Invalid bill parameter")
	}
	if bill, err = dal.Bill.GetBillByID(whc.Context(), bill.ID); err != nil {
		return
	}
	return ShowBillCard(whc, botParams, false, bill, "")
}

func billCardCommand(botParams BotParams) bots.Command {
	return BillCallbackCommand(BILL_CARD_COMMAND,
		func(whc bots.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m bots.MessageFromBot, err error) {
			c := whc.Context()
			log.Debugf(c, "billCardCommand.CallbackAction()")
			return ShowBillCard(whc, botParams, false, bill, "")
		},
	)
}

func BillCardCallbackCommandData(billID string) string {
	return BillCallbackCommandData(BILL_CARD_COMMAND, billID)
}

const BILL_MEMBERS_COMMAND = "bill-members"

func BillCallbackCommandData(command string, billID string) string {
	return command + "?bill=" + billID
}

var billMembersCommand = BillCallbackCommand(BILL_MEMBERS_COMMAND,
	func(whc bots.WebhookContext, callbackURL *url.URL, bill models.Bill) (m bots.MessageFromBot, err error) {
		var buffer bytes.Buffer
		if err = WriteBillCardTitle(whc.Context(), bill, whc.GetBotCode(), &buffer, whc); err != nil {
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
						CallbackData: BillCallbackCommandData(JOIN_BILL_COMMAND, bill.ID),
					},
				},
				{
					{
						Text:         whc.Translate(trans.COMMAND_TEXT_INVITE_MEMBER),
						CallbackData: BillCallbackCommandData(INVITE_BILL_MEMBER_COMMAND, bill.ID),
					},
				},
				{
					{
						Text:         whc.Translate(emoji.RETURN_BACK_ICON),
						CallbackData: BillCardCallbackCommandData(bill.ID),
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
	billCurrency := models.Currency(bill.Currency)
	type MemberRowParams struct {
		N          int
		MemberName string
		Percent    decimal.Decimal64p2
		Owes       models.Amount
		Paid       models.Amount
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
			Owes:       models.NewAmount(billCurrency, member.Owes),
			Paid:       models.NewAmount(billCurrency, member.Paid),
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
		buffer.WriteString("\n")
	}
}

const INVITE_BILL_MEMBER_COMMAND = "invite2bill"

const INLINE_COMMAND_JOIN = "join"

var inviteToBillCommand = BillCallbackCommand(INVITE_BILL_MEMBER_COMMAND,
	func(whc bots.WebhookContext, callbackURL *url.URL, bill models.Bill) (m bots.MessageFromBot, err error) {
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
						CallbackData: BillCardCallbackCommandData(bill.ID),
					},
				},
			},
		}
		return
	},
)

func ShowBillCard(whc bots.WebhookContext, botParams BotParams, isEdit bool, bill models.Bill, footer string) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	m = whc.NewMessage("")
	m.IsEdit = isEdit
	if m.Text, err = GetBillCardMessageText(c, whc.GetBotCode(), whc, bill, true, footer); err != nil {
		return
	}
	if whc.IsInGroup() || whc.Chat() == nil {
		m.Keyboard = botParams.GetGroupBillCardInlineKeyboard(whc, bill)
	} else {
		m.Keyboard = botParams.GetPrivateBillCardInlineKeyboard(whc, whc.GetBotCode(), bill);
	}
	return
}

func WriteBillCardTitle(c context.Context, bill models.Bill, botID string, buffer *bytes.Buffer, translator strongo.SingleLocaleTranslator) error {
	var amount interface{}
	if bill.Currency == "" {
		amount = bill.AmountTotal
	} else {
		amount = bill.TotalAmount()
	}
	titleWithLink := fmt.Sprintf(`<a href="https://t.me/%v?start=bill-%d">%v</a>`, botID, bill.ID, bill.Name)
	log.Debugf(c, "titleWithLink: %v", titleWithLink)
	header := translator.Translate(trans.MESSAGE_TEXT_BILL_CARD_HEADER, amount, titleWithLink)
	log.Debugf(c, "header: %v", header)
	if _, err := buffer.WriteString(header); err != nil {
		log.Errorf(c, "Failed to write bill header")
		return err
	}
	return nil
}

func GetBillCardMessageText(c context.Context, botID string, translator strongo.SingleLocaleTranslator, bill models.Bill, showMembers bool, footer string) (string, error) {
	log.Debugf(c, "GetBillCardMessageText() => bill.BillEntity: %v", bill.BillEntity)

	var buffer bytes.Buffer
	log.Debugf(c, "Will write bill header...")

	if err := WriteBillCardTitle(c, bill, botID, &buffer, translator); err != nil {
		return "", err
	}

	log.Debugf(c, "GetBillCardMessageText() => showGroupMembers=%v", showMembers)

	if showMembers {
		//buffer.WriteString("\n")
		//buffer.WriteString(translator.Translate(trans.MESSAGE_TEXT_SPLIT_LABEL_WITH_VALUE, translator.Translate(string(bill.SplitMode))))
		//if bill.Status != models.BillStatusActive {
		//	buffer.WriteString(", " + translator.Translate(trans.MESSAGE_TEXT_STATUS, bill.Status))
		//}
		//buffer.WriteString(fmt.Sprintf("\n\n<b>%v</b> (%d)\n\n", translator.Translate(trans.MESSAGE_TEXT_MEMBERS_TITLE), bill.MembersCount))
		buffer.WriteString("\n\n")
		writeBillMembersList(c, &buffer, translator, bill, "")
	}

	if footer != "" {
		buffer.WriteString("\n\n")
		buffer.WriteString(footer)
	}
	log.Debugf(c, "GetBillCardMessageText() completed")
	return buffer.String(), nil
}
