package models

//go:generate ffjson $GOFILE

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/pquerna/ffjson/ffjson"
	"strconv"
	"strings"
)

type TransferCounterpartyInfo struct {
	UserID             int64  `json:",omitempty"`
	UserName           string `json:",omitempty"`
	ContactID          int64  `json:",omitempty"`
	ContactName        string `json:",omitempty"`
	Note               string `json:",omitempty"`
	Comment            string `json:",omitempty"`
	ReminderID         int64  `json:",omitempty"` // TODO: Consider deletion as prone to errors if not updated on re-schedule, or find and document the reason we have it
	TgBotID            string `json:",omitempty"`
	TgChatID           int64  `json:",omitempty"`
	TgReceiptByTgMsgID int64  `json:",omitempty"`
}

func NewFrom(userID int64, comment string) *TransferCounterpartyInfo {
	return &TransferCounterpartyInfo{UserID: userID, Comment: comment}
}

func NewTo(counterpartyID int64) *TransferCounterpartyInfo {
	return &TransferCounterpartyInfo{ContactID: counterpartyID}
}

func (t TransferCounterpartyInfo) String() string {
	if s, err := ffjson.MarshalFast(&t); err != nil {
		panic(err)
	} else {
		return string(s)
	}
}

func fixContactName(contactName string) (isFixed bool, s string) {
	if start := strings.Index(contactName, "("); start > 0 {
		if end := strings.Index(contactName, ")"); end > start {
			if l := len(contactName); end == l-1 {
				if (end-start-1)*2 == len(contactName)-3 {
					if s = contactName[start+1: end]; s == contactName[:start-1] {
						isFixed = true
						return
					}
				}
			}
		}
	}
	s = contactName
	return
}

func (c TransferCounterpartyInfo) Name() string {
	if c.ContactName != "" {
		if isFixed, s := fixContactName(c.ContactName); isFixed {
			return s
		}
		return c.ContactName
	} else if c.UserName != "" {
		return c.UserName
	} else {
		var n bytes.Buffer
		if c.UserID != 0 {
			n.WriteString("UserID=" + strconv.FormatInt(c.UserID, 10))
		}
		if c.ContactID != 0 {
			if n.Len() > 0 {
				n.WriteString("&")
			}
			n.WriteString("ContactID=" + strconv.FormatInt(c.ContactID, 10))
		}
		return n.String()
	}
}

func (t *TransferEntity) From() *TransferCounterpartyInfo {
	if t.from == nil {
		t.from = &TransferCounterpartyInfo{}

		if t.C_From != "" {
			if err := ffjson.UnmarshalFast([]byte(t.C_From), t.from); err != nil {
				panic(err.Error())
			}
		} else { // TODO: Migration code to be deleted
			from := t.from
			switch t.DirectionObsoleteProp {
			case TransferDirectionUser2Counterparty:
				if from.UserID == 0 {
					from.UserID = t.CreatorUserID
				} else if from.UserID != t.CreatorUserID {
					panic(fmt.Sprintf("from.UserID:%d != t.CreatorUserID:%d", from.UserID, t.CreatorUserID))
				}
				if from.ContactID == 0 {
					from.ContactID = t.CounterpartyCounterpartyID
				} else if from.ContactID != t.CounterpartyCounterpartyID {
					panic(fmt.Sprintf("from.ContactID != t.CounterpartyCounterpartyID: %v, %v", from.ContactID, t.CounterpartyCounterpartyID))
				}
				if from.ContactName == "" {
					from.ContactName = t.CounterpartyCounterpartyName
				} else if from.ContactName != t.CounterpartyCounterpartyName {
					panic(fmt.Sprintf("from.ContactName != t.CounterpartyCounterpartyName: %v, %v", from.ContactName, t.CounterpartyCounterpartyName))
				}
				if from.Comment == "" {
					from.Comment = t.CreatorComment
				} else if from.Comment != t.CreatorComment {
					panic(fmt.Sprintf("from.Comment != t.CreatorComment: %v, %v", from.Comment, t.CreatorComment))
				}
			case TransferDirectionCounterparty2User:
				if from.UserID == 0 {
					from.UserID = t.CounterpartyUserID
				} else if from.UserID != t.CounterpartyUserID {
					panic(fmt.Sprintf("from.UserID:%d != t.CounterpartyUserID:%d", from.UserID, t.CounterpartyUserID))
				}

				if from.ContactID == 0 {
					from.ContactID = t.CreatorCounterpartyID
				} else if from.ContactID != t.CounterpartyCounterpartyID {
					panic(fmt.Sprintf("from.ContactID != t.CreatorCounterpartyID: %v, %v", from.ContactID, t.CreatorCounterpartyID))
				}
				if from.ContactName == "" {
					from.ContactName = t.CreatorCounterpartyName
				} else if from.ContactName != t.CreatorCounterpartyName {
					panic(fmt.Sprintf("from.ContactName != t.CreatorCounterpartyName: %v, %v", from.ContactName, t.CreatorCounterpartyName))
				}
				if from.Comment == "" {
					from.Comment = t.CounterpartyComment
				} else if from.Comment != t.CounterpartyComment {
					panic(fmt.Sprintf("from.Comment != t.CounterpartyComment: %v, %v", from.Comment, t.CounterpartyComment))
				}
			default:
				if t.DirectionObsoleteProp == "" {
					panic("Cant migrate to new From/To props as DirectionObsoleteProp is empty")
				} else {
					panic("Unknown DirectionObsoleteProp: " + t.DirectionObsoleteProp)
				}
			}
		}
	}
	return t.from
}

func (t *TransferEntity) To() *TransferCounterpartyInfo {
	if t.to == nil {
		t.to = &TransferCounterpartyInfo{}
		if t.C_To != "" {
			if err := ffjson.UnmarshalFast([]byte(t.C_To), t.to); err != nil {
				panic(err.Error())
			}
		} else { // TODO: Migration code to be deleted
			to := t.to
			switch t.DirectionObsoleteProp {
			case TransferDirectionUser2Counterparty:
				if to.UserID == 0 {
					to.UserID = t.CounterpartyUserID
				} else if to.UserID != t.CounterpartyUserID {
					panic(fmt.Sprintf("to.UserID:%d != t.CounterpartyUserID:%d", to.UserID, t.CounterpartyUserID))
				}
				if to.ContactID == 0 {
					to.ContactID = t.CreatorCounterpartyID
				} else if to.ContactID != t.CounterpartyCounterpartyID {
					panic(fmt.Sprintf("to.ContactID != t.CreatorCounterpartyID: %v, %v", to.ContactID, t.CreatorCounterpartyID))
				}
				if to.ContactName == "" {
					to.ContactName = t.CreatorCounterpartyName
				} else if to.ContactName != t.CreatorCounterpartyName {
					panic(fmt.Sprintf("to.ContactName != t.CreatorCounterpartyName: %v, %v", to.ContactName, t.CreatorCounterpartyName))
				}
				if to.Comment == "" {
					to.Comment = t.CounterpartyComment
				} else if to.Comment != t.CounterpartyComment {
					panic(fmt.Sprintf("to.Comment != t.CounterpartyComment: %v, %v", to.Comment, t.CounterpartyComment))
				}
			case TransferDirectionCounterparty2User:
				if to.UserID == 0 {
					to.UserID = t.CreatorUserID
				} else if to.UserID != t.CreatorUserID {
					panic(fmt.Sprintf("to.UserID:%d != t.CreatorUserID:%d", to.UserID, t.CreatorUserID))
				}
				if to.ContactID == 0 {
					to.ContactID = t.CounterpartyCounterpartyID
				} else if to.ContactID != t.CounterpartyCounterpartyID {
					panic(fmt.Sprintf("to.ContactID != t.CounterpartyCounterpartyID: %v, %v", to.ContactID, t.CounterpartyCounterpartyID))
				}
				if to.ContactName == "" {
					to.ContactName = t.CounterpartyCounterpartyName
				} else if to.ContactName != t.CounterpartyCounterpartyName {
					panic(fmt.Sprintf("to.ContactName != t.CounterpartyCounterpartyName: %v, %v", to.ContactName, t.CounterpartyCounterpartyName))
				}
				if to.Comment == "" {
					to.Comment = t.CreatorComment
				} else if to.Comment != t.CreatorComment {
					panic(fmt.Sprintf("to.Comment != t.CreatorComment: %v, %v", to.Comment, t.CreatorComment))
				}
			default:
				panic(fmt.Sprintf("Unknown direction: %v", t.Direction()))
			}
		}
	}
	return t.to
}

func (t *TransferEntity) onSaveSerializeJson() error {
	if t.from != nil {
		if s, err := ffjson.MarshalFast(t.from); err != nil {
			panic(errors.WithMessage(err, "Failed to marshal transfer.from"))
		} else {
			t.C_From = string(s)
		}
	} else if t.C_From == "" && t.DirectionObsoleteProp == "" {
		return errors.New("Transfer should have 'From' counterparty")
	}
	if t.to != nil {
		if s, err := ffjson.MarshalFast(t.to); err != nil {
			return errors.WithMessage(err, "Failed to marshal transfer.to")
		} else {
			t.C_To = string(s)
		}
	} else if t.C_To == "" && t.DirectionObsoleteProp == "" {
		return errors.New("Transfer should have 'To' counterparty")
	}
	return nil
}

//func (t *TransferEntity) onSaveMigrateUserProps() {
//	switch t.Direction() {
//	case TransferDirectionUser2Counterparty:
//		from, to := t.From(), t.To()
//		if from.UserID == 0 {
//			from.UserID = t.CreatorUserID
//		}
//		if t.CounterpartyCounterpartyID != 0 && from.ContactID == 0 {
//			from.ContactID = t.CounterpartyCounterpartyID
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
//		to.ContactID = t.CounterpartyCounterpartyID
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
