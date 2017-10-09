package api

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"encoding/json"
	"github.com/strongo/decimal"
	"golang.org/x/net/context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
)

const (
	mockBillID    = 123
	creatorUserID = 234
)

func TestBillApiCreateBill(t *testing.T) {

	c := context.Background()
	facade.SetupMocks(c)


	if contact, err := dal.Contact.InsertContact(c, creatorUserID, 0, 0, models.ContactDetails{
		FirstName: "First",
	}, models.Balanced{}); err != nil {
		t.Fatal(err)
	} else if contact.ID != 1 {
		t.Fatalf("contact.ID: %v", contact.ID)
	}
	if contact, err := dal.Contact.InsertContact(c, creatorUserID, 0, 0, models.ContactDetails{
		FirstName: "Second",
	}, models.Balanced{}); err != nil {
		t.Fatal(err)
	} else if contact.ID != 2 {
		t.Fatalf("contact.ID != 2: %v", contact.ID)
	}
	if contact, err := dal.Contact.InsertContact(c, creatorUserID, 0, 0, models.ContactDetails{
		FirstName: "Third",
	}, models.Balanced{}); err != nil {
		t.Fatal(err)
	} else if contact.ID != 3 {
		t.Fatalf("contact.ID != 3: %v", contact.ID)
	}

	responseRecorder := httptest.NewRecorder()

	var body io.Reader

	body = strings.NewReader("")
	request, err := http.NewRequest("POST", "/api/bill-create", body)
	if err != nil {
		t.Fatal(err)
	}
	handleCreateBill(c, responseRecorder, request, auth.AuthInfo{UserID: mockBillID})

	if responseRecorder.Code != http.StatusBadRequest {
		t.Error("Expected to return http.StatusBadRequest on empty request body")
		return
	}

	form := make(url.Values, 3)
	form.Add("name", "Test bill")
	form.Add("currency", "EUR")
	form.Add("amount", "0.10")
	form.Add("split", "percentage")
	form.Add("members", `
	[
		{"UserID":1,"Percent":34,"Amount":0.04},
		{"ContactID":2,"Percent":33,"Amount":0.03},
		{"ContactID":3,"Percent":33,"Amount":0.03}
	]`)

	//body = strings.NewReader("name=Test+bill&currency=EUR&amount=1.23")
	responseRecorder = httptest.NewRecorder()
	request = &http.Request{Method: "POST", URL: &url.URL{Path: "/api/bill-create"}, PostForm: form}
	handleCreateBill(c, responseRecorder, request, auth.AuthInfo{UserID: creatorUserID})

	if responseRecorder.Code != http.StatusOK {
		t.Error("Expected to get http.StatusOK, got:", responseRecorder.Code, responseRecorder.Body.String(), form)
		return
	}
	responseObject := make(map[string]BillDto, 1)

	if err = json.Unmarshal(responseRecorder.Body.Bytes(), &responseObject); err != nil {
		t.Errorf("Response body is not valid JSON: %v", string(responseRecorder.Body.String()))
		return
	}
	responseBill := responseObject["Bill"]
	if responseBill.ID != 1 {
		t.Errorf("Response Bill.ID field has unexpected value: %v", responseBill.ID)
	}
	if responseBill.Name != "Test bill" {
		t.Error("Response Bill.ContactName field has unexpected value:", responseBill.Name)
	}
	if responseBill.Amount.Currency != "EUR" {
		t.Error("Response Bill.AmountTotal.Currency field has unexpected value:", responseBill.Amount.Currency)
	}
	if responseBill.Amount.Value != decimal.NewDecimal64p2FromFloat64(0.10) {
		t.Error("Response Bill.AmountTotal.Value field has unexpected value:", responseBill.Amount.Value)
	}
}
