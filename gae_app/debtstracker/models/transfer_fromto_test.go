package models

import (
	"testing"
)

func TestOnSaveSerializeJson(t *testing.T) {
	transferEntity := TransferEntity{
		from: &TransferCounterpartyInfo{
			UserID: 11,
		},
		to: &TransferCounterpartyInfo{
			ContactID: 22,
		},
	}

	transferEntity.onSaveSerializeJson()

	if transferEntity.C_From == "" {
		t.Error("transferEntity.C_From is empty")
	}
	if transferEntity.C_To == "" {
		t.Error("transferEntity.C_To is empty")
	}
}

func TestTransferFromToUpdate(t *testing.T) {
	transferEntity := TransferEntity{
		CreatorUserID: 11,
		from: &TransferCounterpartyInfo{
			UserID: 11,
		},
		to: &TransferCounterpartyInfo{
			ContactID: 22,
		},
	}

	from := transferEntity.From()
	if v := from.UserID; v != 11 {
		t.Errorf("from.UserID != 11: %d", v)
		return
	}

	to := transferEntity.To()
	if v := to.ContactID; v != 22 {
		t.Errorf("to.ContactID != 22: %d", v)
		return
	}

	from.ContactID = 33
	if v := transferEntity.From().ContactID; v != 33 {
		t.Errorf("transferEntity.From().ContactID != 33: %d", v)
		return
	}

	to.UserID = 44
	if v := transferEntity.To().UserID; v != 44 {
		t.Errorf("transferEntity.To().UserID != 44: %d", v)
		return
	}

	transfer := Transfer{ID: 55, TransferEntity: &transferEntity}

	from = transfer.From()

	from.ContactID = 77
	if v := transfer.From().ContactID; v != 77 {
		t.Errorf("transferEntity.From().ContactID != 77: %d", v)
		return
	}

	creator := transfer.Creator()
	creator.ContactID = 88
	if v := transfer.Creator().ContactID; v != 88 {
		t.Errorf("transfer.Creator().ContactID != 88: %d", v)
	}
	if v := transfer.From().ContactID; v != 88 {
		t.Errorf("transfer.From().ContactID != 88: %d", v)
	}
}

func TestTransferCounterpartyInfo_Name(t *testing.T) {
	var contact TransferCounterpartyInfo
	if contact.ContactName = "Alex (Alex)"; contact.Name() != "Alex" {
		t.Errorf("Exected contact.ContactName() == 'Alex', got: %v", contact.Name())
	}
	if contact.ContactName = "Alex1 (Alex2)"; contact.Name() != "Alex1 (Alex2)" {
		t.Errorf("Exected contact.ContactName() == 'Alex1 (Alex2)', got: %v", contact.Name())
	}
	if contact.ContactName = "John Smith (John Smith)"; contact.Name() != "John Smith" {
		t.Errorf("Exected contact.ContactName() == 'John Smith', got: %v", contact.Name())
	}
}

func TestFixContactName(t *testing.T) {
	if isFixed, _ := fixContactName(""); isFixed {
		t.Error("Should not fix empty string")
	}
	if _, s := fixContactName(""); s != "" {
		t.Errorf("Expected empty string, got: %v", s)
	}
	if isFixed, _ := fixContactName("Alex (Alex)"); !isFixed {
		t.Error("Exected 'Alex (Alex)' to be fixed")
	}
	if _, s := fixContactName("Alex (Alex)"); s != "Alex" {
		t.Errorf("Exected contact.ContactName() == 'Alex', got: %v", s)
	}
	if isFixed, s := fixContactName("Alex1 (Alex2)"); isFixed || s != "Alex1 (Alex2)" {
		t.Errorf("Exected isFiexed=false, s='Alex1 (Alex2)'. Got: isFiexed=%v, s=%v", isFixed, s)
	}
}
