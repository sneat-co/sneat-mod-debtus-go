package models

import (
	"context"
	"errors"
	"fmt"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/crediterra/money"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/strongo/app"
	"github.com/strongo/app/user"
	"google.golang.org/appengine/v2/datastore"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

const AppUserKind = "User"

func NewAppUserKey(appUserId int64) *dal.Key {
	if appUserId == 0 {
		return dal.NewIncompleteKey(AppUserKind, reflect.Int64, nil)
	}
	return dal.NewKeyWithID(AppUserKind, appUserId)
}

type AppUser struct {
	record.WithID[int64]
	Data *AppUserData
}

func NewAppUserRecord() dal.Record {
	return dal.NewRecordWithIncompleteKey(AppUserKind, reflect.Int64, new(AppUserData))
}

func NewAppUser(id int64, data *AppUserData) AppUser {
	key := NewAppUserKey(id)
	if data == nil {
		data = new(AppUserData)
	}
	return AppUser{
		WithID: record.NewWithID[int64](id, key, data),
		Data:   data,
	}
}

func NewAppUsers(userIDs []int64) []AppUser {
	users := make([]AppUser, len(userIDs))
	for i, id := range userIDs {
		users[i] = NewAppUser(id, nil)
	}
	return users
}

func AppUserRecords(appUsers []AppUser) (records []dal.Record) {
	records = make([]dal.Record, len(appUsers))
	for i, u := range appUsers {
		records[i] = u.Record
	}
	return
}

func IsKnownUserAccountProvider(p string) bool {
	switch p {
	case "telegram":
	case "google":
	case "fb":
	case "fbm":
	case "email":
	case "viber":
	case "line":
	case "wechat":
	default:
		return false
	}
	return true
}

type ClientInfo struct {
	UserAgent  string
	RemoteAddr string
}

func NewClientInfoFromRequest(r *http.Request) ClientInfo {
	return ClientInfo{
		UserAgent:  r.UserAgent(),
		RemoteAddr: r.RemoteAddr,
	}
}

func NewUser(clientInfo ClientInfo) AppUser {
	return AppUser{
		Data: &AppUserData{
			LastUserAgent:     clientInfo.UserAgent,
			LastUserIpAddress: clientInfo.RemoteAddr,
		},
	}
}

type AppUserData struct {
	UserRewardBalance

	SavedCounter int `datastore:"A"` // Indexing to find most active users

	IsAnonymous        bool   `datastore:",noindex"`
	PasswordBcryptHash []byte `datastore:",noindex"` // TODO: Obsolete

	ContactDetails

	DtAccessGranted time.Time `datastore:",noindex,omitempty"`
	money.Balanced
	TransfersWithInterestCount int `datastore:",noindex"`

	SmsStats
	DtCreated time.Time
	user.LastLogin

	HasDueTransfers bool `datastore:",noindex"` // TODO: Check if we really need this prop and if yes document why

	InvitedByUserID int64  `datastore:",omitempty"` // TODO: Prevent circular references! see users 6032980589936640 & 5998019824582656
	ReferredBy      string `datastore:",omitempty"`

	user.AccountsOfUser

	TelegramUserIDs    []int64 // TODO: Obsolete
	ViberBotID         string  `datastore:",noindex,omitempty"` // TODO: Obsolete
	ViberUserID        string  `datastore:",noindex,omitempty"` // TODO: Obsolete
	VkUserID           int64   `datastore:",noindex,omitempty"` // TODO: Obsolete
	GoogleUniqueUserID string  `datastore:",noindex,omitempty"` // TODO: Obsolete
	//FbUserID           string `datastore:",noindex,omitempty"` // TODO: Obsolete Facebook assigns different IDs to same FB user for FB app & Messenger app.
	//FbmUserID          string `datastore:",noindex,omitempty"` // TODO: Obsolete So we would want to keep both IDs?
	// TODO: How do we support multiple FBM bots? They will have different PSID (PageScopeID)

	OBSOLETE_CounterpartyIDs []int64 `datastore:"CounterpartyIDs,noindex,omitempty"` // TODO: Remove obsolete

	ContactsCount int    `datastore:",noindex,omitempty"` // TODO: Obsolete
	ContactsJson  string `datastore:",noindex,omitempty"` // TODO: Obsolete

	ContactsCountActive   int `datastore:",noindex,omitempty"`
	ContactsCountArchived int `datastore:",noindex,omitempty"`

	ContactsJsonActive   string `datastore:",noindex,omitempty"`
	ContactsJsonArchived string `datastore:",noindex,omitempty"`

	GroupsCountActive   int `datastore:",noindex,omitempty"`
	GroupsCountArchived int `datastore:",noindex,omitempty"`

	GroupsJsonActive   string `datastore:",noindex,omitempty"`
	GroupsJsonArchived string `datastore:",noindex,omitempty"`
	//
	billsHolder
	//
	BillsCountActive int    `datastore:",noindex,omitempty"`
	BillsJsonActive  string `datastore:",noindex,omitempty"`
	//
	BillSchedulesCountActive int    `datastore:",noindex,omitempty"`
	BillSchedulesJsonActive  string `datastore:",noindex,omitempty"`
	//
	//DebtCounterpartyIDs    []int64 `datastore:",noindex"`
	//DebtCounterpartyCount  int     `datastore:",noindex"`
	//
	PreferredLanguage string   `datastore:",noindex,omitempty"`
	PrimaryCurrency   string   `datastore:",noindex,omitempty"`
	LastCurrencies    []string `datastore:",noindex"`
	// Counts
	CountOfInvitesCreated               int `datastore:",noindex,omitempty"`
	CountOfInvitesAccepted              int `datastore:",noindex,omitempty"`
	CountOfAckTransfersByUser           int `datastore:",noindex,omitempty"` // Do not remove, need for hiding balance/history menu in Telegram
	CountOfReceiptsCreated              int `datastore:",noindex,omitempty"`
	CountOfAckTransfersByCounterparties int `datastore:",noindex,omitempty"` // Do not remove, need for hiding balance/history menu in Telegram

	LastUserAgent     string    `datastore:",noindex,omitempty"`
	LastUserIpAddress string    `datastore:",noindex,omitempty"`
	LastFeedbackAt    time.Time `datastore:",noindex,omitempty"`
	LastFeedbackRate  string    `datastore:",noindex,omitempty"`
}

func (entity *AppUserData) GetFullName() string {
	return entity.FullName()
}

func (entity *AppUserData) SetLastCurrency(v string) {
	for i, c := range entity.LastCurrencies {
		if c == v {
			if i > 0 {
				for j := i - 1; j >= 0; j-- {
					entity.LastCurrencies[j+1] = entity.LastCurrencies[j]
				}
				entity.LastCurrencies[0] = v
			}
			return
		}
	}
	entity.LastCurrencies = append([]string{v}, entity.LastCurrencies...)
	if len(entity.LastCurrencies) > 10 {
		entity.LastCurrencies = entity.LastCurrencies[:10]
	}
}

func (entity *AppUserData) TotalContactsCount() int {
	return entity.ContactsCountActive + entity.ContactsCountArchived
}

func userContactsByStatus(contacts []UserContactJson) (active, archived []UserContactJson) {
	for _, contact := range contacts {
		switch contact.Status {
		case STATUS_ACTIVE:
			contact.Status = ""
			active = append(active, contact)
		case STATUS_ARCHIVED:
			contact.Status = ""
			archived = append(archived, contact)
		case "":
			panic("Contact status is not set")
		default:
			panic("Unknown status: " + contact.Status)
		}
	}
	return
}

func (entity *AppUserData) FixObsolete() error {
	fixContactsJson := func() error {
		if entity.ContactsJson != "" {
			contacts := make([]UserContactJson, 0, entity.ContactsCount)
			if err := ffjson.Unmarshal([]byte(entity.ContactsJson), &contacts); err != nil {
				panic(fmt.Errorf("failed to unmarshal user.ContactsJson: %w", err))
			}
			contacts = fixUserContacts(contacts, "")

			active, archived := userContactsByStatus(contacts)

			if entity.ContactsCountActive = len(active); entity.ContactsCountActive == 0 {
				entity.ContactsJsonActive = ""
			} else if jsonBytes, err := ffjson.Marshal(active); err != nil {
				return err
			} else {
				entity.ContactsJsonActive = string(jsonBytes)
			}

			if entity.ContactsCountArchived = len(archived); entity.ContactsCountArchived == 0 {
				entity.ContactsJsonArchived = ""
			} else if jsonBytes, err := ffjson.Marshal(archived); err != nil {
				return err
			} else {
				entity.ContactsJsonArchived = string(jsonBytes)
			}

			//panic(fmt.Sprintf("len(contacts): %v, contacts: %v, len(archived): %v, archived: %v, ContactsJsonActive: %v", len(contacts), contacts, len(archived), archived, entity.ContactsJsonActive))

			entity.ContactsJson = ""
			entity.ContactsCount = 0
		}

		return nil
	}
	return fixContactsJson()
}

func (entity *AppUserData) ContactIDs() (ids []int64) {
	contacts := entity.Contacts()
	ids = make([]int64, len(contacts))
	for i, c := range contacts {
		ids[i] = c.ID
	}
	return ids
}

func (entity *AppUserData) RemoveContact(contactID int64) (changed bool) {
	contacts := entity.Contacts()
	for i, contact := range contacts {
		if contact.ID == contactID {
			contacts = append(contacts[:i], contacts[i+1:]...)
			entity.SetContacts(contacts)
			return true
		}
	}
	return false
}

func (u AppUser) AddOrUpdateContact(c Contact) (contactJson UserContactJson, changed bool) {
	if c.Data == nil {
		panic("c.ContactData == nil")
	}
	if u.ID != c.Data.UserID {
		panic(fmt.Sprintf("appUser.ID:%d != contact.UserID:%d", u.ID, c.Data.UserID))
	}
	contactJson = NewUserContactJson(c.ID, c.Data.Status, c.Data.FullName(), c.Data.Balanced)
	contactJson.Transfers = c.Data.GetTransfersInfo()
	contactJson.TgUserID = c.Data.TelegramUserID
	contacts := u.Data.Contacts()
	found := false
	for i, c1 := range contacts {
		if c1.ID == c.ID {
			found = true
			if !c1.Equal(contactJson) {
				contacts[i] = contactJson
				changed = true
			}
			break
		}
	}
	if !found {
		contacts = append(contacts, contactJson)
		changed = true
	}
	if changed {
		u.Data.SetContacts(contacts)
	}
	return
}

func (entity *AppUserData) SetContacts(contacts []UserContactJson) {
	{ // store to internal properties
		active, archived := userContactsByStatus(contacts)
		entity.setContacts(STATUS_ACTIVE, active)
		entity.setContacts(STATUS_ARCHIVED, archived)
	}

	{ // update balance
		balance := make(money.Balance)
		for _, contact := range contacts {
			for c, v := range contact.Balance() {
				if newVal := balance[c] + v; newVal == 0 {
					delete(balance, c)
				} else {
					balance[c] = newVal
				}
			}
		}
		if err := entity.SetBalance(balance); err != nil {
			panic(err)
		}
	}

	entity.ContactsJson = "" // TODO: Clean obsolete - remove later
	entity.ContactsCount = 0 // TODO: Clean obsolete - remove later
}

func (entity *AppUserData) setContacts(status string, contacts []UserContactJson) {
	switch status {
	case STATUS_ACTIVE:
		if entity.ContactsCountActive = len(contacts); entity.ContactsCountActive == 0 {
			entity.ContactsJsonActive = ""
		} else if jsonBytes, err := ffjson.Marshal(contacts); err != nil {
			panic(fmt.Errorf("failed to marshal contacts: %w", err))
		} else {
			entity.ContactsJsonActive = string(jsonBytes)
		}
	case STATUS_ARCHIVED:
		if entity.ContactsCountArchived = len(contacts); entity.ContactsCountArchived == 0 {
			entity.ContactsJsonArchived = ""
		} else if jsonBytes, err := ffjson.Marshal(contacts); err != nil {
			panic(fmt.Errorf("failed to marshal contacts: %w", err))
		} else {
			entity.ContactsJsonArchived = string(jsonBytes)
		}
	default:
		panic("unknown status")
	}
}

func (entity *AppUserData) Contacts() (contacts []UserContactJson) {
	return append(entity.ActiveContacts(), entity.ArchivedContacts()...)
}

func (entity *AppUserData) ContactByID(id int64) (contact *UserContactJson) {
	if id == 0 {
		panic("*AppUserData.ContactByID() => id == 0")
	}
	for _, c := range entity.ActiveContacts() {
		if c.ID == id {
			return &c
		}
	}
	for _, c := range entity.ArchivedContacts() {
		if c.ID == id {
			return &c
		}
	}
	return
}

func (entity *AppUserData) ContactsByID() (contactsByID map[int64]UserContactJson) {
	contacts := entity.Contacts()
	contactsByID = make(map[int64]UserContactJson, len(contacts))
	for _, contact := range contacts {
		contactsByID[contact.ID] = contact
	}
	return
}

func fixUserContacts(contacts []UserContactJson, status string) []UserContactJson {
	for i, c := range contacts {
		if isFixed, s := fixContactName(c.Name); isFixed {
			c.Name = s
		}
		if status != "" && c.Status != status {
			c.Status = status // Required!
		}
		contacts[i] = c
	}
	return contacts
}

func (entity *AppUserData) ActiveContacts() (contacts []UserContactJson) {
	if entity.ContactsJsonActive != "" {
		contacts = make([]UserContactJson, 0, entity.ContactsCountActive)
		if err := ffjson.Unmarshal([]byte(entity.ContactsJsonActive), &contacts); err != nil {
			panic(fmt.Errorf("failed to unmarshal user.ContactsJsonActive: %w", err))
		}
		contacts = fixUserContacts(contacts, STATUS_ACTIVE)
	}
	return
}

func (entity *AppUserData) ArchivedContacts() (contacts []UserContactJson) {
	if entity.ContactsJsonArchived != "" {
		contacts = make([]UserContactJson, 0, entity.ContactsCountArchived)
		if err := ffjson.Unmarshal([]byte(entity.ContactsJsonArchived), &contacts); err != nil {
			panic(fmt.Errorf("failed to unmarshal user.ContactsJsonArchived: %w", err))
		}
		contacts = fixUserContacts(contacts, STATUS_ARCHIVED)
	}
	return
}

func (entity *AppUserData) LatestCounterparties(limit int) (contacts []UserContactJson) { //TODO: Need implement sorting
	allCounterparties := entity.Contacts()
	if len(allCounterparties) > limit {
		contacts = make([]UserContactJson, limit)
	} else {
		contacts = make([]UserContactJson, len(allCounterparties))
	}
	for i, cp := range allCounterparties {
		if i >= limit {
			break
		}
		contacts[i] = cp
	}
	return
}

func (entity *AppUserData) ActiveContactsWithBalance() (contacts []UserContactJson) {
	activeContacts := entity.ActiveContacts()
	contacts = make([]UserContactJson, 0, len(activeContacts))
	for _, cp := range activeContacts {
		if cp.BalanceJson != nil {
			contacts = append(contacts, cp)
		}
	}
	return
}

func (entity *AppUserData) AddGroup(group Group, tgBot string) (changed bool) {
	groups := entity.ActiveGroups()
	for i, g := range groups {
		if g.ID == group.ID {
			if g.Name != group.Data.Name || g.Note != group.Data.Note || g.MembersCount != group.Data.MembersCount {
				g.Name = group.Data.Name
				g.Note = group.Data.Note
				g.MembersCount = group.Data.MembersCount
				groups[i] = g
				changed = true
			}
			if tgBot != "" {
				for _, b := range g.TgBots {
					if b == tgBot {
						goto found
					}
				}
				g.TgBots = append(g.TgBots, tgBot)
				changed = true
			found:
			}
			return
		}
	}
	g := UserGroupJson{
		ID:           group.ID,
		Name:         group.Data.Name,
		Note:         group.Data.Note,
		MembersCount: group.Data.MembersCount,
	}
	if tgBot != "" {
		g.TgBots = []string{tgBot}
	}
	groups = append(groups, g)
	entity.SetActiveGroups(groups)
	changed = true
	return
}

func (entity *AppUserData) ActiveGroups() (groups []UserGroupJson) {
	if entity.GroupsJsonActive != "" {
		if err := ffjson.Unmarshal([]byte(entity.GroupsJsonActive), &groups); err != nil {
			panic(fmt.Errorf("failed to unmarhal user.ContactsJson: %w", err))
		}
	}
	return
}

func (entity *AppUserData) SetActiveGroups(groups []UserGroupJson) {
	if len(groups) == 0 {
		entity.GroupsJsonActive = ""
		entity.GroupsCountActive = 0
	} else {
		if data, err := ffjson.Marshal(&groups); err != nil {
			panic(err.Error())
		} else {
			entity.GroupsJsonActive = (string)(data)
			entity.GroupsCountActive = len(groups)
		}
	}
}

var _ botsfw.BotAppUser = (*AppUserData)(nil)

func (entity *AppUserData) GetCurrencies() []string {
	return entity.LastCurrencies
}

func (entity *AppUserData) SetBotUserID(platform, botID, botUserID string) {
	entity.AddAccount(user.Account{
		Provider: platform,
		App:      botID,
		ID:       botUserID,
	})
}

func (entity *AppUserData) GetPreferredLocale() string {
	if entity.PreferredLanguage != "" {
		return entity.PreferredLanguage
	} else {
		return strongo.LocaleEnUS.Code5
	}
}

func (entity *AppUserData) SetPreferredLocale(code5 string) error {
	if len(code5) != 5 {
		return errors.New("code5 length should be 5")
	}
	entity.PreferredLanguage = code5
	return nil
}

func (entity *AppUserData) SetNames(first, last, user string) {
	entity.FirstName = first
	entity.LastName = last
	entity.Username = user
}

func (entity *AppUserData) Load(ps []datastore.Property) (err error) {
	// Load I and J as usual.
	p2 := make([]datastore.Property, 0, len(ps))
	for _, p := range ps {
		switch p.Name {
		case "AA":
			continue // Ignore legacy
		case "FirstDueTransferOn":
			continue // Ignore legacy
		case "ActiveGroupsJson":
			p.Name = "GroupsJsonActive"
		case "ActiveGroupsCount":
			p.Name = "GroupsCountActive"
		case "CounterpartiesCount":
			p.Name = "ContactsCount"
		case "ContactsCount":
			continue // Ignore legacy
		case "FbUserID":
			if v, ok := p.Value.(string); ok && v != "" {
				entity.AddAccount(user.Account{
					Provider: "fb",
					ID:       v,
				})
			}
			continue
		case "FmbUserID":
			if v, ok := p.Value.(string); ok && v != "" {
				entity.AddAccount(user.Account{
					Provider: "fbm",
					ID:       v,
				})
			}
			continue
		case "FbmUserID":
			if v, ok := p.Value.(string); ok && v != "" {
				entity.AddAccount(user.Account{
					Provider: "fbm",
					ID:       v,
				})
			}
			continue
		case "ViberUserID":
			continue
		case "ViberBotID":
			continue
		case "TelegramUserID":
			if telegramUserID, ok := p.Value.(int64); ok && telegramUserID != 0 {
				entity.AccountsOfUser.Accounts = append(entity.AccountsOfUser.Accounts, "telegram::"+strconv.FormatInt(telegramUserID, 10))
			}
			continue
		case "TelegramUserIDs":
			switch p.Value.(type) {
			case int64:
				if id := p.Value.(int64); id != 0 {
					entity.AccountsOfUser.Accounts = append(entity.AccountsOfUser.Accounts, "telegram::"+strconv.FormatInt(id, 10))
				}
			default:
				err = fmt.Errorf("unknown type of TelegramUserIDs value: %T", p.Value)
				return
			}
			continue
		case "GoogleUniqueUserID":
			if v, ok := p.Value.(string); ok && v != "" {
				entity.AddAccount(user.Account{
					Provider: "google",
					App:      "debtstracker",
					ID:       v,
				})
			}
		default:
			if p.Name == "CounterpartiesJson" {
				p.Name = "ContactsJson"
			}
			if p.Name == "ContactsJson" {
				contactsJson := p.Value.(string)
				if contactsJson != "" {
					entity.ContactsJson = contactsJson
					if err := entity.FixObsolete(); err != nil {
						return err
					}
					//panic(fmt.Sprintf("contactsJson: %v\n ContactsJson: %v\n ContactsJsonActive: %v", contactsJson, entity.ContactsJson, entity.ContactsJsonActive))
					if entity.ContactsCountActive > 0 {
						p.Name = "ContactsJsonActive"
						p.Value = entity.ContactsJsonActive
						p2 = append(p2, p)
						//
						p.Name = "ContactsCountActive"
						p.Value = int64(entity.ContactsCountActive)
						p2 = append(p2, p)
					}

					if entity.ContactsCountArchived > 0 {
						p.Name = "ContactsJsonArchived"
						p.Value = entity.ContactsJsonArchived
						p2 = append(p2, p)
						//
						p.Name = "ContactsCountArchived"
						p.Value = int64(entity.ContactsCountArchived)
						p2 = append(p2, p)

					}
					continue
				}
			}
		}
		p2 = append(p2, p)
	}
	if err = datastore.LoadStruct(entity, p2); err != nil {
		return
	}
	return
}

//var userPropertiesToClean = map[string]gaedb.IsOkToRemove{
//	"AA":              gaedb.IsObsolete,
//	"FmbUserID":       gaedb.IsObsolete,
//	"CounterpartyIDs": gaedb.IsObsolete,
//	//
//	"ContactsCount": gaedb.IsZeroInt,   // TODO: Obsolete
//	"ContactsJson":  gaedb.IsEmptyJSON, // TODO: Obsolete
//	//
//	"GroupsCountActive":                   gaedb.IsZeroInt,
//	"GroupsJsonActive":                    gaedb.IsEmptyJSON,
//	"GroupsCountArchived":                 gaedb.IsZeroInt,
//	"GroupsJsonArchived":                  gaedb.IsEmptyJSON,
//	"BillsCountActive":                    gaedb.IsZeroInt,
//	"BillsJsonActive":                     gaedb.IsEmptyJSON,
//	"BillSchedulesCountActive":            gaedb.IsZeroInt,
//	"BillSchedulesJsonActive":             gaedb.IsEmptyJSON,
//	"BalanceCount":                        gaedb.IsZeroInt,
//	"BalanceJson":                         gaedb.IsEmptyString,
//	"CountOfAckTransfersByCounterparties": gaedb.IsZeroInt,
//	"CountOfAckTransfersByUser":           gaedb.IsZeroInt,
//	"CountOfInvitesAccepted":              gaedb.IsZeroInt,
//	"CountOfInvitesCreated":               gaedb.IsZeroInt,
//	"CountOfReceiptsCreated":              gaedb.IsZeroInt,
//	"CountOfTransfers":                    gaedb.IsZeroInt,
//	"ContactsCountActive":                 gaedb.IsZeroInt,
//	"ContactsJsonActive":                  gaedb.IsEmptyJSON,
//	"ContactsCountArchived":               gaedb.IsZeroInt,
//	"ContactsJsonArchived":                gaedb.IsEmptyJSON,
//	"DtAccessGranted":                     gaedb.IsZeroTime,
//	"EmailAddress":                        gaedb.IsEmptyString,
//	"EmailAddressOriginal":                gaedb.IsEmptyString,
//	"FirstName":                           gaedb.IsEmptyString,
//	"HasDueTransfers":                     gaedb.IsFalse,
//	"InvitedByUserID":                     gaedb.IsZeroInt,
//	"IsAnonymous":                         gaedb.IsFalse,
//	"LastName":                            gaedb.IsEmptyString,
//	"LastTransferAt":                      gaedb.IsZeroTime,
//	"LastTransferID":                      gaedb.IsZeroInt,
//	"LastFeedbackAt":                      gaedb.IsZeroTime,
//	"LastFeedbackRate":                    gaedb.IsEmptyString,
//	"LastUserAgent":                       gaedb.IsEmptyString,
//	"LastUserIpAddress":                   gaedb.IsEmptyString,
//	"Nickname":                            gaedb.IsEmptyString,
//	"PhoneNumber":                         gaedb.IsZeroInt,
//	"PhoneNumberConfirmed":                gaedb.IsFalse, // TODO: Duplicate of PhoneNumberIsConfirmed
//	"PhoneNumberIsConfirmed":              gaedb.IsFalse, // TODO: Duplicate of PhoneNumberConfirmed
//	"EmailConfirmed":                      gaedb.IsFalse,
//	"PreferredLanguage":                   gaedb.IsEmptyString,
//	"PrimaryCurrency":                     gaedb.IsEmptyString,
//	"ScreenName":                          gaedb.IsEmptyString,
//	"SmsCost":                             gaedb.IsZeroFloat,
//	"SmsCostUSD":                          gaedb.IsZeroInt,
//	"SmsCount":                            gaedb.IsZeroInt,
//	"Username":                            gaedb.IsEmptyString,
//	"VkUserID":                            gaedb.IsZeroInt,
//	"DtLastLogin":                         gaedb.IsZeroTime,
//	"PasswordBcryptHash":                  gaedb.IsObsolete,
//	"TransfersWithInterestCount":          gaedb.IsZeroInt,
//	//
//	"ViberBotID":         gaedb.IsObsolete,
//	"ViberUserID":        gaedb.IsObsolete,
//	"FbmUserID":          gaedb.IsObsolete,
//	"FbUserID":           gaedb.IsObsolete,
//	"FbUserIDs":          gaedb.IsObsolete,
//	"GoogleUniqueUserID": gaedb.IsObsolete,
//	"TelegramUserID":     gaedb.IsObsolete,
//	"TelegramUserIDs":    gaedb.IsObsolete,
//	//
//}

func (entity *AppUserData) cleanProps(properties []datastore.Property) ([]datastore.Property, error) {
	var err error
	//if properties, err = gaedb.CleanProperties(properties, userPropertiesToClean); err != nil {
	//	return properties, err
	//}
	//if properties, err = entity.UserRewardBalance.cleanProperties(properties); err != nil {
	//	return properties, err
	//}
	return properties, err
}

func (entity *AppUserData) TotalBalanceFromContacts() (balance money.Balance) {
	balance = make(money.Balance, entity.BalanceCount)

	for _, contact := range entity.Contacts() {
		for currency, cv := range contact.Balance() {
			if v := balance[currency] + cv; v == 0 {
				delete(balance, currency)
			} else {
				balance[currency] = v
			}
		}
	}

	return
}

var ErrDuplicateContactName = errors.New("user has at least 2 contacts with same name")
var ErrDuplicateTgUserID = errors.New("user has at least 2 contacts with same TgUserID")

func (entity *AppUserData) BeforeSave() (err error) {
	if entity.GroupsJsonActive != "" && entity.GroupsCountActive == 0 {
		return errors.New(`entity.GroupsJsonActive != "" && entity.GroupsCountActive == 0`)
	}

	contacts := entity.Contacts()

	if len(contacts) != entity.ContactsCountActive+entity.ContactsCountArchived {
		panic("len(contacts) != entity.ContactsCountActive + entity.ContactsCountArchived")
	}

	contactsCount := len(contacts)

	contactsByName := make(map[string]int, contactsCount)
	contactsByTgUserID := make(map[int64]int, contactsCount)

	entity.TransfersWithInterestCount = 0
	for i, contact := range contacts {
		if contact.ID == 0 {
			panic(fmt.Sprintf("contact[%d].ID == 0, contact: %v, contacts: %v", i, contact, contacts))
		}
		if contact.Name == "" {
			panic(fmt.Sprintf("contact[%d].ContactName is Empty string, contact: %v, contacts: %v", i, contact, contacts))
		}
		if contact.Status == "" {
			panic(fmt.Sprintf("contact[%d].Status is Empty string, contact: %v, contacts: %v", i, contact, contacts))
		}
		{
			if duplicateOf, isDuplicate := contactsByName[contact.Name]; isDuplicate {
				err = fmt.Errorf("%w: id1=%d, id2=%d => %v", ErrDuplicateContactName, contacts[duplicateOf].ID, contact.ID, contact.Name)
				return
			}
			contactsByName[contact.Name] = i
		}
		if contact.TgUserID != 0 {
			if duplicateOf, isDuplicate := contactsByTgUserID[contact.TgUserID]; isDuplicate {
				err = fmt.Errorf("%s: %d for contacts %d & %d", ErrDuplicateTgUserID, contact.TgUserID, contacts[duplicateOf].ID, contact.ID)
				return
			}
			contactsByTgUserID[contact.TgUserID] = i
		}
		if contact.Transfers != nil {
			entity.TransfersWithInterestCount += len(contact.Transfers.OutstandingWithInterest)
		}
	}
	return
}

func (entity *AppUserData) Save() (properties []datastore.Property, err error) {
	if err = entity.BeforeSave(); err != nil {
		return
	}

	//entity.SavedCounter += 1
	if properties, err = datastore.SaveStruct(entity); err != nil {
		return
	}
	if properties, err = entity.cleanProps(properties); err != nil {
		return
	}

	//checkHasProperties(AppUserKind, properties)
	return properties, err
}

func (entity *AppUserData) BalanceWithInterest(c context.Context, periodEnds time.Time) (balance money.Balance, err error) {
	if entity.TransfersWithInterestCount == 0 {
		balance = entity.Balance()
	} else if entity.TransfersWithInterestCount > 0 {
		//var (
		//	userBalance Balance
		//)
		//userBalance = entity.Balance()
		balance = make(money.Balance, entity.BalanceCount)
		for _, contact := range entity.Contacts() {
			var contactBalance money.Balance
			if contactBalance, err = contact.BalanceWithInterest(c, periodEnds); err != nil {
				err = fmt.Errorf("failed to get balance with interest for user's contact JSON %v: %w", contact.ID, err)
				return
			}
			for currency, value := range contactBalance {
				balance[currency] += value
			}
		}
		//if len(balance) != entity.BalanceCount { // Theoretically can be eliminated by interest
		//	panic(fmt.Sprintf("len(balance) != entity.BalanceCount: %v != %v", len(balance), entity.BalanceCount))
		//}
		//for c, v := range balance { // It can be less if we have different interest condition for 2 contacts different direction!!!
		//	if tv := userBalance[c]; v < tv {
		//		panic(fmt.Sprintf("For currency %v balance with interest is less than total balance: %v < %v", c, v, tv))
		//	}
		//}
	} else {
		panic(fmt.Sprintf("TransfersWithInterestCount > 0: %v", entity.TransfersWithInterestCount))
	}
	return
}

func (entity *AppUserData) GetOutstandingBalance() (balance money.Balance) {
	balance = make(money.Balance, 2)
	for _, bill := range entity.GetOutstandingBills() {
		balance[bill.Currency] += bill.UserBalance
	}
	return
}
