package common4debtus

import (
	"bytes"
	"context"
	"github.com/crediterra/money"
	"github.com/sneat-co/sneat-go-core/utm"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/models4debtus"
	"github.com/sneat-co/sneat-translations/trans"
	"github.com/strongo/i18n"
	"regexp"
	"testing"
)

func TestWriteReceiptText(t *testing.T) {
	var (
		buffer bytes.Buffer
	)

	//c, done, err := aetest.NewContext()
	//if err != nil {
	//	t.Fatal(err)
	//}
	//defer done()

	c := context.TODO()

	//logger := &botscore.MockLogger{T: t}

	translator := i18n.NewSingleMapTranslator(i18n.LocaleEnUS, i18n.NewMapTranslator(c, i18n.LocaleCodeEnUK, trans.TRANS))

	transfer := models4debtus.NewTransfer("123", models4debtus.NewTransferData(
		"12",
		false,
		money.Amount{Currency: "EUR", Value: 98765},
		&models4debtus.TransferCounterpartyInfo{
			ContactID:   "23",
			ContactName: "John Whites",
		},
		&models4debtus.TransferCounterpartyInfo{
			UserID:   "12",
			UserName: "Anna Blacks",
		},
	))

	receiptTextBuilder := newReceiptTextBuilder(translator, transfer, ShowReceiptToCounterparty)

	utmParams := utm.Params{
		Source:   "BotIdUnitTest",
		Medium:   "telegram",
		Campaign: "unit-test-campaign",
	}

	_ = receiptTextBuilder.WriteReceiptText(context.Background(), &buffer, utmParams)

	re := regexp.MustCompile(`Anna Blacks borrowed from you <b>987.65 EUR</b>.`)
	if matched := re.MatchString(buffer.String()); !matched {
		t.Errorf("Unexpected output:\nOutput:\n%v\nRegex:\n%v", buffer.String(), re.String())
	}
}
