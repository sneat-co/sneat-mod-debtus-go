package bot_shared

import (
	"net/url"
	"regexp"
	"strings"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/decimal"
	"github.com/strongo/log"
	"golang.org/x/net/context"
)

func choosenInlineResultHandler(botParams BotParams) bots.Command {
	return bots.Command{
		Code:       "choosen-inline-result-command",
		InputTypes: []bots.WebhookInputType{bots.WebhookInputChosenInlineResult},
		Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
			log.Debugf(whc.Context(), "splitus.choosenInlineResultHandler.Action()")
			choosenResult := whc.Input().(bots.WebhookChosenInlineResult)
			resultID := choosenResult.GetResultID()
			if strings.HasPrefix(resultID, "bill?") {
				return createBillFromInlineChoosenResult(whc, botParams, choosenResult)
			}
			return
		},
	}
}

var reDecimal = regexp.MustCompile(`\d+(\.\d+)?`)

func createBillFromInlineChoosenResult(whc bots.WebhookContext, botParams BotParams, choosenResult bots.WebhookChosenInlineResult) (m bots.MessageFromBot, err error) {
	c := whc.Context()

	resultID := choosenResult.GetResultID()

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
		if reMatches := reInlineQueryNewBill.FindStringSubmatch(choosenResult.GetQuery()); reMatches != nil {
			billName = strings.TrimSpace(reMatches[3])
		} else {
			billName = whc.Translate(trans.NO_NAME)
		}

		amountStr := values.Get("amount")
		amountIdx := reDecimal.FindStringIndex(amountStr)
		amountNum := amountStr[:amountIdx[1]]
		amountCcy := models.Currency(amountStr[amountIdx[1]:])

		var amount decimal.Decimal64p2
		if amount, err = decimal.ParseDecimal64p2(amountNum); err != nil {
			return
		}
		bill := models.Bill{
			BillEntity: &models.BillEntity{
				BillCommon: models.BillCommon{

					TgInlineMessageIDs: []string{choosenResult.GetInlineMessageID()},
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
		//_, _, _, _, members := bill.AddOrGetMember(userID, 0, appUserEntity.FullName())
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
		err = dal.DB.RunInTransaction(c, func(tc context.Context) (err error) {
			if bill, err = facade.Bill.CreateBill(c, tc, bill.BillEntity); err != nil {
				return
			}
			return
		}, dal.SingleGroupTransaction)
		if err != nil {
			err = errors.WithMessage(err, "Failed to call facade.Bill.CreateBill()")
			return
		}
		log.Infof(c, "createBillFromInlineChoosenResult() => Bill created")

		botCode := whc.GetBotCode()

		log.Infof(c, "createBillFromInlineChoosenResult() => suxx 0")

		footer := strings.Repeat("―", 21) + "\n"

		if bill.Currency == "" {
			footer += whc.Translate(trans.MESSAGE_TEXT_ASK_BILL_CURRENCY)
		} else {
			footer += whc.Translate(trans.MESSAGE_TEXT_ASK_BILL_PAYER)
		}

		if m.Text, err = GetBillCardMessageText(c, botCode, whc, bill, false, footer); err != nil {
			log.Errorf(c, "Failed to create bill card")
			return
		} else if strings.TrimSpace(m.Text) == "" {
			err = errors.New("GetBillCardMessageText() returned empty string")
			log.Errorf(c, err.Error())
			return
		}

		log.Infof(c, "createBillFromInlineChoosenResult() => suxx 1")

		if m, err = whc.NewEditMessage(m.Text, bots.MessageFormatHTML); err != nil { // TODO: Unnecessary hack?
			log.Infof(c, "createBillFromInlineChoosenResult() => suxx 1.2")
			log.Errorf(c, err.Error())
			return
		}

		log.Infof(c, "createBillFromInlineChoosenResult() => suxx 2")

		if bill.Currency == "" {
			m.Keyboard = CurrenciesInlineKeyboard(BillCallbackCommandData(SET_BILL_CURRENCY_COMMAND, bill.ID))
		} else {
			m.Keyboard = botParams.OnAfterBillCurrencySelected(whc, bill.ID)
		}

		var response bots.OnMessageSentResponse
		log.Debugf(c, "createBillFromInlineChoosenResult() => Sending bill card: %v", m)

		if response, err = whc.Responder().SendMessage(c, m, bots.BotApiSendMessageOverHTTPS); err != nil {
			log.Errorf(c, "createBillFromInlineChoosenResult() => %v", err)
			return
		}

		log.Debugf(c, "response: %v", response)
		m.Text = bots.NoMessageToSend
	}

	return
}

var reBillUrl = regexp.MustCompile(`\?start=bill-(\d+)$`)

func getBillIDFromUrlInEditedMessage(whc bots.WebhookContext) (billID string) {
	tgInput, ok := whc.Input().(telegram_bot.TelegramWebhookInput)
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
		if groupID, err = GetUserGroupID(whc); err != nil {
			return
		} else if groupID == "" {
			log.Warningf(c, "group.ID is empty string")
			return
		}

		changed := false
		err = dal.DB.RunInTransaction(c, func(c context.Context) error {
			var bill models.Bill
			if bill, err = dal.Bill.GetBillByID(c, billID); err != nil {
				return err
			}

			if groupID != "" && bill.UserGroupID() != groupID { // TODO: Should we check for empty bill.UserGroupID() or better fail?
				if bill, _, err = facade.Bill.AssignBillToGroup(c, bill, groupID, whc.AppUserStrID()); err != nil {
					return err
				}
				changed = true
			}

			if changed {
				return dal.Bill.SaveBill(c, bill)
			}

			return err
		}, dal.CrossGroupTransaction)
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
