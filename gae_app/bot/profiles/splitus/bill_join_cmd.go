package splitus

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/shared_group"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/trans"
	"github.com/crediterra/money"
	"github.com/pkg/errors"
	"github.com/strongo/app"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/db"
	"github.com/strongo/decimal"
	"github.com/strongo/log"
	"net/url"
	"strconv"
	"strings"
	//"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/shared_all"
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/shared_all"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"context"
	"fmt"
	"time"
)

const joinBillCommandCode = "join_bill"
const leaveBillCommandCode = "leave_bill"

var joinBillCommand = bots.Command{
	Code: joinBillCommandCode,
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		text := whc.Input().(bots.WebhookTextMessage).Text()
		var bill models.Bill
		if bill.ID = strings.Replace(text, "/start join_bill-", "", 1); bill.ID == "" {
			err = errors.New("Missing bill ID")
			return
		}
		if err = dal.DB.RunInTransaction(whc.Context(), func(c context.Context) (err error) {
			if bill, err = facade.GetBillByID(whc.Context(), bill.ID); err != nil {
				return
			}
			m, err = joinBillAction(whc, bill, "", false)
			return
		}, db.CrossGroupTransaction); err != nil {
			return
		}
		return
	},
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		_ = whc.AppUserIntID() // Make sure we have user before transaction starts, TODO: it smells, should be refactored?
		//
		return shared_all.TransactionalCallbackAction(db.CrossGroupTransaction, billCallbackAction(func(whc bots.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m bots.MessageFromBot, err error) {
			c := whc.Context()
			log.Debugf(c, "joinBillCommand.CallbackAction()")
			memberStatus := callbackUrl.Query().Get("i")
			return joinBillAction(whc, bill, memberStatus, true)
		}))(whc, callbackUrl)
	},
}

func joinBillAction(whc bots.WebhookContext, bill models.Bill, memberStatus string, isEditMessage bool) (m bots.MessageFromBot, err error) {
	if bill.ID == "" {
		panic("bill.ID is empty string")
	}
	c := whc.Context()
	log.Debugf(c, "joinBillAction(bill.ID=%v)", bill.ID)

	userID := strconv.FormatInt(whc.AppUserIntID(), 10)
	var appUser bots.BotAppUser
	if appUser, err = whc.GetAppUser(); err != nil {
		return
	}
	user := appUser.(*models.AppUserEntity)

	isAlreadyMember := func(members []models.BillMemberJson) (member models.BillMemberJson, isMember bool) {
		for _, member = range bill.GetBillMembers() {
			if isMember = member.UserID == userID; isMember {
				return
			}
		}
		return
	}

	_, isMember := isAlreadyMember(bill.GetBillMembers())

	userName := user.FullName()

	if userName == "" {
		err = errors.New("userName is empty string")
		return
	}

	if memberStatus == "" && isMember {
		log.Infof(c, "User is already member of the bill before transaction, memberStatus: "+memberStatus)
		callbackAnswer := tgbotapi.NewCallback("", whc.Translate(trans.MESSAGE_TEXT_ALREADY_BILL_MEMBER, userName))
		callbackAnswer.ShowAlert = true
		m.BotMessage = telegram.CallbackAnswer(callbackAnswer)
		whc.LogRequest()
		if update := whc.Input().(telegram.TgWebhookInput).TgUpdate(); update.CallbackQuery.Message != nil {
			if m2, err := ShowBillCard(whc, true, bill, ""); err != nil {
				return m2, err
			} else if m2.Text != update.CallbackQuery.Message.Text {
				log.Debugf(c, "Need to update bill card")
				if _, err = whc.Responder().SendMessage(c, m2, bots.BotAPISendMessageOverHTTPS); err != nil {
					return m2, err
				}
			} else {
				log.Debugf(c, "m.Text: %v", m2.Text)
			}
		}
		return
	}

	//if err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
	//if bill, err = facade.GetBillByID(c, bill.ID); err != nil {
	//	return
	//}

	billChanged := false
	if bill.Currency == "" {
		guessCurrency := func() money.Currency {
			switch whc.Locale().Code5 {
			case strongo.LocalCodeRuRu:
				return money.CURRENCY_RUB
			case strongo.LocaleCodeDeDE:
				return money.CURRENCY_EUR
			case strongo.LocaleCodeFrFR:
				return money.CURRENCY_EUR
			case strongo.LocaleCodeItIT:
				return money.CURRENCY_EUR
			case strongo.LocaleCodePtPT:
				return money.CURRENCY_EUR
			case strongo.LocaleCodeEnUK:
				return money.CURRENCY_GBP
			default:
				return money.CURRENCY_USD
			}
		}

		if whc.IsInGroup() {
			var group models.Group
			if group, err = shared_group.GetGroup(whc, nil); err != nil {
				return
			}
			if group.GroupEntity != nil {
				if group.DefaultCurrency != "" {
					bill.Currency = group.DefaultCurrency
				} else {
					bill.Currency = guessCurrency()
				}
			}
		} else if user.PrimaryCurrency != "" {
			bill.Currency = money.Currency(user.PrimaryCurrency)
		} else if len(user.LastCurrencies) > 0 {
			bill.Currency = money.Currency(user.LastCurrencies[0])
		}
		if bill.Currency == "" {
			bill.Currency = guessCurrency()
		}
		billChanged = true
	}

	var isJoined bool

	var paid decimal.Decimal64p2
	switch memberStatus {
	case "paid":
		paid = bill.AmountTotal
	case "owe":
	default:
	}

	billChanged2 := false
	if bill, _, billChanged2, isJoined, err = facade.Bill.AddBillMember(c, userID, bill, "", userID, userName, paid); err != nil {
		return
	}
	if billChanged = billChanged2 || billChanged; billChanged {
		if err = dal.Bill.SaveBill(c, bill); err != nil {
			return
		}
		if isJoined {
			delayUpdateBillCardOnUserJoin(c, bill.ID, whc.Translate(fmt.Sprintf("%v: ", time.Now())+trans.MESSAGE_TEXT_USER_JOINED_BILL, userName))
		}
	}
	//return
	//}

	return ShowBillCard(whc, isEditMessage, bill, "")
}
