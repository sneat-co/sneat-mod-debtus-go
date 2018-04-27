package splitus

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/platforms/tgbots"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"fmt"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app"
	"github.com/strongo/app/gae"
	"github.com/strongo/app/gaestandard"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/log"
	"google.golang.org/appengine/delay"
	"google.golang.org/appengine/urlfetch"
	"strings"
)

var (
	delayUpdateBillCards      = delay.Func("UpdateBillCards", delayedUpdateBillCards)
	delayUpdateBillTgChatCard = delay.Func("UpdateBillTgChatCard", delayedUpdateBillTgChartCard)
)

func delayUpdateBillCardOnUserJoin(c context.Context, billID string, message string) error {
	if err := gae.CallDelayFunc(
		c,
		common.QUEUE_BILLS,
		"update-bill-cards",
		delayUpdateBillCards,
		billID,
		message,
	); err != nil {
		log.Errorf(c, "Failed to queue update of bill cards: %v", err)
	}
	return nil
}

func delayedUpdateBillCards(c context.Context, billID string, footer string) error {
	log.Debugf(c, "delayedUpdateBillCards(billID=%d)", billID)
	if bill, err := dal.Bill.GetBillByID(c, billID); err != nil {
		return err
	} else {
		for _, tgChatMessageID := range bill.TgChatMessageIDs {
			if err = gae.CallDelayFunc(c, common.QUEUE_BILLS, "update-bill-tg-chat-card", delayUpdateBillTgChatCard, billID, tgChatMessageID, footer); err != nil {
				log.Errorf(c, "Failed to queue updated for %v: %v", tgChatMessageID, err)
				return err
			}
		}
	}
	return nil
}

func delayedUpdateBillTgChartCard(c context.Context, billID string, tgChatMessageID, footer string) error {
	log.Debugf(c, "delayedUpdateBillTgChartCard(billID=%d, tgChatMessageID=%v)", billID, tgChatMessageID)
	if bill, err := dal.Bill.GetBillByID(c, billID); err != nil {
		return err
	} else {
		ids := strings.Split(tgChatMessageID, "@")
		inlineMessageID, botCode, localeCode5 := ids[0], ids[1], ids[2]
		translator := strongo.NewSingleMapTranslator(strongo.GetLocaleByCode5(localeCode5), strongo.NewMapTranslator(c, trans.TRANS))

		editMessage := tgbotapi.NewEditMessageText(0, 0, inlineMessageID, "")
		editMessage.ParseMode = "HTML"
		editMessage.DisableWebPagePreview = true

		if err := updateInlineBillCardMessage(c, translator, true, editMessage, bill, botCode, footer); err != nil {
			return err
		} else {
			telegramBots := tgbots.Bots(gaestandard.GetEnvironment(c), nil)
			botSettings, ok := telegramBots.ByCode[botCode]
			if !ok {
				log.Errorf(c, "No bot settings for bot: "+botCode)
				return nil
			}

			tgApi := tgbotapi.NewBotAPIWithClient(botSettings.Token, urlfetch.Client(c))
			if _, err := tgApi.Send(editMessage); err != nil {
				log.Errorf(c, "Failed to sent message to Telegram: %v", err)
				return err
			}
		}
	}
	return nil
}

func updateInlineBillCardMessage(c context.Context, translator strongo.SingleLocaleTranslator, isGroupChat bool, editedMessage *tgbotapi.EditMessageTextConfig, bill models.Bill, botCode string, footer string) (err error) {
	if bill.ID == "" {
		panic("bill.ID is empty string")
	}
	if bill.BillEntity == nil {
		panic("bill.BillEntity == nil")
	}

	if editedMessage.Text, err = getBillCardMessageText(c, botCode, translator, bill, true, footer); err != nil {
		return
	}
	if isGroupChat {
		editedMessage.ReplyMarkup = getPublicBillCardInlineKeyboard(translator, botCode, bill.ID)
	} else {
		editedMessage.ReplyMarkup = getPrivateBillCardInlineKeyboard(translator, botCode, bill)
	}

	return
}

func getPublicBillCardInlineKeyboard(translator strongo.SingleLocaleTranslator, botCode string, billID string) *tgbotapi.InlineKeyboardMarkup {
	goToBotLink := func(command string) string {
		return fmt.Sprintf("https://t.me/%v?start=%v-%v", botCode, command, billID)
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			{
				Text: translator.Translate(trans.BUTTON_TEXT_JOIN),
				URL:  goToBotLink(joinBillCommandCode),
			},
		},
		[]tgbotapi.InlineKeyboardButton{
			{
				Text: translator.Translate(trans.BUTTON_TEXT_EDIT_BILL),
				URL:  goToBotLink(editBillCommandCode),
			},
			{
				Text:         translator.Translate(trans.BUTTON_TEXT_DUE, translator.Translate(trans.NOT_SET)),
				CallbackData: billCallbackCommandData(setBillDueDateCommandCode, billID),
			},
		},
	)
	return keyboard
}
