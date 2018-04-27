package dtb_transfer

import (
	"github.com/DebtsTracker/translations/trans"
	//"fmt"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/app"
	"github.com/strongo/decimal"
)

func getTestMocks(t *testing.T, locale strongo.Locale) BalanceMessageBuilder {
	translator := strongo.NewMapTranslator(context.TODO(), trans.TRANS)
	singleLocaleTranslator := strongo.NewSingleMapTranslator(locale, translator)
	return NewBalanceMessageBuilder(singleLocaleTranslator)
}

func enMock(t *testing.T) BalanceMessageBuilder { return getTestMocks(t, strongo.LocaleEnUS) }
func ruMock(t *testing.T) BalanceMessageBuilder { return getTestMocks(t, strongo.LocaleRuRu) }

var (
	ruLinker = common.NewLinker(strongo.EnvLocal, 123, strongo.LocaleRuRu.Code5, "unit-Test")
	enLinker = common.NewLinker(strongo.EnvLocal, 123, strongo.LocaleEnUS.Code5, "unit-Test")
)

type testBalanceDataProvider struct {
}

func TestBalanceMessageSingleCounterparty(t *testing.T) {
	balanceJson := json.RawMessage(`{"USD": 10}`)
	counterparties := []models.UserContactJson{
		{
			ID:     1,
			Name:   "John Doe",
			Status: "active",
			//UserID: 1,
			BalanceJson: &balanceJson,
		},
	}

	c := context.TODO()
	expectedEn := `<a href="https://debtstracker.local/contact?id=1&lang=en-US&secret=SECRET">John Doe</a>`
	expectedRu := `<a href="https://debtstracker.local/contact?id=1&lang=ru-RU&secret=SECRET">John Doe</a>`

	assert(t, strongo.LocaleEnUS, 0, fmt.Sprintf("%v - owes you 10 USD", expectedEn), enMock(t).ByContact(c, enLinker, counterparties))

	assert(t, strongo.LocaleRuRu, 0, fmt.Sprintf("%v - долг вам 10 USD", expectedRu), ruMock(t).ByContact(c, ruLinker, counterparties))

	balanceJson = json.RawMessage(`{"USD": -10}`)
	counterparties[0].BalanceJson = &balanceJson
	assert(t, strongo.LocaleRuRu, 0, fmt.Sprintf("%v - вы должны 10 USD", expectedRu), ruMock(t).ByContact(c, ruLinker, counterparties))

	balanceJson = json.RawMessage(`{"USD": 10, "EUR": 20}`)
	counterparties[0].BalanceJson = &balanceJson
	assert(t, strongo.LocaleEnUS, 0, fmt.Sprintf("%v - owes you 20 EUR and 10 USD", expectedEn), enMock(t).ByContact(c, enLinker, counterparties))

	balanceJson = json.RawMessage(`{"USD": 10, "EUR": 20, "RUB": 15}`)
	counterparties[0].BalanceJson = &balanceJson
	assert(t, strongo.LocaleEnUS, 0, fmt.Sprintf("%v - owes you 20 EUR, 15 RUB and 10 USD", expectedEn), enMock(t).ByContact(c, enLinker, counterparties))

}

func TestBalanceMessageTwoCounterparties(t *testing.T) {
	john := models.UserContactJson{
		ID:   1,
		Name: "Johnny The Doe",
	}

	jack := models.UserContactJson{
		ID:   2,
		Name: "Jacky Dark Brown",
	}

	c := context.TODO()

	johnLink := fmt.Sprintf(`<a href="https://debtstracker.local/contact?id=1&lang=en-US&secret=SECRET">%v</a>`, john.Name)
	jackLink := fmt.Sprintf(`<a href="https://debtstracker.local/contact?id=2&lang=en-US&secret=SECRET">%v</a>`, jack.Name)

	var johnBalance, jackBalance json.RawMessage
	johnBalance = json.RawMessage(`{"USD": 10}`)
	john.BalanceJson = &johnBalance
	jackBalance = json.RawMessage(`{"USD": 15}`)
	jack.BalanceJson = &jackBalance
	assert(t, strongo.LocaleEnUS, 0, fmt.Sprintf("%v - owes you 10 USD\n%v - owes you 15 USD", johnLink, jackLink), enMock(t).ByContact(c, enLinker, []models.UserContactJson{john, jack}))

	johnBalance = json.RawMessage(`{"USD": 10, "EUR": 20}`)
	john.BalanceJson = &johnBalance
	jackBalance = json.RawMessage(`{"USD": 40, "EUR": 15}`)
	jack.BalanceJson = &jackBalance
	assert(t, strongo.LocaleEnUS, 0, fmt.Sprintf("%v - owes you 20 EUR and 10 USD\n%v - owes you 40 USD and 15 EUR", johnLink, jackLink), enMock(t).ByContact(c, enLinker, []models.UserContactJson{john, jack}))

	johnBalance = json.RawMessage(`{"USD": 10, "EUR": 20, "RUB": 100}`)
	john.BalanceJson = &johnBalance
	jackBalance = json.RawMessage(`{"USD": 40, "EUR": 15}`)
	jack.BalanceJson = &jackBalance
	assert(t, strongo.LocaleEnUS, 0, fmt.Sprintf("%v - owes you 100 RUB, 20 EUR and 10 USD\n%v - owes you 40 USD and 15 EUR", johnLink, jackLink), enMock(t).ByContact(c, enLinker, []models.UserContactJson{john, jack}))

	johnBalance = json.RawMessage(`{"USD": -10}`)
	john.BalanceJson = &johnBalance
	jackBalance = json.RawMessage(`{"USD": -15}`)
	jack.BalanceJson = &jackBalance
	assert(t, strongo.LocaleEnUS, 0, fmt.Sprintf("%v - you owe 10 USD\n%v - you owe 15 USD", johnLink, jackLink), enMock(t).ByContact(c, enLinker, []models.UserContactJson{john, jack}))

	johnBalance = json.RawMessage(`{"USD": -10}`)
	john.BalanceJson = &johnBalance
	jackBalance = json.RawMessage(`{"USD": 15}`)
	jack.BalanceJson = &jackBalance
	assert(t, strongo.LocaleEnUS, 0, fmt.Sprintf("%v - you owe 10 USD\n%v - owes you 15 USD", johnLink, jackLink), enMock(t).ByContact(c, enLinker, []models.UserContactJson{john, jack}))

}

func TestBalanceMessageBuilder_ByCurrency(t *testing.T) {
	balance := models.Balance{
		models.CURRENCY_USD: decimal.NewDecimal64p2(10, 0),
		models.CURRENCY_RUB: decimal.NewDecimal64p2(50, 0),
		models.CURRENCY_EUR: decimal.NewDecimal64p2(15, 0),
	}
	assert(t, strongo.LocaleRuRu, 0, "<b>Всего</b>\nВам должны 50 RUB, 15 EUR и 10 USD", ruMock(t).ByCurrency(true, balance))
}

var reCleanSecret = regexp.MustCompile(`secret.+?"`)

func assert(t *testing.T, locale strongo.Locale, warningsCount int, expected, actual string) {
	actual = reCleanSecret.ReplaceAllString(actual, `secret=SECRET"`)
	if actual != expected {
		t.Errorf("Unexpected output for locale %v:\nExpected:\n%v\nActual:\n%v", locale.Code5, expected, actual)
	}
	//if len(log.Warnings) != warningsCount {
	//	t.Errorf("Unexpected warnings count: %v", log.Warnings)
	//}
}
