package dtb_transfer

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app"
	"golang.org/x/net/context"
	"golang.org/x/net/html"
)

type BalanceMessageBuilder struct {
	translator strongo.SingleLocaleTranslator
	NeedsTotal bool
}

func NewBalanceMessageBuilder(translator strongo.SingleLocaleTranslator) BalanceMessageBuilder {
	return BalanceMessageBuilder{translator: translator}
}

func (m BalanceMessageBuilder) ByCounterparty(c context.Context, linker common.Linker, counterparties []models.UserContactJson) string {
	var buffer bytes.Buffer
	translator := m.translator

	writeBalanceRow := func(counterparty models.UserContactJson, b models.Balance, msg string) {
		if len(b) > 0 {
			amounts := b.CommaSeparatedUnsignedWithSymbols(translator)
			msg = m.translator.Translate(msg)
			name := html.EscapeString(counterparty.Name)
			name = fmt.Sprintf(`<a href="%v">%v</a>`, linker.UrlToContact(counterparty.ID), name)
			buffer.WriteString(fmt.Sprintf(msg, name, amounts) + "\n")
		}
	}

	var (
		counterpartiesWithZeroBalance      bytes.Buffer
		counterpartiesWithZeroBalanceCount int
	)

	now := time.Now()

	for _, counterparty := range counterparties {
		counterpartyBalanceWithInterest := counterparty.BalanceWithInterest(c, now)
		//counterpartyBalance := counterparty.Balance()
		//log.Debugf(c, "counterpartyBalanceWithInterest: %v\ncounterpartyBalance: %v", counterpartyBalanceWithInterest, counterpartyBalance)
		if counterpartyBalanceWithInterest.IsZero() {
			counterpartiesWithZeroBalanceCount += 1
			counterpartiesWithZeroBalance.WriteString(strconv.FormatInt(counterparty.ID, 10))
			counterpartiesWithZeroBalance.WriteString(", ")
			continue
		}
		writeBalanceRow(counterparty, counterpartyBalanceWithInterest.OnlyPositive(), trans.MESSAGE_TEXT_BALANCE_SINGLE_CURRENCY_COUNTERPARTY_DEBT_TO_USER)
		writeBalanceRow(counterparty, counterpartyBalanceWithInterest.OnlyNegative(), trans.MESSAGE_TEXT_BALANCE_SINGLE_CURRENCY_COUNTERPARTY_DEBT_BY_USER)
	}
	//if counterpartiesWithZeroBalanceCount > 0 {
	//	log.Debugf(c, "There are %d counterparties with zero balance: %v", counterpartiesWithZeroBalanceCount, strings.TrimRight(counterpartiesWithZeroBalance.String(), ", "))
	//}
	if l := buffer.Len() - 1; l > 0 {
		buffer.Truncate(l)
	}
	return buffer.String()
}

func (m BalanceMessageBuilder) ByCurrency(isTotal bool, balance models.Balance) string {
	var buffer bytes.Buffer
	translator := m.translator
	if isTotal {
		buffer.WriteString("<b>" + translator.Translate(trans.MESSAGE_TEXT_BALANCE_CURRENCY_TOTAL_INTRO) + "</b>\n")
	}
	debtByUser := balance.OnlyNegative()
	debtToUser := balance.OnlyPositive()
	commaSeparatedAmounts := func(prefix string, owed models.Balance) {
		if !owed.IsZero() {
			buffer.WriteString(fmt.Sprintf(translator.Translate(prefix), owed.CommaSeparatedUnsignedWithSymbols(translator)) + "\n")
		}
	}
	commaSeparatedAmounts(trans.MESSAGE_TEXT_BALANCE_CURRENCY_ROW_DEBT_BY_USER, debtByUser)
	commaSeparatedAmounts(trans.MESSAGE_TEXT_BALANCE_CURRENCY_ROW_DEBT_TO_USER, debtToUser)

	if l := buffer.Len() - 1; l > 0 {
		buffer.Truncate(l)
	}
	return buffer.String()
}

func BalanceForCounterpartyWithHeader(counterpartyLink string, b models.Balance, translator strongo.SingleLocaleTranslator) string {
	balanceMessageBuilder := NewBalanceMessageBuilder(translator)
	header := fmt.Sprintf("<b>%v</b>: %v", translator.Translate(trans.MESSAGE_TEXT_BALANCE_HEADER), counterpartyLink)
	return "\n" + header + common.HORIZONTAL_LINE + balanceMessageBuilder.ByCurrency(false, b) + common.HORIZONTAL_LINE
}
