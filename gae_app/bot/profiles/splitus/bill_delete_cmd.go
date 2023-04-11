package splitus

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/bots-go-framework/bots-api-telegram/tgbotapi"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"net/url"
)

const deleteBillCommandCode = "delete_bill"

var deleteBillCommand = billCallbackCommand(deleteBillCommandCode,
	func(whc botsfw.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m botsfw.MessageFromBot, err error) {
		c := whc.Context()
		if _, err = facade.Bill.DeleteBill(c, bill.ID, whc.AppUserIntID()); err != nil {
			if err == facade.ErrSettledBillsCanNotBeDeleted {
				m.Text = whc.Translate(err.Error())
				err = nil
			}
			return
		}
		m.Text = fmt.Sprintf("Bill #%v has been deleted", bill.ID)
		m.IsEdit = true
		m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
			[]tgbotapi.InlineKeyboardButton{
				{
					Text:         "Restore",
					CallbackData: billCallbackCommandData(restoreBillCommandCode, bill.ID),
				},
			},
		)
		return
	},
)

const restoreBillCommandCode = "restore_bill"

var restoreBillCommand = billCallbackCommand(restoreBillCommandCode,
	func(whc botsfw.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m botsfw.MessageFromBot, err error) {
		c := whc.Context()
		if _, err = facade.Bill.RestoreBill(c, bill.ID, whc.AppUserIntID()); err != nil {
			if err == facade.ErrSettledBillsCanNotBeDeleted {
				m.Text = whc.Translate(err.Error())
				err = nil
			}
			return
		}
		if m.Text, err = getBillCardMessageText(c, whc.GetBotCode(), whc, bill, false, "Bill has been restored"); err != nil {
			return
		}
		m.Format = botsfw.MessageFormatHTML
		m.IsEdit = true
		return
	},
)
