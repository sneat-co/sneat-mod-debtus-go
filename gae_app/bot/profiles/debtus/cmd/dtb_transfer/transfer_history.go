package dtb_transfer

import (
	"bitbucket.com/debtstracker/gae_app/bot/profiles/debtus/cmd/dtb_general"
	"bitbucket.com/debtstracker/gae_app/debtstracker/common"
	"bitbucket.com/debtstracker/gae_app/debtstracker/dal"
	"github.com/DebtsTracker/translations/emoji"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
	"bytes"
	"fmt"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"net/url"
	"strings"
	"time"
	"github.com/yaa110/go-persian-calendar/ptime"
)

const HistoryTopLimit = 5
const HistoryMoreLimit = 10

const HISTORY_COMMAND = "history"

var HistoryCommand = bots.Command{
	Code:     HISTORY_COMMAND,
	Icon:     emoji.HISTORY_ICON,
	Title:    trans.COMMAND_TEXT_HISTORY,
	Commands: trans.Commands(trans.COMMAND_HISTORY, emoji.QUESTION_ICON), // TODO: Check icon!
	Titles:   map[string]string{bots.SHORT_TITLE: emoji.QUESTION_ICON}, // TODO: Check icon!
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		return showHistoryCard(whc, HistoryTopLimit)
	},
}

func showHistoryCard(whc bots.WebhookContext, limit int) (m bots.MessageFromBot, err error) {
	c := whc.Context()

	transfers, hasMore, err := dal.Transfer.LoadTransfersByUserID(c, whc.AppUserIntID(), 0, limit)

	if len(transfers) == 0 {
		m = whc.NewMessage(whc.Translate(trans.MESSAGE_TEXT_HISTORY_NO_RECORDS) + common.HORIZONTAL_LINE + dtb_general.AdSlot(whc, UTM_CAMPAIGN_TRANSFER_HISTORY))
	} else {
		m = whc.NewMessage(whc.Translate(
			trans.MESSAGE_TEXT_HISTORY_LIST,
			whc.Translate(trans.MESSAGE_TEXT_HISTORY_HEADER),
			len(transfers),
			transferHistoryRows(whc, transfers),
		) + common.HORIZONTAL_LINE + dtb_general.AdSlot(whc, UTM_CAMPAIGN_TRANSFER_HISTORY))
		if hasMore {
			transfers = transfers[:limit]
			utmParams := common.FillUtmParams(whc, common.UtmParams{Campaign: UTM_CAMPAIGN_TRANSFER_HISTORY})
			m.Keyboard = &tgbotapi.InlineKeyboardMarkup{
				InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
					{
						tgbotapi.NewInlineKeyboardButtonURL(
							whc.Translate(trans.INLINE_BUTTON_SHOW_FULL_HISTORY),
							//fmt.Sprintf("transfer-history?offset=%v", len(transfers)),
							fmt.Sprintf("https://debtstracker.io/%v/history?user=%v#%v", whc.Locale().SiteCode(), common.EncodeID(whc.AppUserIntID()), utmParams),
						),
					},
				},
			}
		}
	}

	m.Format = bots.MessageFormatHTML
	m.DisableWebPagePreview = true
	return m, nil
}

const (
	UTM_CAMPAIGN_TRANSFER_HISTORY = "transfer-history"
)

func transferHistoryRows(whc bots.WebhookContext, transfers []models.Transfer) string {
	var s bytes.Buffer
	for _, transfer := range transfers {
		isCreator := whc.AppUserIntID() == transfer.CreatorUserID
		var counterpartyName string
		if isCreator {
			counterpartyName = transfer.Counterparty().ContactName
		} else {
			counterpartyName = transfer.Creator().ContactName
		}
		amount := fmt.Sprintf(`<a href="%v">%s</a>`,
			common.GetTransferUrlForUser(
				transfer.ID,
				whc.AppUserIntID(),
				whc.Locale(),
				common.NewUtmParams(whc, "history"),
			),
			transfer.GetAmount(),
		)
		if (isCreator && transfer.Direction() == models.TransferDirectionUser2Counterparty) || (!isCreator && transfer.Direction() == models.TransferDirectionCounterparty2User) {
			s.WriteString(whc.Translate(trans.MESSAGE_TEXT_HISTORY_ROW_FROM_USER_WITH_NAME, shortDate(transfer.DtCreated, whc), counterpartyName, amount))
		} else {
			s.WriteString(whc.Translate(trans.MESSAGE_TEXT_HISTORY_ROW_TO_USER_WITH_NAME, shortDate(transfer.DtCreated, whc), counterpartyName, amount))
		}
		s.WriteString("\n\n")
	}
	return strings.TrimSpace(s.String())
}

var TransferHistoryCallbackCommand = bots.NewCallbackCommand("transfer-history", callbackTransferHistory)

func callbackTransferHistory(whc bots.WebhookContext, _ *url.URL) (bots.MessageFromBot, error) {
	return whc.NewMessage("TODO: Show more history records"), nil
}

func shortDate(t time.Time, translator strongo.SingleLocaleTranslator) string {
	switch translator.Locale().Code5 {
	case strongo.LOCALE_EN_US:
		return t.Format("02 Jan 2006")
	case strongo.LOCALE_FA_IR:
		pt := ptime.New(t)
		return pt.Format("dd MMM yyyy")
	default:
		month := t.Format("Jan")
		return fmt.Sprintf("%v %v %v", t.Format("02"), translator.Translate(month), t.Format("2006"))
	}
}
