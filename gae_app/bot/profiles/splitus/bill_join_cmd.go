package splitus

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/db"
	"github.com/strongo/decimal"
	"github.com/strongo/log"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/shared_all"
)

const joinBillCommandCode = "join_bill"
const leaceBillCommandCode = "leave_bill"

var joinBillCommand = bots.Command{
	Code: joinBillCommandCode,
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		text := whc.Input().(bots.WebhookTextMessage).Text()
		var bill models.Bill
		if bill.ID = strings.Replace(text, "/start join_bill-", "", 1); bill.ID == "" {
			err = errors.New("Missing bill ID")
			return
		}
		if bill, err = dal.Bill.GetBillByID(whc.Context(), bill.ID); err != nil {
			return
		}
		return joinBillAction(whc, bill, "", false)
	},
	CallbackAction: shared_all.TransactionalCallbackAction(db.CrossGroupTransaction, billCallbackAction(func(whc bots.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "joinBillCommand.CallbackAction()")
		memberStatus := callbackUrl.Query().Get("i")
		return joinBillAction(whc, bill, memberStatus, true)
	})),
}

func joinBillAction(whc bots.WebhookContext, bill models.Bill, memberStatus string, isEditMessage bool) (m bots.MessageFromBot, err error) {
	if bill.ID == "" {
		panic("bill.ID is empty string")
	}
	c := whc.Context()
	log.Debugf(c, "joinBillAction(bill.ID=%v)", bill.ID)
	userID := strconv.FormatInt(whc.AppUserIntID(), 10)

	isAlreadyMember := func(members []models.BillMemberJson) (member models.BillMemberJson, isMember bool) {
		for _, member = range bill.GetBillMembers() {
			if isMember = member.UserID == userID; isMember {
				return
			}
		}
		return
	}

	var appUser bots.BotAppUser
	if appUser, err = whc.GetAppUser(); err != nil {
		return
	}
	user := appUser.(*models.AppUserEntity)

	userName := user.FullName()

	if userName == "" {
		err = errors.New("userName is empty string")
		return
	}

	_, isMember := isAlreadyMember(bill.GetBillMembers())
	if memberStatus == "" && isMember {
		log.Infof(c, "User is already member of the bill before transaction, memberStatus: "+memberStatus)
		callbackAnswer := tgbotapi.NewCallback("", whc.Translate(trans.MESSAGE_TEXT_ALREADY_BILL_MEMBER, userName))
		callbackAnswer.ShowAlert = true
		m.BotMessage = telegram_bot.CallbackAnswer(callbackAnswer)
		whc.LogRequest()
		if update := whc.Input().(telegram_bot.TelegramWebhookInput).TgUpdate(); update.CallbackQuery.Message != nil {
			if m2, err := ShowBillCard(whc, true, bill, ""); err != nil {
				return m, err
			} else if m2.Text != update.CallbackQuery.Message.Text {
				log.Debugf(c, "Need to update bill card")
				if _, err = whc.Responder().SendMessage(c, m2, bots.BotApiSendMessageOverHTTPS); err != nil {
					return m, err
				}
			} else {
				log.Debugf(c, "m.Text: %v", m2.Text)
			}
		}
		return
	}

	var isJoined bool

	var paid decimal.Decimal64p2
	switch memberStatus {
	case "paid":
		paid = bill.AmountTotal
	case "owe":
	default:
	}

	if bill, _, _, isJoined, err = facade.Bill.AddBillMember(c, userID, bill, "", userID, userName, paid); err != nil {
		return
	}

	log.Debugf(c, "isJoined=%v", isJoined)
	if isJoined {
		delayUpdateBillCardOnUserJoin(c, bill.ID, whc.Translate(fmt.Sprintf("%v: ", time.Now())+trans.MESSAGE_TEXT_USER_JOINED_BILL, userName))
	}

	return ShowBillCard(whc, isEditMessage, bill, "")
}
