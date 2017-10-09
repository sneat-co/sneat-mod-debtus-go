package common

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bytes"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/platforms/telegram"
	"golang.org/x/net/context"
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

	//logger := &bots.MockLogger{T: t}

	translator := strongo.NewSingleMapTranslator(strongo.LocaleEnUS, strongo.NewMapTranslator(c, trans.TRANS))
	ec := strongo.NewExecutionContext(c, translator)

	transfer := models.Transfer{
		ID: 123,
		TransferEntity: models.NewTransferEntity(
			12,
			false,
			models.Amount{Currency: "EUR", Value: 98765},
			&models.TransferCounterpartyInfo{
				ContactID:   23,
				ContactName: "John White",
			},
			&models.TransferCounterpartyInfo{
				UserID:   12,
				UserName: "Anna Black",
			},
		)}

	receiptTextBuilder := newReceiptTextBuilder(ec, transfer, ShowReceiptToCounterparty)

	utmParams := UtmParams{
		Source:   "BotIdUnitTest",
		Medium:   telegram_bot.TelegramPlatformID,
		Campaign: "unit-test-campaign",
	}

	receiptTextBuilder.WriteReceiptText(&buffer, utmParams)

	re := regexp.MustCompile(`Anna Black borrowed from you 987.65 EUR.`)
	if matched := re.MatchString(buffer.String()); !matched {
		t.Errorf("Unexpected output:\nOutput:\n%v\nRegex:\n%v", buffer.String(), re.String())
	}
}
