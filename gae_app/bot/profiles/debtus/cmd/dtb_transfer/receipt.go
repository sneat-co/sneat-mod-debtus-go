package dtb_transfer

import (
	"bytes"
	"fmt"
	"html/template"
	"net/url"
	"strings"
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/analytics"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bitbucket.com/asterus/debtstracker-server/gae_app/general"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/log"
)

//func InlineAcceptTransfer(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
//	inlineQuery := whc.InputInlineQuery()
//	m.TelegramInlineCongig = &tgbotapi.InlineConfig{
//		InlineQueryID: inlineQuery.GetInlineQueryID(),
//		SwitchPMText: "Accept transfer",
//		SwitchPMParameter: "accept?transfer=ABC",
//	}
//	return m, err
//}

const CREATE_RECEIPT_IF_NO_INLINE_CHOOSEN_NOTIFICATION = "create-receipt"

var CreateReceiptIfNoInlineNotificationCommand = bots.Command{
	Code:       CREATE_RECEIPT_IF_NO_INLINE_CHOOSEN_NOTIFICATION,
	InputTypes: []bots.WebhookInputType{bots.WebhookInputCallbackQuery},
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		return OnInlineChoosenCreateReceipt(whc, whc.Input().(bots.WebhookCallbackQuery).GetInlineMessageID(), callbackUrl)
	},
}

func InlineSendReceipt(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	log.Debugf(c, "InlineSendReceipt()")
	inlineQuery := whc.Input().(bots.WebhookInlineQuery)
	query := inlineQuery.GetQuery()
	values, err := url.ParseQuery(query[len("receipt?"):])
	if err != nil {
		return m, err
	}
	idParam := values.Get("id")
	if cleanID := strings.Trim(idParam, " \",.;!@#$%^&*(){}[]`~?/\\|"); cleanID != idParam {
		log.Debugf(c, "Unclean receipt ID: %v, cleaned: %v", idParam, cleanID)
		idParam = cleanID
	}
	transferID, err := common.DecodeID(idParam)
	if err != nil {
		log.Warningf(c, "Failed to decode receipt?id=[%v]", idParam)
		return m, err
	}
	var transfer models.Transfer
	transfer, err = dal.Transfer.GetTransferByID(c, transferID)
	if err != nil {
		log.Infof(c, "Faield to get transfer by ID: %v", transferID)
		return m, err
	}

	log.Debugf(c, "Loaded transfer: %v", transfer)
	creator := whc.GetSender()

	m.BotMessage = telegram_bot.InlineBotMessage(tgbotapi.InlineConfig{
		InlineQueryID: inlineQuery.GetInlineQueryID(),
		//SwitchPmText: "Accept invite",
		//SwitchPmParameter: "invite?code=ABC",
		Results: []interface{}{
			tgbotapi.InlineQueryResultArticle{
				ID:          query,
				Type:        "article",                                                          // TODO: Move to constructor
				ThumbURL:    "https://debtstracker-io.appspot.com/img/debtstracker-512x512.png", //TODO: Replace with receipt image
				ThumbHeight: 512,
				ThumbWidth:  512,
				Title:       fmt.Sprintf(whc.Translate(trans.INLINE_RECEIPT_TITLE), transfer.GetAmount()),
				Description: whc.Translate(trans.INLINE_RECEIPT_DESCRIPTION),
				InputMessageContent: tgbotapi.InputTextMessageContent{
					Text:      getInlineReceiptMessageText(whc, whc.GetBotCode(), whc.Locale().Code5, fmt.Sprintf("%v %v", creator.GetFirstName(), creator.GetLastName()), 0),
					ParseMode: "HTML",
				},
				ReplyMarkup: &tgbotapi.InlineKeyboardMarkup{
					InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
						{
							{
								Text:         whc.Translate(trans.COMMAND_TEXT_WAIT_A_SECOND),
								CallbackData: CREATE_RECEIPT_IF_NO_INLINE_CHOOSEN_NOTIFICATION + fmt.Sprintf("?id=%v", common.EncodeID(transferID)),
							},
						},
					},
				},
			},
		},
	})
	log.Debugf(c, "MessageFromBot: %v", m)

	//log.Debugf(c, "Calling botApi.Send(inlineConfig=%v)", inlineConfig)
	//
	//botApi := &tgbotapi.BotAPI{
	//	Token:  whc.GetBotToken(),
	//	Debug:  true,
	//	Client: whc.GetHttpClient(),
	//}
	//mes, err := botApi.AnswerInlineQuery(inlineConfig)
	//if err != nil {
	//	log.Errorf(c, "Failed to send inline results: %v", err)
	//}
	//s, err := json.Marshal(mes)
	//if err != nil {
	//	log.Errorf(c, "Failed to marshal inline results response: %v, %v", err, mes)
	//}
	//log.Infof(c, "botApi.Send(inlineConfig): %v", string(s))
	return m, err
}

func getInlineReceiptMessageText(t strongo.SingleLocaleTranslator, botCode, localeCode5, creator string, receiptID int64) string {
	data := map[string]interface{}{
		"Creator":  creator,
		"SiteLink": template.HTML(`<a href="https://debtstracker.io/#utm_source=telegram&utm_medium=bot&utm_campaign=receipt-inline">DebtsTracker.IO</a>`),
	}
	if receiptID != 0 {
		data["ReceiptUrl"] = GetUrlForReceiptInTelegram(botCode, receiptID, localeCode5)
	}
	var buf bytes.Buffer
	if receiptID == 0 {
		buf.WriteString(t.Translate(trans.INLINE_RECEIPT_GENERATING_MESSAGE, data))
	} else {
		buf.WriteString(t.Translate(trans.INLINE_RECEIPT_MESSAGE, data))
	}

	//buf.WriteString("\n\n" + t.Translate(trans.INLINE_RECEIPT_FOOTER, data))

	if receiptID != 0 {
		buf.WriteString("\n\n" + t.Translate(trans.INLINE_RECEIPT_CHOOSE_LANGUAGE, data))
	}

	return buf.String()
}

func OnInlineChoosenCreateReceipt(whc bots.WebhookContext, inlineMessageID string, queryUrl *url.URL) (m bots.MessageFromBot, err error) {
	c := whc.Context()

	log.Debugf(c, "OnInlineChoosenCreateReceipt(queryUrl: %v)", queryUrl)
	transferEncodedID := queryUrl.Query().Get("id")
	transferID, err := common.DecodeID(transferEncodedID)
	if err != nil {
		return m, err
	}
	creator := whc.GetSender()
	creatorName := fmt.Sprintf("%v %v", creator.GetFirstName(), creator.GetLastName())

	transfer, err := dal.Transfer.GetTransferByID(c, transferID)
	if err != nil {
		return m, err
	}
	receipt := models.NewReceiptEntity(whc.AppUserIntID(), transferID, transfer.Counterparty().UserID, whc.Locale().Code5, telegram_bot.TelegramPlatformID, "", general.CreatedOn{
		CreatedOnID:       whc.GetBotCode(), // TODO: Replace with method call.
		CreatedOnPlatform: whc.BotPlatform().Id(),
	})
	receiptID, err := dal.Receipt.CreateReceipt(c, &receipt)
	if err != nil {
		return m, err
	}

	dal.Receipt.DelayedMarkReceiptAsSent(c, receiptID, transferID, time.Now())
	m, err = showReceiptAnnouncement(whc, receiptID, creatorName)

	analytics.ReceiptSentFromBot(whc, "telegram")

	//_, err = whc.Responder().SendMessage(c, m, bots.BotApiSendMessageOverHTTPS)
	//if err != nil {
	//	log.Errorf(c, "Failed to send inline response: %v", err.Error())
	//}
	//m = whc.NewMessage("")
	return
}
