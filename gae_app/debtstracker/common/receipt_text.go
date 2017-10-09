package common

import (
	"github.com/DebtsTracker/translations/emoji"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
	"bytes"
	"fmt"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app"
	"github.com/strongo/app/log"
	"github.com/strongo/bots-framework/platforms/telegram"
	"html"
	"html/template"
)

type ShowReceiptTo int8

const (
	ShowReceiptToAutodetect ShowReceiptTo = iota
	ShowReceiptToCreator
	ShowReceiptToCounterparty
)

type ReceiptPartyAction int8

const (
	ReceiptPartyGive = iota
	ReceiptPartyGot
)

type receiptTextBuilder struct {
	//whc bots.WebhookContext
	transfer      models.Transfer
	showReceiptTo ShowReceiptTo
	viewerUserID  int64
	partyAction   ReceiptPartyAction
	//
	strongo.ExecutionContext
	//
	//showAds        bool
}

func newReceiptTextBuilder(ec strongo.ExecutionContext, transfer models.Transfer, showReceiptTo ShowReceiptTo) receiptTextBuilder {
	if transfer.ID == 0 {
		panic("transferID == 0")
	}
	r := receiptTextBuilder{
		transfer:         transfer,
		showReceiptTo:    showReceiptTo,
		ExecutionContext: ec,
	}
	switch showReceiptTo {
	case ShowReceiptToCreator:
		r.viewerUserID = transfer.CreatorUserID
	case ShowReceiptToCounterparty:
		r.viewerUserID = transfer.Counterparty().UserID
	default:
		panic(fmt.Sprintf("Unknown showReceiptTo: %v", showReceiptTo))
	}
	if (showReceiptTo == ShowReceiptToCreator && r.transfer.Direction() == models.TransferDirectionCounterparty2User) ||
		(showReceiptTo == ShowReceiptToCounterparty && r.transfer.Direction() == models.TransferDirectionUser2Counterparty) {
		r.partyAction = ReceiptPartyGot
	} else if (showReceiptTo == ShowReceiptToCounterparty && r.transfer.Direction() == models.TransferDirectionCounterparty2User) ||
		(showReceiptTo == ShowReceiptToCreator && r.transfer.Direction() == models.TransferDirectionUser2Counterparty) {
		r.partyAction = ReceiptPartyGive
	} else {
		if showReceiptTo != ShowReceiptToCreator && showReceiptTo != ShowReceiptToCounterparty {
			panic(fmt.Sprintf("Unknown ShowReceiptTo: %v", r.showReceiptTo))
		}
		panic(fmt.Sprintf("Invalid direction (%v) or showReceiptTo (%v)", r.transfer.Direction(), showReceiptTo))
	}
	return r
}

func (r receiptTextBuilder) validateRequiredParams() {
}

func (r receiptTextBuilder) receiptCommonFooter(buffer *bytes.Buffer) {
	transfer := r.transfer
	if r.showReceiptTo == ShowReceiptToCreator && transfer.Creator().Note != "" {
		buffer.WriteString("\n" + fmt.Sprintf(emoji.MEMO_ICON+" <b>%v</b>: %v", r.Translate(trans.MESSAGE_TEXT_NOTE), html.EscapeString(transfer.Creator().Note)))
	}
	if r.showReceiptTo == ShowReceiptToCounterparty && transfer.Counterparty().Note != "" {
		buffer.WriteString("\n" + fmt.Sprintf(emoji.MEMO_ICON+" <b>%v</b>: %v", r.Translate(trans.MESSAGE_TEXT_NOTE), html.EscapeString(transfer.Counterparty().Note)))
	}

	if transfer.Creator().Comment != "" {
		label := r.Translate(trans.MESSAGE_TEXT_COMMENT)
		buffer.WriteString("\n" + fmt.Sprintf(emoji.NEWSPAPER_ICON+" <b>%v</b>: %v", label, html.EscapeString(transfer.Creator().Comment)))
	}
	if transfer.Counterparty().Comment != "" {
		label := r.Translate(trans.MESSAGE_TEXT_COMMENT)
		buffer.WriteString("\n" + fmt.Sprintf(emoji.NEWSPAPER_ICON+" <b>%v</b>: %v", label, html.EscapeString(transfer.Counterparty().Comment)))
	}

	//if r.counterpartyID > 0 {
	//	if transfer.CreatorNote != "" || transfer.CreatorComment != "" {
	//		buffer.WriteString(common.HORIZONTAL_LINE)
	//	} else {
	//		buffer.WriteString("\n\n")
	//	}
	//
	//	counterpartyBalance, _ := counterparty.Balance()
	//	utmParams := NewUtmParams(whc, UTM_CAMPAIGN_RECEIPT)
	//	if len(counterpartyBalance) == 0 {
	//		counterpartyLink := GetCounterpartyLink(whc.AppUserIntID(), whc.Locale(), counterparty.Info(counterpartyID, "", ""), utmParams)
	//		switch transfer.Direction {
	//		case TransferDirectionCounterparty2User:
	//			buffer.WriteString(whc.Translate(trans.MESSAGE_TEXT_ON_RETURN_COUNTERPARTY_DOES_NOT_OWE_ANYTHING_TO_USER_ANYMORE, counterpartyLink))
	//		case TransferDirectionUser2Counterparty:
	//			buffer.WriteString(whc.Translate(trans.MESSAGE_TEXT_ON_RETURN_USER_DOES_NOT_OWE_ANYTHING_TO_COUNTERPARTY_ANYMORE, counterpartyLink))
	//		default:
	//			panic(fmt.Sprintf("Transfer %v has unkown direction: [%v]", tm.transferID, transfer.Direction))
	//		}
	//	} else {
	//		r.addBalance(whc, buffer, counterpartyID, counterparty, counterpartyBalance, utmParams)
	//	}
	//}

	//if r.showAds {
	//	if !strings.HasSuffix(buffer.String(), common.HORIZONTAL_LINE) {
	//		buffer.WriteString(common.HORIZONTAL_LINE)
	//	}
	//	buffer.WriteString(dtb_general.AdSlot(r.whc, UTM_CAMPAIGN_RECEIPT))
	//}
}

func TextReceiptForTransfer(ec strongo.ExecutionContext, transfer models.Transfer, showToUserID int64, showReceiptTo ShowReceiptTo, utmParams UtmParams) string {
	if transfer.ID == 0 {
		panic("transferID == 0")
	}
	if transfer.TransferEntity == nil {
		panic("transferID == 0")
	}
	c := ec.Context()
	log.Debugf(c, "TextReceiptForTransfer(transferID=%v)", transfer.ID)

	//transferEntity := transfer.TransferEntity

	switch showReceiptTo {
	case ShowReceiptToCreator:
		if showToUserID != 0 && showToUserID != transfer.CreatorUserID {
			panic("showToUserID != 0 && showToUserID != transferEntity.CreatorUserID")
		}
	case ShowReceiptToCounterparty:
		if showToUserID != 0 && transfer.Counterparty().UserID != 0 && showToUserID != transfer.Counterparty().UserID {
			panic("showToUserID != 0 && showToUserID != transferEntity.Counterparty().UserID")
		}
	case ShowReceiptToAutodetect:
		switch showToUserID {
		case transfer.CreatorUserID:
			showReceiptTo = ShowReceiptToCreator
		case transfer.Counterparty().UserID:
			showReceiptTo = ShowReceiptToCounterparty
		default:
			if transfer.Counterparty().UserID == 0 {
				showReceiptTo = ShowReceiptToCounterparty
			} else {
				panic(fmt.Sprintf("Parameter showToUserID=%v is not related to transferEntity with id=%v", showToUserID, transfer.ID))
			}
		}
	}

	r := newReceiptTextBuilder(ec, transfer, showReceiptTo)

	var buffer bytes.Buffer
	r.WriteReceiptText(&buffer, utmParams)
	r.receiptCommonFooter(&buffer)
	return buffer.String()
}

func (r receiptTextBuilder) getReceiptCounterparty() *models.TransferCounterpartyInfo {
	switch r.showReceiptTo {
	case ShowReceiptToCreator:
		return r.transfer.Counterparty()
	case ShowReceiptToCounterparty:
		return r.transfer.Creator()
	default:
		panic(fmt.Sprintf("Unknown ShowReceiptTo: %v", r.showReceiptTo))
	}
}

func (r receiptTextBuilder) receiptOnReturn(utmParams UtmParams) string {
	var messageTextToTranslate string
	return r.translateAndFormatMessage(messageTextToTranslate, r.transfer.GetAmount(), utmParams)
}

func (r receiptTextBuilder) WriteReceiptText(buffer *bytes.Buffer, utmParams UtmParams) {
	var messageTextToTranslate string
	if r.transfer.IsReturn {
		switch r.partyAction {
		case ReceiptPartyGive:
			messageTextToTranslate = trans.MESSAGE_TEXT_RECEIPT_RETURN_FROM_USER
		case ReceiptPartyGot:
			messageTextToTranslate = trans.MESSAGE_TEXT_RECEIPT_RETURN_TO_USER
		default:
			panic(fmt.Sprintf("Unknown partyAction: %v", r.partyAction))
		}
	} else {
		switch r.partyAction {
		case ReceiptPartyGive:
			messageTextToTranslate = trans.MESSAGE_TEXT_RECEIPT_NEW_DEBT_FROM_USER
		case ReceiptPartyGot:
			messageTextToTranslate = trans.MESSAGE_TEXT_RECEIPT_NEW_DEBT_TO_USER
		default:
			panic(fmt.Sprintf("Unknown partyAction: %v", r.partyAction))
		}
	}

	buffer.WriteString(r.translateAndFormatMessage(messageTextToTranslate, r.transfer.GetAmount(), utmParams))
	if !r.transfer.DtDueOn.IsZero() {
		buffer.WriteString("\n" + emoji.ALARM_CLOCK_ICON + " " + fmt.Sprintf(r.Translate(trans.MESSAGE_TEXT_DUE_ON), r.transfer.DtDueOn.Format("2006-01-02 15:04")))
	}

	if r.transfer.AmountInCentsReturned > 0 && r.transfer.AmountInCentsReturned != r.transfer.AmountInCents {
		buffer.WriteString("\n" + r.translateAndFormatMessage(trans.MESSAGE_TEXT_RECEIPT_ALREADY_RETURNED_AMOUNT, r.transfer.GetReturnedAmount(), utmParams))
	}

	if r.transfer.AmountInCentsOutstanding > 0 && r.transfer.AmountInCentsOutstanding != r.transfer.AmountInCents {
		buffer.WriteString("\n" + r.translateAndFormatMessage(trans.MESSAGE_TEXT_RECEIPT_OUTSTANDING_AMOUNT, r.transfer.GetOutstandingAmount(), utmParams))
	}
}

func (r receiptTextBuilder) translateAndFormatMessage(messageTextToTranslate string, amount models.Amount, utmParams UtmParams) string {
	userID := r.viewerUserID

	counterpartyInfo := r.getReceiptCounterparty()

	var counterpartyText string
	{
		// TODO: Disabled URL due to issue with Telegram parser
		if userID == 0 || utmParams.Medium == UTM_MEDIUM_SMS || utmParams.Medium == telegram_bot.TelegramPlatformID {
			counterpartyText = counterpartyInfo.Name()
		} else {
			counterpartyUrl := GetCounterpartyUrl(counterpartyInfo.ContactID, userID, r.Locale(), utmParams)
			counterpartyText = fmt.Sprintf(`<a href="%v"><b>%v</b></a>`, counterpartyUrl, html.EscapeString(counterpartyInfo.Name()))
		}
		// TODO: Add a @counterparty Telegram nickname if sending receipt to Telegram channel
	}

	//var amountText string
	//{
	//	transferUrl := GetTransferUrlForUser(r.transfer.ID, userID, r.Locale(), utmParams)
	//	if utmParams.Medium == UTM_MEDIUM_SMS {
	//		amountText = fmt.Sprintf(`%v - %v`, r.transfer.GetAmount(), transferUrl)
	//	} else {
	//		amountText = fmt.Sprintf(`<a href="%v">%v</a>`, transferUrl, r.transfer.GetAmount())
	//	}
	//}
	amountText := fmt.Sprintf("%v", amount)

	return r.Translate(messageTextToTranslate, map[string]interface{}{
		"Counterparty": template.HTML(counterpartyText),
		"Amount":       template.HTML(amountText),
	})
}

//func (r receiptBuilder) addBalance(buffer *bytes.Buffer, counterpartyBalance Balance, utmParams UtmParams) string {
//	if counterpartyID == 0 {
//		return ""
//	}
//	showBalanceMessage := true
//	transfer := tm.transfer
//	if len(counterpartyBalance) == 1 {
//		transferAmount := transfer.GetAmount()
//		if singleCurrencyVal, ok := counterpartyBalance[transferAmount.Currency]; !ok || singleCurrencyVal != transfer.Amount {
//			showBalanceMessage = true
//		}
//	}
//	if showBalanceMessage {
//		counterpartyLink := GetCounterpartyLink(whc.AppUserIntID(), whc.Locale(), counterparty.Info(counterpartyID, "", ""), utmParams)
//		buffer.WriteString(BalanceForCounterpartyWithHeader(counterpartyLink, counterpartyBalance, tm.executionContext.Logger(), tm.executionContext))
//	}
//	return buffer.String()
//}
