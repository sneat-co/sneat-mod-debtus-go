package dtb_transfer

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/analytics"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app"
	"github.com/strongo/app/log"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"html/template"
	"net/url"
	"strings"
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

const CREATE_INVITE_IF_NO_INLINE_NOTIFICATION = "create-invite-if-no-inline-notification"

var CreateInviteIfNoInlineNotificationCommand = bots.Command{
	Code:       CREATE_INVITE_IF_NO_INLINE_NOTIFICATION,
	InputTypes: []bots.WebhookInputType{bots.WebhookInputCallbackQuery},
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		return InlineReceipt(whc, whc.Input().(bots.WebhookCallbackQuery).GetInlineMessageID(), callbackUrl)
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
	receiptID := values.Get("id")
	if cleanReceiptID := strings.Trim(receiptID, " \",.;!@#$%^&*(){}[]`~?/\\|"); cleanReceiptID != receiptID {
		log.Debugf(c, "Unclean receipt ID: %v, cleaned: %v", receiptID, cleanReceiptID)
		receiptID = cleanReceiptID
	}
	transferID, err := common.DecodeID(receiptID)
	if err != nil {
		log.Warningf(c, "Failed to decode receipt?id=[%v]", receiptID)
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
	receiptUrl := getReceiptUrl(whc.GetBotCode(), receiptID, whc.Locale().Code5)
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
					Text:      getInlineReceiptAnnouncementMessage(whc, false, fmt.Sprintf("%v %v", creator.GetFirstName(), creator.GetLastName()), receiptUrl),
					ParseMode: "HTML",
				},
				ReplyMarkup: &tgbotapi.InlineKeyboardMarkup{
					InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
						[]tgbotapi.InlineKeyboardButton{
							{
								Text:         whc.Translate(trans.COMMAND_TEXT_WAIT_A_SECOND),
								CallbackData: CREATE_INVITE_IF_NO_INLINE_NOTIFICATION + fmt.Sprintf("?id=%v", common.EncodeID(transferID)),
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

func getInlineReceiptAnnouncementMessage(t strongo.SingleLocaleTranslator, inviteCreated bool, creator, receiptUrl string) string {
	data := map[string]interface{}{
		"Creator":  creator,
		"SiteLink": template.HTML(`<a href="https://debtstracker.io/#utm_source=telegram&utm_medium=bot&utm_campaign=receipt-inline-1st">DebtsTracker.IO</a>`),
		"ReceiptUrl": receiptUrl,
	}
	result := t.Translate(trans.INLINE_RECEIPT_MESSAGE, data)
	if inviteCreated {
		result += "\n\n" + t.Translate(trans.INLINE_RECEIPT_CHOOSE_LANGUAGE, data)
	}
	return result
}

func InlineReceipt(whc bots.WebhookContext, inlineMessageID string, queryUrl *url.URL) (m bots.MessageFromBot, err error) {
	c := whc.Context()

	log.Debugf(c, "InlineReceipt(queryUrl: %v)", queryUrl)
	transferEncodedID := queryUrl.Query().Get("id")
	transferID, err := common.DecodeID(transferEncodedID)
	if err != nil {
		return m, err
	}
	creator := whc.GetSender()
	creatorName := fmt.Sprintf("%v %v", creator.GetFirstName(), creator.GetLastName())

	m, err = CreateReceiptAndEditMessage(
		whc,
		inlineMessageID,
		transferID,
		creatorName,
		//whc.Translate(trans.COMMAND_TEXT_SEE_RECEIPT_DETAILS),
	)
	m.DisableWebPagePreview = true

	analytics.ReceiptSentFromBot(whc, "telegram")

	//_, err = whc.Responder().SendMessage(c, m, bots.BotApiSendMessageOverHTTPS)
	//if err != nil {
	//	log.Errorf(c, "Failed to send inline response: %v", err.Error())
	//}
	//m = whc.NewMessage("")
	return
}
