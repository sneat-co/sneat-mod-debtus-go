package models4debtus

//go:generate ffjson $GOFILE

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/sneat-co/sneat-core-modules/auth/models4auth"
)

type TransferCounterpartyInfo struct {
	SpaceID            string `json:"spaceID,omitempty"`
	UserID             string `json:"userID,omitempty"`
	UserName           string `json:"userName,omitempty"`
	ContactID          string `json:"contactID,omitempty"`
	ContactName        string `json:"contactName,omitempty"`
	Note               string `json:"note,omitempty"`
	Comment            string `json:"comment,omitempty"`
	ReminderID         string `json:"reminderID,omitempty"` // TODO: Consider deletion as prone to errors if not updated on re-schedule, or find and document the reason we have it
	TgBotID            string `json:"tgBotID,omitempty"`
	TgChatID           int64  `json:"tgChatID,omitempty"`           // Needs to be INT64 as it is INT64 in Telegram API
	TgReceiptByTgMsgID int64  `json:"tgReceiptByTgMsgID,omitempty"` // Needs to be INT64 as it is INT64 in Telegram API
}

func NewFrom(userID, spaceID, comment string) *TransferCounterpartyInfo {
	return &TransferCounterpartyInfo{UserID: userID, SpaceID: spaceID, Comment: comment}
}

func NewTo(spaceID, counterpartyID string) *TransferCounterpartyInfo {
	return &TransferCounterpartyInfo{SpaceID: spaceID, ContactID: counterpartyID}
}

func (t TransferCounterpartyInfo) String() string {
	if s, err := ffjson.MarshalFast(&t); err != nil {
		panic(err)
	} else {
		return string(s)
	}
}

func (c TransferCounterpartyInfo) Name() string {
	if c.ContactName != "" {
		if isFixed, s := models4auth.FixContactName(c.ContactName); isFixed {
			return s
		}
		return c.ContactName
	} else if c.UserName != "" {
		return c.UserName
	} else {
		var n bytes.Buffer
		if c.UserID != "" {
			n.WriteString("UserID=" + c.UserID)
		}
		if c.ContactID != "" {
			if n.Len() > 0 {
				n.WriteString("&")
			}
			n.WriteString("ContactID=" + c.ContactID)
		}
		return n.String()
	}
}

func (t *TransferData) From() *TransferCounterpartyInfo {
	if t.from == nil {
		t.from = &TransferCounterpartyInfo{}

		if t.FromJson != "" {
			if err := ffjson.UnmarshalFast([]byte(t.FromJson), t.from); err != nil {
				panic(err.Error())
			}
		} else {
			panic("FromJson is empty")
			// // TODO: Migration code to be deleted
			// from := t.from
			// switch t.DirectionObsoleteProp {
			// case TransferDirectionUser2Counterparty:
			// 	if from.UserID == 0 {
			// 		from.UserID = t.CreatorUserID
			// 	} else if from.UserID != t.CreatorUserID {
			// 		panic(fmt.Sprintf("from.UserID:%d != t.CreatorUserID:%d", from.UserID, t.CreatorUserID))
			// 	}
			// 	if from.ContactID == 0 {
			// 		from.ContactID = t.CounterpartyContactID
			// 	} else if from.ContactID != t.CounterpartyContactID {
			// 		panic(fmt.Sprintf("from.ContactID != t.CounterpartyContactID: %v, %v", from.ContactID, t.CounterpartyContactID))
			// 	}
			// 	if from.ContactName == "" {
			// 		from.ContactName = t.CounterpartyCounterpartyName
			// 	} else if from.ContactName != t.CounterpartyCounterpartyName {
			// 		panic(fmt.Sprintf("from.ContactName != t.CounterpartyCounterpartyName: %v, %v", from.ContactName, t.CounterpartyCounterpartyName))
			// 	}
			// 	if from.Comment == "" {
			// 		from.Comment = t.CreatorComment
			// 	} else if from.Comment != t.CreatorComment {
			// 		panic(fmt.Sprintf("from.Comment != t.CreatorComment: %v, %v", from.Comment, t.CreatorComment))
			// 	}
			// case TransferDirectionCounterparty2User:
			// 	if from.UserID == 0 {
			// 		from.UserID = t.CounterpartyUserID
			// 	} else if from.UserID != t.CounterpartyUserID {
			// 		panic(fmt.Sprintf("from.UserID:%d != t.CounterpartyUserID:%d", from.UserID, t.CounterpartyUserID))
			// 	}
			//
			// 	if from.ContactID == 0 {
			// 		from.ContactID = t.CreatorCounterpartyID
			// 	} else if from.ContactID != t.CounterpartyContactID {
			// 		panic(fmt.Sprintf("from.ContactID != t.CreatorCounterpartyID: %v, %v", from.ContactID, t.CreatorCounterpartyID))
			// 	}
			// 	if from.ContactName == "" {
			// 		from.ContactName = t.CreatorCounterpartyName
			// 	} else if from.ContactName != t.CreatorCounterpartyName {
			// 		panic(fmt.Sprintf("from.ContactName != t.CreatorCounterpartyName: %v, %v", from.ContactName, t.CreatorCounterpartyName))
			// 	}
			// 	if from.Comment == "" {
			// 		from.Comment = t.CounterpartyComment
			// 	} else if from.Comment != t.CounterpartyComment {
			// 		panic(fmt.Sprintf("from.Comment != t.CounterpartyComment: %v, %v", from.Comment, t.CounterpartyComment))
			// 	}
			// default:
			// 	if t.DirectionObsoleteProp == "" {
			// 		panic("Cant migrate to new From/To props as DirectionObsoleteProp is empty")
			// 	} else {
			// 		panic("Unknown DirectionObsoleteProp: " + t.DirectionObsoleteProp)
			// 	}
			// }
		}
	}
	return t.from
}

func (t *TransferData) To() *TransferCounterpartyInfo {
	if t.to == nil {
		t.to = &TransferCounterpartyInfo{}
		if t.ToJson != "" {
			if err := ffjson.UnmarshalFast([]byte(t.ToJson), t.to); err != nil {
				panic(err.Error())
			}
		} else { // TODO: Migration code to be deleted
			panic("ToJson is empty")
			// to := t.to
			// switch t.DirectionObsoleteProp {
			// case TransferDirectionUser2Counterparty:
			// 	if to.UserID == 0 {
			// 		to.UserID = t.CounterpartyUserID
			// 	} else if to.UserID != t.CounterpartyUserID {
			// 		panic(fmt.Sprintf("to.UserID:%d != t.CounterpartyUserID:%d", to.UserID, t.CounterpartyUserID))
			// 	}
			// 	if to.ContactID == 0 {
			// 		to.ContactID = t.CreatorCounterpartyID
			// 	} else if to.ContactID != t.CounterpartyContactID {
			// 		panic(fmt.Sprintf("to.ContactID != t.CreatorCounterpartyID: %v, %v", to.ContactID, t.CreatorCounterpartyID))
			// 	}
			// 	if to.ContactName == "" {
			// 		to.ContactName = t.CreatorCounterpartyName
			// 	} else if to.ContactName != t.CreatorCounterpartyName {
			// 		panic(fmt.Sprintf("to.ContactName != t.CreatorCounterpartyName: %v, %v", to.ContactName, t.CreatorCounterpartyName))
			// 	}
			// 	if to.Comment == "" {
			// 		to.Comment = t.CounterpartyComment
			// 	} else if to.Comment != t.CounterpartyComment {
			// 		panic(fmt.Sprintf("to.Comment != t.CounterpartyComment: %v, %v", to.Comment, t.CounterpartyComment))
			// 	}
			// case TransferDirectionCounterparty2User:
			// 	if to.UserID == 0 {
			// 		to.UserID = t.CreatorUserID
			// 	} else if to.UserID != t.CreatorUserID {
			// 		panic(fmt.Sprintf("to.UserID:%d != t.CreatorUserID:%d", to.UserID, t.CreatorUserID))
			// 	}
			// 	if to.ContactID == 0 {
			// 		to.ContactID = t.CounterpartyContactID
			// 	} else if to.ContactID != t.CounterpartyContactID {
			// 		panic(fmt.Sprintf("to.ContactID != t.CounterpartyContactID: %v, %v", to.ContactID, t.CounterpartyContactID))
			// 	}
			// 	if to.ContactName == "" {
			// 		to.ContactName = t.CounterpartyCounterpartyName
			// 	} else if to.ContactName != t.CounterpartyCounterpartyName {
			// 		panic(fmt.Sprintf("to.ContactName != t.CounterpartyCounterpartyName: %v, %v", to.ContactName, t.CounterpartyCounterpartyName))
			// 	}
			// 	if to.Comment == "" {
			// 		to.Comment = t.CreatorComment
			// 	} else if to.Comment != t.CreatorComment {
			// 		panic(fmt.Sprintf("to.Comment != t.CreatorComment: %v, %v", to.Comment, t.CreatorComment))
			// 	}
			// default:
			// 	panic(fmt.Sprintf("Unknown direction: %v", t.Direction()))
			// }
		}
	}
	return t.to
}

func (t *TransferData) onSaveSerializeJson() error {
	if t.from != nil {
		if s, err := json.Marshal(t.from); err != nil {
			panic(fmt.Errorf("failed to marshal transfer.from: %w", err))
		} else {
			t.FromJson = string(s)
		}
	} else if t.FromJson == "" {
		return errors.New("TransferEntry should have 'From' counterparty")
	}
	if t.to != nil {
		if s, err := json.Marshal(t.to); err != nil {
			return fmt.Errorf("failed to marshal transfer.to: %w", err)
		} else {
			t.ToJson = string(s)
		}
	} else if t.ToJson == "" {
		return errors.New("TransferEntry should have 'To' counterparty")
	}
	return nil
}

//func (t *TransferData) onSaveMigrateUserProps() {
//	switch t.Direction() {
//	case TransferDirectionUser2Counterparty:
//		from, to := t.From(), t.To()
//		if from.UserID == 0 {
//			from.UserID = t.CreatorUserID
//		}
//		if t.CounterpartyContactID != 0 && from.ContactID == 0 {
//			from.ContactID = t.CounterpartyContactID
//		}
//		if from.ContactName == "" && t.CounterpartyCounterpartyName != "" {
//			from.ContactName = t.CounterpartyCounterpartyName
//		}
//
//		from.Comment = t.CreatorComment
//		from.Note = t.CreatorNote
//		to.UserID = t.CounterpartyUserID
//		to.ContactID = t.CreatorCounterpartyID
//		to.ContactName = t.CreatorCounterpartyName
//		to.Comment = t.CounterpartyComment
//		to.Note = t.CounterpartyNote
//	case TransferDirectionCounterparty2User:
//		from, to := t.From(), t.To()
//		to.UserID = t.CreatorUserID
//		to.ContactID = t.CounterpartyContactID
//		to.ContactName = t.CounterpartyCounterpartyName
//		to.Comment = t.CreatorComment
//		to.Note = t.CreatorNote
//		from.UserID = t.CounterpartyUserID
//		from.ContactID = t.CreatorCounterpartyID
//		from.ContactName = t.CreatorCounterpartyName
//		from.Comment = t.CounterpartyComment
//		from.Note = t.CounterpartyNote
//	}
//}
