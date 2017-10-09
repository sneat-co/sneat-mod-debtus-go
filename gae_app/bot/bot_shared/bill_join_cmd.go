package bot_shared

import (
	"github.com/strongo/bots-framework/core"
	"net/url"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/trans"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"golang.org/x/net/context"
	"github.com/pkg/errors"
	"github.com/strongo/app/log"
	"strings"
	"github.com/strongo/bots-api-telegram"
	"fmt"
	"time"
	"github.com/strongo/bots-framework/platforms/telegram"
)

const JOIN_BILL_COMMAND = "join_bill"
const LEAVE_BILL_COMMAND = "leave_bill"

func JoinBillCommand(botParams BotParams) bots.Command {
	return bots.Command{
		Code: JOIN_BILL_COMMAND,
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
			return joinBillAction(whc, botParams, bill, "", false)
		},
		CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
			c := whc.Context()
			log.Debugf(c, "joinBillCommand.CallbackAction()")
			var bill models.Bill
			if bill, err = getBill(c, callbackUrl); err != nil {
				return m, err
			}
			memberStatus := callbackUrl.Query().Get("i")

			return joinBillAction(whc, botParams, bill, memberStatus, true)
		},
	}
}

func joinBillAction(whc bots.WebhookContext, botParams BotParams, bill models.Bill, memberStatus string, isEditMessage bool) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	userID := whc.AppUserIntID()

	isAlreadyMember := func(members []models.BillMemberJson) (member models.BillMemberJson, isMember bool) {
		for _, member = range bill.GetBillMembers() {
			if isMember = member.UserID == userID; isMember {
				return
			}
		}
		return
	}

	var group models.Group
	if group, err = GetGroup(whc); err != nil {
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
			if m2, err := ShowBillCard(whc, botParams, true, bill, ""); err != nil {
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

	isJoined := false
	changed := false

	if err = dal.DB.RunInTransaction(c, func(c context.Context) error {
		if bill, err = dal.Bill.GetBillByID(c, bill.ID); err != nil {
			return err
		}

		var (
			index int
			member models.BillMemberJson
			members []models.BillMemberJson
		)
		_, changed, index, member, members = bill.AddOrGetMember(userID, 0, userName)

		switch memberStatus {
		case "paid":
			if member.Paid != bill.AmountTotal {
				member.Paid = bill.AmountTotal
				changed = true
			}
		case "owe":
		default:
		}
		if !changed {
			return nil
		}
		members[index] = member

		if err := bill.SetBillMembers(members); err != nil {
			return err
		}

		if group.ID != "" {
			changed = bill.AddUserGroupID(group.ID) || changed
		}

		if changed {
			if err := dal.Bill.UpdateBill(c, bill); err != nil {
				return err
			}
		}
		isJoined = true
		return nil
	}, dal.SingleGroupTransaction); err != nil {
		return
	}

	log.Debugf(c, "isJoined=%v", isJoined)
	if isJoined {
		botParams.DelayUpdateBillCardOnUserJoin(c, bill.ID, whc.Translate(fmt.Sprintf("%v: ", time.Now())+trans.MESSAGE_TEXT_USER_JOINED_BILL, userName))
	}

	return ShowBillCard(whc, botParams, isEditMessage, bill, "")
}


