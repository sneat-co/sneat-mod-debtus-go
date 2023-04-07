package splitus

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/shared_group"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"errors"
	"github.com/DebtsTracker/translations/trans"
	"github.com/crediterra/money"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/decimal"
	"github.com/strongo/log"
	"net/url"
	"regexp"
	"strings"
)

var chosenInlineResultCommand = bots.Command{
	Code:       "chosen-inline-result-command",
	InputTypes: []bots.WebhookInputType{bots.WebhookInputChosenInlineResult},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		log.Debugf(whc.Context(), "splitus.chosenInlineResultHandler.Action()")
		chosenResult := whc.Input().(bots.WebhookChosenInlineResult)
		resultID := chosenResult.GetResultID()
		if strings.HasPrefix(resultID, "bill?") {
			return createBillFromInlineChosenResult(whc, chosenResult)
		}
		return
	},
}

var reDecimal = regexp.MustCompile(`\d+(\.\d+)?`)

func createBillFromInlineChosenResult(whc bots.WebhookContext, chosenResult bots.WebhookChosenInlineResult) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	log.Debugf(c, "createBillFromInlineChosenResult()")

	resultID := chosenResult.GetResultID()

	const prefix = "bill?"

	if !strings.HasPrefix(resultID, prefix) {
		err = errors.New("Unexpected resultID: " + resultID)
		return
	}

	switch {
	case true:
		userID := whc.AppUserStrID()
		var values url.Values
		if values, err = url.ParseQuery(resultID[len(prefix):]); err != nil {
			return
		}
		if lang := values.Get("lang"); lang != "" {
			if err = whc.SetLocale(lang); err != nil {
				return
			}
		}
		var billName string
		if reMatches := reInlineQueryNewBill.FindStringSubmatch(chosenResult.GetQuery()); reMatches != nil {
			billName = strings.TrimSpace(reMatches[3])
		} else {
			billName = whc.Translate(trans.NO_NAME)
		}

		amountStr := values.Get("amount")
		amountIdx := reDecimal.FindStringIndex(amountStr)
		amountNum := amountStr[:amountIdx[1]]
		amountCcy := money.Currency(amountStr[amountIdx[1]:])

		var amount decimal.Decimal64p2
		if amount, err = decimal.ParseDecimal64p2(amountNum); err != nil {
			return
		}
		bill := models.Bill{
			BillEntity: &models.BillEntity{
				BillCommon: models.BillCommon{

					TgInlineMessageIDs: []string{chosenResult.GetInlineMessageID()},
					Name:               billName,
					AmountTotal:        amount,
					Status:             models.STATUS_DRAFT,
					CreatorUserID:      userID,
					UserIDs:            []string{userID},
					SplitMode:          models.SplitModeEqually,
					Currency:           amountCcy,
				},
			},
		}

		//var (
		//	user          bots.BotAppUser
		//	appUserEntity *models.AppUserEntity
		//)
		//if user, err = whc.GetAppUser(); err != nil {
		//	return
		//}
		//appUserEntity = user.(*models.AppUserEntity)
		//_, _, _, _, members := bill.AddOrGetMember(userID, 0, appUserEntity.GetFullName())
		//if err = bill.setBillMembers(members); err != nil {
		//	return
		//}
		//billMember.Paid = bill.AmountTotal
		//switch values.Get("i") {
		//case "paid":
		//	billMember.Paid = bill.AmountTotal
		//case "owe":
		//default:
		//	err = fmt.Errorf("unknown value of 'i' parameter: %v", query.Get("i"))
		//	return
		//}

		defer func() {
			if r := recover(); r != nil {
				whc.LogRequest()
				panic(r)
			}
		}()
		err = dtdal.DB.RunInTransaction(c, func(tc context.Context) (err error) {
			if bill, err = facade.Bill.CreateBill(c, tc, bill.BillEntity); err != nil {
				return
			}
			return
		}, dtdal.SingleGroupTransaction)
		if err != nil {
			err = errors.WithMessage(err, "Failed to call facade.Bill.CreateBill()")
			return
		}
		log.Infof(c, "createBillFromInlineChosenResult() => Bill created")

		botCode := whc.GetBotCode()

		log.Infof(c, "createBillFromInlineChosenResult() => suxx 0")

		footer := strings.Repeat("―", 15) + "\n" + whc.Translate(trans.MESSAGE_TEXT_ASK_BILL_PAYER)

		if m.Text, err = getBillCardMessageText(c, botCode, whc, bill, false, footer); err != nil {
			log.Errorf(c, "Failed to create bill card")
			return
		} else if strings.TrimSpace(m.Text) == "" {
			err = errors.New("getBillCardMessageText() returned empty string")
			log.Errorf(c, err.Error())
			return
		}

		log.Infof(c, "createBillFromInlineChosenResult() => suxx 1")

		if m, err = whc.NewEditMessage(m.Text, bots.MessageFormatHTML); err != nil { // TODO: Unnecessary hack?
			log.Infof(c, "createBillFromInlineChosenResult() => suxx 1.2")
			log.Errorf(c, err.Error())
			return
		}

		log.Infof(c, "createBillFromInlineChosenResult() => suxx 2")

		m.Keyboard = getWhoPaidInlineKeyboard(whc, bill.ID)

		var response bots.OnMessageSentResponse
		log.Debugf(c, "createBillFromInlineChosenResult() => Sending bill card: %v", m)

		if response, err = whc.Responder().SendMessage(c, m, bots.BotAPISendMessageOverHTTPS); err != nil {
			log.Errorf(c, "createBillFromInlineChosenResult() => %v", err)
			return
		}

		log.Debugf(c, "response: %v", response)
		m.Text = bots.NoMessageToSend
	}

	return
}

var reBillUrl = regexp.MustCompile(`\?start=bill-(\d+)$`)

func getBillIDFromUrlInEditedMessage(whc bots.WebhookContext) (billID string) {
	tgInput, ok := whc.Input().(telegram.TgWebhookInput)
	if !ok {
		return
	}
	tgUpdate := tgInput.TgUpdate()
	if tgUpdate.EditedMessage == nil {
		return
	}
	if tgUpdate.EditedMessage.Entities == nil {
		return
	}
	for _, entity := range *tgUpdate.EditedMessage.Entities {
		if entity.Type == "text_link" {
			if s := reBillUrl.FindStringSubmatch(entity.URL); len(s) != 0 {
				billID = s[1]
				if billID == "" {
					log.Errorf(whc.Context(), "Missing bill ID")
				}
				return
			}
		}
	}
	return
}

var EditedBillCardHookCommand = bots.Command{ // TODO: seems to be not used anywhere
	Code: "edited-bill-card",
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		whc.LogRequest()
		c := whc.Context()
		billID := getBillIDFromUrlInEditedMessage(whc)
		log.Debugf(c, "editedBillCardHookCommand.Action() => billID: %d", billID)
		if billID == "" {
			panic("billID is empty string")
		}

		m.Text = bots.NoMessageToSend

		var groupID string
		if groupID, err = shared_group.GetUserGroupID(whc); err != nil {
			return
		} else if groupID == "" {
			log.Warningf(c, "group.ID is empty string")
			return
		}

		changed := false
		err = dtdal.DB.RunInTransaction(c, func(c context.Context) error {
			var bill models.Bill
			if bill, err = facade.GetBillByID(c, billID); err != nil {
				return err
			}

			if groupID != "" && bill.GetUserGroupID() != groupID { // TODO: Should we check for empty bill.GetUserGroupID() or better fail?
				if bill, _, err = facade.Bill.AssignBillToGroup(c, bill, groupID, whc.AppUserStrID()); err != nil {
					return err
				}
				changed = true
			}

			if changed {
				return dtdal.Bill.SaveBill(c, bill)
			}

			return err
		}, dtdal.CrossGroupTransaction)
		if err != nil {
			return
		}
		if changed {
			log.Debugf(c, "Bill updated with group ID")
		}
		return
	},
	Matcher: func(command bots.Command, whc bots.WebhookContext) (result bool) {
		result = whc.IsInGroup() && getBillIDFromUrlInEditedMessage(whc) != ""
		log.Debugf(whc.Context(), "editedBillCardHookCommand.Matcher(): %v", result)
		return
	},
}
