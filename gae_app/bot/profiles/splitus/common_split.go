package splitus

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bytes"
	"errors"
	"fmt"
	"github.com/strongo/app"
	"github.com/strongo/app/log"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/decimal"
	"html"
	"net/url"
	"strconv"
)

func editSplitCallbackAction(
	whc bots.WebhookContext,
	callbackURL *url.URL,
	editCommandPrefix, backCommandPrefix string,
	msgTextAskToSplit string,
	members []models.MemberJson,
	totalAmount models.Amount,
	writeTitle func(buffer *bytes.Buffer) error,
	addShares func(memberID string, addValue int) (member models.MemberJson, err error),
) (m bots.MessageFromBot, err error) {
	c := whc.Context()

	q := callbackURL.Query()

	var (
		addValue int
		member   models.MemberJson
	)

	if member, addValue, err = getSplitParamsAndCurrentMember(q, members); err != nil {
		return
	}

	log.Debugf(c, "current member: %v", member)

	if addValue != 0 {
		log.Debugf(c, "add=%d", addValue)

		if member, err = addShares(member.ID, addValue); err != nil {
			return
		}
		for i, m := range members {
			if m.ID == member.ID {
				members[i] = member
				break
			}
		}
	}

	buffer := new(bytes.Buffer)

	if writeTitle != nil {
		if err = writeTitle(buffer); err != nil {
			return
		}
		buffer.WriteString("\n\n")
	}

	fmt.Fprintf(buffer, "<b>%v</b>\n\n", whc.Translate(msgTextAskToSplit))

	writeSplitMembers(buffer, members, member.ID, totalAmount)

	writeSplitInstructions(buffer, member.TgUserID, member.Name)

	m.Text = buffer.String()
	m.Format = bots.MessageFormatHTML

	tgKeyboard := &tgbotapi.InlineKeyboardMarkup{}
	tgKeyboard.InlineKeyboard = addEditSplitInlineKeyboardButtons(tgKeyboard.InlineKeyboard, whc,
		editCommandPrefix+"&m="+member.ID+"&",
		backCommandPrefix,
	)
	m.Keyboard = tgKeyboard

	m.IsEdit = true
	return
}

func getSplitParamsAndCurrentMember(q url.Values, members []models.MemberJson) (member models.MemberJson, add int, err error) {
	if len(members) == 0 {
		err = errors.New("len(members) == 0")
		return
	}

	if memberID := q.Get("m"); memberID == "" {
		member = members[0]
	} else if memberID == "0" {
		err = errors.New("parameter 'm' is 0")
		return
	} else {
		member.ID = q.Get("m")
		var (
			i    int
			m    models.MemberJson
			move string
		)
		for i, m = range members {
			if m.ID == member.ID {
				break
			}
		}

		if move = q.Get("move"); move != "" {
			switch move {
			case "up":
				if i -= 1; i < 0 {
					if i = len(members) - 1; i < 0 {
						i = 0
					}
				}
			case "down":
				if i += 1; i >= len(members) {
					i = 0
				}
			default:
				err = fmt.Errorf("unknown move: %v", q.Get("move"))
				return
			}
			member = members[i]
		} else {
			if addStr := q.Get("add"); addStr != "" {
				if add, err = strconv.Atoi(addStr); err != nil {
					return
				}
			}
		}
	}

	return
}

func writeSplitInstructions(buffer *bytes.Buffer, tgUserID int64, memberName string) {
	buffer.WriteString("Use ⬆ & ⬇ to choose a member.")
	buffer.WriteString("\n\n")
	if tgUserID == 0 {
		buffer.WriteString(fmt.Sprintf("<b>Selected:</b> %v", memberName))
	} else {
		buffer.WriteString(fmt.Sprintf(`<b>Selected:</b> <a href="tg://user?id=%d">%v</a>`, tgUserID, memberName))
	}
}

func writeSplitMembers(buffer *bytes.Buffer, members []models.MemberJson, currentMemberID string, amount models.Amount) {
	var totalShares int
	for _, m := range members {
		totalShares += m.Shares
	}
	if totalShares == 0 {
		totalShares = 1
	}
	for i, m := range members {
		if m.ID == currentMemberID {
			buffer.WriteString(fmt.Sprintf("  <b>%d. %v</b>\n", i+1, html.EscapeString(m.Name)))
		} else {
			buffer.WriteString(fmt.Sprintf("  %d. %v\n", i+1, html.EscapeString(m.Name)))
		}
		mAmount := amount
		mAmount.Value = decimal.Decimal64p2(int64(amount.Value) * int64(m.Shares) / int64(totalShares))
		buffer.WriteString(fmt.Sprintf("     <i>Shares: %d</i> — <code>%v%%</code>", m.Shares, decimal.Decimal64p2(m.Shares*100*100/totalShares)))
		if amount.Value != 0 {
			buffer.WriteString(" = " + mAmount.String())
		}
		buffer.WriteString("\n\n")
	}
}

func addEditSplitInlineKeyboardButtons(kb [][]tgbotapi.InlineKeyboardButton, translator strongo.SingleLocaleTranslator, callbackDataPrefix, backCallbackData string) [][]tgbotapi.InlineKeyboardButton {
	return append(kb, // TODO: Move to Telegram specific package
		[]tgbotapi.InlineKeyboardButton{
			{
				Text:         "-10",
				CallbackData: callbackDataPrefix + "add=-10",
			},
			{
				Text:         "-1",
				CallbackData: callbackDataPrefix + "add=-1",
			},
			{
				Text:         "50/50",
				CallbackData: callbackDataPrefix + "set=50x50",
			},
			{
				Text:         "+1",
				CallbackData: callbackDataPrefix + "add=1",
			},
			{
				Text:         "+10",
				CallbackData: callbackDataPrefix + "add=10",
			},
		},
		[]tgbotapi.InlineKeyboardButton{
			{
				Text:         "⬆️",
				CallbackData: callbackDataPrefix + "move=up",
			},
			{
				Text:         "⬇️",
				CallbackData: callbackDataPrefix + "move=down",
			},
			{
				Text:         translator.Translate("✅ Done"),
				CallbackData: backCallbackData,
			},
		},
	)
}
