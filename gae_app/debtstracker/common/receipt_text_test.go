package common

import (
	"bytes"
	"github.com/crediterra/money"
	"regexp"
	"testing"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/bots-go-framework/bots-fw-telegram"
	"github.com/strongo/app"
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

	//logger := &bots.MockLogger{T: t}

	translator := strongo.NewSingleMapTranslator(strongo.LocaleEnUS, strongo.NewMapTranslator(c, trans.TRANS))
	ec := strongo.NewExecutionContext(c, translator)

	transfer := models.Transfer{
		IntegerID: db.IntegerID{ID: 123},
		TransferEntity: models.NewTransferData(
			12,
			false,
			money.Amount{Currency: "EUR", Value: 98765},
			&models.TransferCounterpartyInfo{
				ContactID:   23,
				ContactName: "John Whites",
			},
			&models.TransferCounterpartyInfo{
				UserID:   12,
				UserName: "Anna Blacks",
			},
		)}

	receiptTextBuilder := newReceiptTextBuilder(ec, transfer, ShowReceiptToCounterparty)

	utmParams := UtmParams{
		Source:   "BotIdUnitTest",
		Medium:   telegram.PlatformID,
		Campaign: "unit-test-campaign",
	}

	receiptTextBuilder.WriteReceiptText(&buffer, utmParams)

	re := regexp.MustCompile(`Anna Blacks borrowed from you <b>987.65 EUR</b>.`)
	if matched := re.MatchString(buffer.String()); !matched {
		t.Errorf("Unexpected output:\nOutput:\n%v\nRegex:\n%v", buffer.String(), re.String())
	}
}
