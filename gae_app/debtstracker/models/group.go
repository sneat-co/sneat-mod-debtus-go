package models

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/strongo/app/db"
	"github.com/strongo/app/gaedb"
	"google.golang.org/appengine/datastore"
	"strings"
	"github.com/strongo/decimal"
)

const GroupKind = "Group"

type Group struct {
	db.StringID
	db.NoIntID
	*GroupEntity
}

func (Group) Kind() string {
	return GroupKind
}

func (group Group) Entity() interface{} {
	return group.GroupEntity
}

func (group Group) SetEntity(entity interface{}) {
	if entity == nil {
		group.GroupEntity = nil
	} else {
		group.GroupEntity = entity.(*GroupEntity)
	}
}

var _ db.EntityHolder = (*Group)(nil)

type GroupEntity struct {
	CreatorUserID string
	//IsUser2User         bool   `datastore:",noindex"`
	Name                string   `datastore:",noindex"`
	Note                string   `datastore:",noindex"`
	DefaultCurrency     Currency `datastore:",noindex"`
	members             []GroupMemberJson
	MembersCount        int      `datastore:",noindex"`
	MembersJson         string   `datastore:",noindex"`
	telegramGroups      []GroupTgChatJson
	TelegramGroupsCount int      `datastore:"TgGroupsCount,noindex"`
	TelegramGroupsJson  string   `datastore:"TgGroupsJson,noindex"`
	billsHolder
}

func (entity *GroupEntity) ApplyBillBalanceDifference(currency Currency, diff BillBalanceDifference) (changed bool, err error) {
	if currency == "" {
		panic("currency parameter is required")
	}
	if strings.TrimSpace(string(currency)) != string(currency) {
		panic("currency parameter has leading ot closing spaces: " + currency)
	}

	groupMembers := entity.GetGroupMembers()

	var diffTotal, balanceTotal decimal.Decimal64p2
	diffCopy := make(BillBalanceDifference, len(diff))

	for i := range groupMembers {
		groupMemberID := groupMembers[i].ID

		if memberDifference, ok := diff[groupMemberID]; ok {
			delete(diff, groupMemberID)
			diffCopy[groupMemberID] = memberDifference
			if memberDifference == 0 {
				panic("memberDifference.Paid == 0 && memberDifference.Owes == 0, memberID: " + groupMemberID)
			}
			diffTotal += memberDifference
			if diffAmount := memberDifference; diffAmount != 0 {
				if groupMembers[i].Balance == nil || len(groupMembers) == 0 {
					groupMembers[i].Balance = Balance{currency: diffAmount}
					balanceTotal += diffAmount
				} else {
					groupMembers[i].Balance[currency] += diffAmount
					if len(groupMembers[i].Balance) == 0 {
						groupMembers[i].Balance = nil
					} else {
						balanceTotal += groupMembers[i].Balance[currency]
					}
				}
			}
		}
	}

	if len(diff) > 0 {
		err = errors.WithMessage(ErrNonGroupMember, fmt.Sprintf("%v", diff))
		return
	}

	if diffTotal != 0 {
		err = errors.WithMessage(ErrBillOwesDiffTotalIsNotZero, fmt.Sprintf("diffTotal=%v, diff=%v", diffTotal, diffCopy))
		return
	}

	if balanceTotal != 0 {
		err = errors.WithMessage(GroupTotalBalanceHasNonZeroValue, fmt.Sprintf("balanceTotal=%v, diff=%v", balanceTotal, diffCopy))
		return
	}
	return entity.SetGroupMembers(groupMembers), err
}

func (entity *GroupEntity) GetTelegramGroups() (tgGroups []GroupTgChatJson, err error) {
	if entity.telegramGroups != nil {
		return entity.telegramGroups, nil
	}
	if entity.TelegramGroupsJson != "" {
		if err = ffjson.Unmarshal([]byte(entity.TelegramGroupsJson), &tgGroups); err != nil {
			return
		} else if len(tgGroups) != entity.TelegramGroupsCount {
			err = errors.WithMessage(ErrJsonCountMismatch, "len([]GroupTgChatJson) != entity.TelegramGroupsCount")
			return
		}
		entity.telegramGroups = tgGroups
	}
	return
}

func (entity *GroupEntity) SetTelegramGroups(tgGroups []GroupTgChatJson) (changed bool) {
	if data, err := ffjson.Marshal(tgGroups); err != nil {
		panic(err.Error())
	} else {
		if s := string(data); s != entity.TelegramGroupsJson {
			entity.TelegramGroupsJson = s
			changed = true
		}
		if l := len(tgGroups); l != entity.TelegramGroupsCount {
			entity.TelegramGroupsCount = l
			changed = true
		}
	}
	return
}

func (entity *GroupEntity) AddOrGetMember(userID, contactID, name string) (isNew, changed bool, index int, member GroupMemberJson, groupMembers []GroupMemberJson) {
	members := entity.GetMembers()
	groupMembers = entity.GetGroupMembers()
	var m MemberJson
	if index, m, isNew, changed = addOrGetMember(members, "", userID, contactID, name); isNew {
		member = GroupMemberJson{
			MemberJson: m,
		}
		groupMembers = append(groupMembers, member)
		if index != len(groupMembers)-1 {
			panic("index != len(groupMembers) - 1")
		}
		changed = true
	} else /* existing member */ if member = groupMembers[index]; member.ID != m.ID {
		panic("member.ID != m.ID")
	}
	if member.ID == "" {
		panic("member.ID is empty string")
	}
	return
}

func addOrGetMember(members []MemberJson, memberID, userID, contactID, name string) (index int, member MemberJson, isNew, changed bool) {
	if userID != "" || contactID != "" {
		for i, m := range members {
			if m.ID == memberID || m.UserID == userID {
				member = m
				index = i
				if contactID != "" {
					for _, cID := range m.ContactIDs {
						if cID == contactID {
							goto contactFound
						}
					}
					m.ContactIDs = append(m.ContactIDs, contactID)
					changed = true
				contactFound:
				}
				member = m
				index = i
				return
			} else if contactID != "" {
				for _, cID := range m.ContactIDs {
					if cID == contactID {
						member = m
						index = i
						return
					}
				}
			}
		}
	}
	member = MemberJson{
		ID:     memberID,
		Name:   name,
		UserID: userID,
	}
	if member.ID == "" {
		for j := 0; j < 100; j++ {
			member.ID = db.RandomStringID(7)
			for _, m := range members {
				if m.ID == member.ID {
					goto duplicate
				}
			}
			break
		duplicate:
		}
		if member.ID == "" {
			panic("Failed to generate random member ID")
		}
	}
	return len(members), member, true, true
}

func (entity *GroupEntity) GetGroupMembers() []GroupMemberJson {
	members := make([]GroupMemberJson, entity.MembersCount)
	if entity.members != nil && len(entity.members) == entity.MembersCount {
		copy(members, entity.members)
		return members
	}
	if entity.MembersJson != "" {
		if err := ffjson.Unmarshal(([]byte)(entity.MembersJson), &members); err != nil {
			panic(err.Error())
		}
	}
	if len(members) != entity.MembersCount {
		panic("len(members) != entity.MembersCount")
	}
	entity.members = make([]GroupMemberJson, entity.MembersCount, entity.MembersCount)
	copy(entity.members, members)
	return members
}

func (entity *GroupEntity) GetGroupMemberByID(id string) (GroupMemberJson, error) {
	if id == "" {
		return GroupMemberJson{}, errors.WithMessage(db.ErrRecordNotFound, "empty id")
	}
	for _, m := range entity.GetGroupMembers() {
		if m.ID == id {
			return m, nil
		}
	}
	return GroupMemberJson{}, errors.WithMessage(db.ErrRecordNotFound, "unknown id="+id)
}

func (entity *GroupEntity) GetMembers() (members []MemberJson) {
	groupMembers := entity.GetGroupMembers()
	members = make([]MemberJson, len(groupMembers))
	for i, gm := range groupMembers {
		members[i] = gm.MemberJson
	}
	return
}

func (entity *GroupEntity) GetSplitMode() SplitMode {
	if entity.MembersCount == 0 {
		return SplitModeEqually
	}
	var min, max int
	for _, m := range entity.GetGroupMembers() {
		if m.Shares < min || min == 0 {
			min = m.Shares
		}
		if m.Shares > max {
			max = m.Shares
		}
	}
	if min == max {
		return SplitModeEqually
	}
	return SplitModeShare
}

func (entity *GroupEntity) TotalShares() (n int) {
	for _, m := range entity.GetGroupMembers() {
		n += m.Shares
	}
	return
}

func (entity *GroupEntity) UserIsMember(userID string) bool {
	for _, m := range entity.GetGroupMembers() {
		if m.UserID == userID {
			return true
		}
	}
	return false
}

func (entity *GroupEntity) SetGroupMembers(members []GroupMemberJson) (changed bool) {
	if len(members) == 0 {
		if changed = entity.MembersJson != ""; changed {
			entity.members = make([]GroupMemberJson, 0)
			entity.MembersJson = ""
			entity.MembersCount = 0
		}
		return
	}
	if err := entity.validateMembers(members, len(members)); err != nil {
		panic(err)
	}
	if data, err := ffjson.Marshal(members); err != nil {
		ffjson.Pool(data)
		panic(err)
	} else if membersJson := (string)(data); membersJson != entity.MembersJson {
		ffjson.Pool(data)
		if membersJson == "[]" {
			if entity.MembersJson == "" {
				return
			}
			membersJson = ""
		}
		changed = true
		entity.MembersJson = membersJson
		entity.members = make([]GroupMemberJson, len(members))
		copy(entity.members, members)
		entity.MembersCount = len(members)
	} else {
		ffjson.Pool(data)
	}
	return
}

func (entity *GroupEntity) validateMembers(members []GroupMemberJson, membersCount int) error {
	if membersCount != len(members) {
		return errors.New(fmt.Sprintf("entity.MembersCount != len(members), %d != %d", entity.MembersCount, len(members)))
	}

	type Empty struct {
	}

	EMPTY := Empty{}

	totalBalance := make(Balance)

	userIDs := make(map[string]Empty, entity.MembersCount)
	contactIDs := make(map[string]Empty, entity.MembersCount)

	memberIDs := make(map[string]Empty, entity.MembersCount)

	for i, m := range members {
		if m.ID == "" {
			return fmt.Errorf("members[%d].ID is empty string", i)
		}
		if _, ok := memberIDs[m.ID]; ok {
			return fmt.Errorf("members[%d]: Duplicate ID: %d", i, m.ID)
		}
		memberIDs[m.ID] = EMPTY
		if m.UserID == "" && len(m.ContactIDs) == 0 {
			return fmt.Errorf("members[%d]: m.UserID == 0 && len(m.ContactIDs) == 0", i)
		}
		if m.UserID != "" {
			if _, ok := userIDs[m.UserID]; ok {
				return fmt.Errorf("members[%d]: Duplicate UserID: %d", i, m.UserID)
			}
			userIDs[m.UserID] = EMPTY
		} else if len(m.ContactIDs) > 0 {
			for _, contactID := range m.ContactIDs {
				if _, ok := contactIDs[contactID]; ok {
					return fmt.Errorf("members[%d]: Duplicate ContactID: %d", i, contactID)
				}
				contactIDs[contactID] = EMPTY
			}
		}
		for currency, amount := range m.Balance {
			totalBalance[currency] += amount
		}
	}

	// Validate total balance is 0
	for currency, amount := range totalBalance {
		if amount != 0 {
			return errors.WithMessage(GroupTotalBalanceHasNonZeroValue, fmt.Sprintf("%v=%v", currency, amount))
		}
	}
	return nil
}

func (entity *GroupEntity) Load(ps []datastore.Property) (err error) {
	if ps, err = gaedb.CleanProperties(ps, map[string]gaedb.IsOkToRemove{
		"Status": gaedb.IsObsolete,
	}); err != nil {
		return
	}
	if err = datastore.LoadStruct(entity, ps); err != nil {
		return err
	}
	return nil
}

func (entity *GroupEntity) Save() ([]datastore.Property, error) {
	if entity.CreatorUserID == "" {
		return nil, errors.New("CreatorUserID == 0")
	}
	if strings.TrimSpace(entity.Name) == "" {
		return nil, errors.New("strings.TrimSpace(entity.Name) is empty string")
	}
	if err := entity.validateMembers(entity.GetGroupMembers(), entity.MembersCount); err != nil {
		return nil, err
	}
	ps, err := datastore.SaveStruct(entity)
	if ps, err = gaedb.CleanProperties(ps, map[string]gaedb.IsOkToRemove{
		"DefaultCurrency":       gaedb.IsEmptyString,
		"MembersCount":          gaedb.IsZeroInt,
		"MemberLastID":          gaedb.IsZeroInt,
		"MembersJson":           gaedb.IsEmptyJson,
		"Note":                  gaedb.IsEmptyString,
		"OutstandingBillsJson":  gaedb.IsEmptyJson,
		"OutstandingBillsCount": gaedb.IsZeroInt,
		"TgGroupsCount":         gaedb.IsZeroInt,
		"TgGroupsJson":          gaedb.IsEmptyJson,
	}); err != nil {
		return ps, err
	}
	if err == nil {
		if err = checkHasProperties(AppUserKind, ps); err != nil {
			return ps, err
		}
	}
	return ps, err
}
