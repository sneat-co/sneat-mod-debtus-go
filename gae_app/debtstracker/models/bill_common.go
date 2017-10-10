package models

import (
	"fmt"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/strongo/app/gaedb"
	"github.com/strongo/decimal"
	"google.golang.org/appengine/datastore"
	"time"
	"github.com/pkg/errors"
)

type SplitMode string

const (
	SplitModeAdjustment  SplitMode = "adjustment"
	SplitModeEqually     SplitMode = "equally"
	SplitModeExactAmount SplitMode = "exact-amount"
	SplitModePercentage  SplitMode = "percentage"
	SplitModeShare       SplitMode = "shares"
)

type PayMode string

const (
	PayModePrepay  = "prepay"
	PayModeBillpay = "billpay"
)

type BillCommon struct {
	PayMode            PayMode
	CreatorUserID      int64               `datastore:",noindex"`
	userGroupID        string              `datastore:"UserGroupID,noindex"`
	TgInlineMessageIDs []string            `datastore:",noindex"`
	SplitMode          SplitMode           `datastore:",noindex"`
	Status             string
	DtCreated          time.Time
	Name               string              `datastore:",noindex"`
	AmountTotal        decimal.Decimal64p2 `datastore:"AmountTotal"`
	Currency           string              `datastore:",noindex"`
	UserIDs            []int64
	members            []BillMemberJson
	MembersJson        string              `datastore:",noindex"`
	MembersCount       int                 `datastore:",noindex"`
	ContactIDs         []int64 // Holds contact IDs so we can update names in MembersJson on contact changed
	Shares             int                 `datastore:",noindex"`
}

func (entity BillCommon) UserGroupID() string {
	return entity.userGroupID
}

var ErrBillAlreadyAssignedToAnotherGroup = errors.New("bill already assigned to another group ")

func (entity *BillCommon) AssignToGroup(groupID string) (err error) {
	if entity.userGroupID == "" {
		entity.userGroupID = groupID
	} else if entity.userGroupID != groupID {
		err = errors.WithMessage(ErrBillAlreadyAssignedToAnotherGroup, entity.userGroupID)
	}
	return
}

func (entity *BillCommon) AddOrGetMember(userID, contactID int64, name string) (isNew, changed bool, index int, member BillMemberJson, billMembers []BillMemberJson) {
	members := entity.GetMembers()
	billMembers = entity.GetBillMembers()
	var m MemberJson
	if index, m, isNew, changed = addOrGetMember(members, userID, contactID, name); isNew {
		member = BillMemberJson{
			MemberJson: m,
		}
		billMembers = append(billMembers, member)
		if index != len(billMembers)-1 {
			panic("index != len(billMembers) - 1")
		}
		changed = true
	} else /* existing member */ if member = billMembers[index]; member.ID != m.ID {
		panic("member.ID != m.ID")
	}
	if member.ID == "" {
		panic("member.ID is empty string")
	}
	return
}

func (entity *BillCommon) IsOkToSplit() bool {
	if entity.MembersCount <= 1 {
		return false
	}

	var paidByMembers decimal.Decimal64p2
	for _, m := range entity.GetBillMembers() {
		paidByMembers += m.Paid
		//owedByMembers += m.Owes
	}
	return paidByMembers == entity.AmountTotal
}

func (entity *BillCommon) TotalAmount() Amount {
	return NewAmount(Currency(entity.Currency), entity.AmountTotal)
}

func (entity *BillCommon) GetBillMembers() (members []BillMemberJson) {
	if entity.members != nil {
		members = make([]BillMemberJson, len(entity.members))
		copy(members, entity.members)
		return entity.members
	}
	if entity.MembersJson != "" {
		if err := ffjson.Unmarshal([]byte(entity.MembersJson), &entity.members); err != nil {
			panic(err)
		}
		members = make([]BillMemberJson, len(entity.members))
		copy(members, entity.members)
	}
	return members
}

func (entity *BillCommon) GetMembers() (members []MemberJson) {
	billMembers := entity.GetBillMembers()
	members = make([]MemberJson, len(billMembers))
	for i, bm := range billMembers {
		members[i] = bm.MemberJson
	}
	return
}

func (entity *BillCommon) SetBillMembers(members []BillMemberJson) (err error) {
	// Verify members on duplicate user IDs
	{
		isEquallySplit := true

		uniqueUserIDs := make(map[int64]int, len(members))
		for i, member := range members {
			if member.ID == "" {
				return fmt.Errorf("members[%d].ID is empty string, Name: %v", i, member.Name)
			}
			if isEquallySplit {
				if member.Adjustment != 0 || (i > 0 && member.Shares != members[i-1].Shares) {
					isEquallySplit = false
				}
			}
			if member.UserID != 0 {
				for _, uniqueUserID := range uniqueUserIDs {
					if i0, ok := uniqueUserIDs[member.UserID]; ok {
						return fmt.Errorf("duplicate members with same UserID=%d: members[%d].UserID == members[%d].UserID", uniqueUserID, i, i0)
					}
				}
				uniqueUserIDs[member.UserID] = i
			}
			if member.Name == "" && len(member.ContactByUser) == 0 {
				err = fmt.Errorf("no name for the members[%d]", i)
				return
			}
			if member.Owes > entity.AmountTotal {
				err = fmt.Errorf("members[%d].Owes > entity.AmountTotal", i)
				return
			}
			if member.Adjustment > entity.AmountTotal || (member.Adjustment < 0 && -1*member.Adjustment > entity.AmountTotal) {
				err = fmt.Errorf("members[%d].Adjustment is too big", i)
				return
			}
		}

		if isEquallySplit {
			entity.SplitMode = SplitModeEqually
		} else if entity.SplitMode == SplitModeEqually {
			entity.SplitMode = SplitModeShare
		}
	}

	if json, err := ffjson.Marshal(members); err != nil {
		return err
	} else {
		entity.members = make([]BillMemberJson, len(members))
		if copied := copy(entity.members, members); copied != len(members) {
			panic("copied != len(members)")
		}
		entity.MembersCount = len(members)
		entity.MembersJson = string(json)
	}
	for _, m := range members {
		if m.UserID != 0 {
			for _, userID := range entity.UserIDs {
				if userID == m.UserID {
					goto userIDfound
				}
			}
			entity.UserIDs = append(entity.UserIDs, m.UserID)
		userIDfound:
		}
	}
	return nil
}

func (entity *BillCommon) load(ps []datastore.Property) error {
	for _, p := range ps {
		switch p.Name {
		case "UserGroupID":
			entity.userGroupID = p.Value.(string)
			break
		}
	}
	return nil
}

func (entity *BillCommon) save(properties []datastore.Property) (filtered []datastore.Property, err error) {
	if entity.CreatorUserID == 0 {
		panic("entity.CreatorUserID == 0")
	}
	if entity.SplitMode == "" {
		panic("entity.SplitMode is empty string")
	}
	if entity.Status == "" {
		panic("entity.Status is empty string")
	}
	if entity.DtCreated.IsZero() {
		panic("entity.DtCreated is zero")
	}
	if filtered, err = gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
		"MembersCount": gaedb.IsZeroInt,
		"MembersJson":  gaedb.IsEmptyJson,
		"PayMode":      gaedb.IsEmptyString,
		"ContactName":  gaedb.IsEmptyString,
		"SplitMode":    gaedb.IsEmptyString,
		"Shares":       gaedb.IsZeroInt,
	}); err != nil {
		return
	}
	if entity.userGroupID != "" {
		filtered = append(filtered, datastore.Property{Name: "UserGroupID", Value: entity.userGroupID, NoIndex: true})
	}
	return
}
