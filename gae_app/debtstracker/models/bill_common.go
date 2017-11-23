package models

import (
	"fmt"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/strongo/db/gaedb"
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

var ErrUnknownSplitMode = errors.New("Unknown split mode")

type PayMode string

const (
	PayModePrepay  = "prepay"
	PayModeBillpay = "billpay"
)

type BillCommon struct {
	PayMode            PayMode
	CreatorUserID      string              `datastore:",noindex"`
	userGroupID        string              `datastore:"UserGroupID"`
	TgInlineMessageIDs []string            `datastore:",noindex"`
	SplitMode          SplitMode           `datastore:",noindex"`
	Status             string
	DtCreated          time.Time
	Name               string              `datastore:",noindex"`
	AmountTotal        decimal.Decimal64p2 `datastore:"AmountTotal"`
	Currency           Currency
	UserIDs            []string
	members            []BillMemberJson
	MembersJson        string              `datastore:",noindex"`
	MembersCount       int                 `datastore:",noindex"`
	ContactIDs         []string // Holds contact IDs so we can update names in MembersJson on contact changed
	Shares             int                 `datastore:",noindex"`
}

func (entity BillCommon) UserGroupID() string {
	return entity.userGroupID
}

var (
	ErrBillAlreadyAssignedToAnotherGroup = errors.New("bill already assigned to another group ")
)

func (entity *BillCommon) AssignToGroup(groupID string) (err error) {
	if groupID == "" {
		err = errors.New("*BillCommon.AssignToGroup(): parameter groupID is required")
		return
	}
	if entity.userGroupID == "" {
		entity.userGroupID = groupID
	} else if entity.userGroupID != groupID {
		err = errors.WithMessage(ErrBillAlreadyAssignedToAnotherGroup, entity.userGroupID)
	}
	return
}

func (entity *BillCommon) AddOrGetMember(groupMemberID, userID, contactID, name string) (isNew, changed bool, index int, member BillMemberJson, billMembers []BillMemberJson) {
	members := entity.GetMembers()
	billMembers = entity.GetBillMembers()
	var m MemberJson
	if index, m, isNew, changed = addOrGetMember(members, groupMemberID, userID, contactID, name); isNew {
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
	if entity.members == nil {
		if entity.MembersJson == "" {
			if entity.MembersCount != 0 {
				panic("entity.MembersJson is empty string && entity.MembersCount != 0")
			}
			return []BillMemberJson{}
		} else if err := ffjson.Unmarshal([]byte(entity.MembersJson), &entity.members); err != nil {
			panic(err)
		}
	}
	if len(entity.members) != entity.MembersCount {
		panic("len(entity.members) != entity.MembersCount")
	}
	// copy to make sure we don't expose cache
	return append(make([]BillMemberJson, 0, len(entity.members)), entity.members...)
}

func (entity *BillCommon) GetMembers() (members []MemberJson) {
	billMembers := entity.GetBillMembers()
	members = make([]MemberJson, len(billMembers))
	for i, bm := range billMembers {
		members[i] = bm.MemberJson
	}
	return
}

func (entity *BillCommon) validateMembersForDuplicatesAndBasicChecks(members []BillMemberJson) (error) {
	isEquallySplit := true
	//maxShares := 0

	uniqueUserIDs := make(map[string]int, len(members))
	for i, member := range members {
		if member.ID == "" {
			return fmt.Errorf("members[%d].ID is empty string, Name: %v", i, member.Name)
		}
		if isEquallySplit {
			//if member.Shares > maxShares {
			//	maxShares = member.Shares
			//}
			if member.Adjustment != 0 || (i > 0 && member.Shares != members[i-1].Shares) {
				isEquallySplit = false
			}
		}
		if member.UserID != "" {
			for _, uniqueUserID := range uniqueUserIDs {
				if i0, ok := uniqueUserIDs[member.UserID]; ok {
					return fmt.Errorf("duplicate members with same UserID=%d: members[%d].UserID == members[%d].UserID", uniqueUserID, i, i0)
				}
			}
			uniqueUserIDs[member.UserID] = i
		}
		if member.Name == "" {
			return fmt.Errorf("no name for the members[%d]", i)
		}
		if member.Owes > entity.AmountTotal {
			return fmt.Errorf("members[%d].Owes > entity.AmountTotal", i)
		}
		if member.Adjustment > entity.AmountTotal || (member.Adjustment < 0 && -1*member.Adjustment > entity.AmountTotal) {
			return fmt.Errorf("members[%d].Adjustment is too big", i)
		}
	}

	if isEquallySplit {
		entity.SplitMode = SplitModeEqually
	} else if entity.SplitMode == SplitModeEqually {
		entity.SplitMode = SplitModeShare
	}
	return nil
}

func (entity *BillCommon) marshalMembersToJsonAndSetMembersCount(members []BillMemberJson) (error) {
	if json, err := ffjson.Marshal(members); err != nil {
		return err
	} else {
		entity.MembersCount = len(members)
		entity.members = append(make([]BillMemberJson, 0, entity.MembersCount), members...)
		entity.validateMembersForDuplicatesAndBasicChecks(entity.members)
		entity.MembersJson = string(json)
	}
	return nil
}

func (entity *BillCommon) setUserIDs(members []BillMemberJson) {
	for _, m := range members {
		if m.UserID != "" {
			for _, userID := range entity.UserIDs {
				if userID == m.UserID {
					goto userIdFound
				}
			}
			entity.UserIDs = append(entity.UserIDs, m.UserID)
		userIdFound:
		}
	}
}

func (entity *BillCommon) setBillMembers(members []BillMemberJson) (err error) {
	//if err = entity.updateMemberOwes(members); err != nil {
	//	return
	//}

	if err = entity.validateMembersForDuplicatesAndBasicChecks(members); err != nil {
		return err
	}

	if err := entity.marshalMembersToJsonAndSetMembersCount(members); err != nil {
		return err
	}

	entity.setUserIDs(members)
	return nil
}

func (entity *BillCommon) load(ps []datastore.Property) []datastore.Property {
	for i, p := range ps {
		if p.Name == "UserGroupID" {
			entity.userGroupID = p.Value.(string)
			return append(ps[:i], ps[i+1:]...)
		}
	}
	return ps
}

func (entity *BillCommon) save(properties []datastore.Property) (filtered []datastore.Property, err error) {
	if entity.CreatorUserID == "" {
		panic("entity.CreatorUserID is empty string")
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
		filtered = append(filtered, datastore.Property{Name: "UserGroupID", Value: entity.userGroupID, NoIndex: false})
	}
	return
}
