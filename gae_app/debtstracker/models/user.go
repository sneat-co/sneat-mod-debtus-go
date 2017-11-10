package models

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/strongo/app"
	"github.com/strongo/app/db"
	"github.com/strongo/app/gaedb"
	"github.com/strongo/app/user"
	"github.com/strongo/bots-framework/core"
	"google.golang.org/appengine/datastore"
	"net/http"
	"strconv"
	"time"
)

const AppUserKind = "User"

type AppUser struct {
	db.NoStrID
	ID int64
	*AppUserEntity
}

func (_ *AppUser) Kind() string {
	return AppUserKind
}

func (u *AppUser) IntID() int64 {
	return u.ID
}

func (u *AppUser) Entity() interface{} {
	if u.AppUserEntity == nil {
		u.AppUserEntity = new(AppUserEntity)
	}
	return u.AppUserEntity
}

func (u *AppUser) SetEntity(entity interface{}) {
	u.AppUserEntity = entity.(*AppUserEntity)
}

func (u *AppUser) SetIntID(id int64) {
	u.ID = id
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
		AppUserEntity: &AppUserEntity{
			LastUserAgent:     clientInfo.UserAgent,
			LastUserIpAddress: clientInfo.RemoteAddr,
		},
	}
}

type AppUserEntity struct {
	UserRewardBalance

	SavedCounter int `datastore:"A"` // Indexing to find most active users

	IsAnonymous        bool   `datastore:",noindex"`
	PasswordBcryptHash []byte `datastore:",noindex"` // TODO: Obsolete

	ContactDetails

	DtAccessGranted            time.Time `datastore:",noindex,omitempty"`
	Balanced
	TransfersWithInterestCount int       `datastore:",noindex"`

	SmsStats
	DtCreated time.Time
	user.LastLogin

	HasDueTransfers bool `datastore:",noindex"` // TODO: Check if we really need this prop and if yes document why

	InvitedByUserID int64 `datastore:",omitempty"`

	user.Accounts

	TelegramUserIDs    []int64                                 // TODO: Obsolete
	ViberBotID         string `datastore:",noindex,omitempty"` // TODO: Obsolete
	ViberUserID        string `datastore:",noindex,omitempty"` // TODO: Obsolete
	VkUserID           int64  `datastore:",noindex,omitempty"` // TODO: Obsolete
	GoogleUniqueUserID string `datastore:",noindex,omitempty"` // TODO: Obsolete
	//FbUserID           string `datastore:",noindex"` // TODO: Obsolete Facebook assigns different IDs to same FB user for FB app & Messenger app.
	//FbmUserID          string `datastore:",noindex"` // TODO: Obsolete So we would want to keep both IDs?
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

func (u *AppUserEntity) SetLastCurrency(v string) {
	for i, c := range u.LastCurrencies {
		if c == v {
			if i > 0 {
				for j := i - 1; j >= 0; j-- {
					u.LastCurrencies[j+1] = u.LastCurrencies[j]
				}
				u.LastCurrencies[0] = v
			}
			return
		}
	}
	u.LastCurrencies = append([]string{v}, u.LastCurrencies...)
	if len(u.LastCurrencies) > 10 {
		u.LastCurrencies = u.LastCurrencies[:10]
	}
}

func (u *AppUserEntity) TotalContactsCount() int {
	return u.ContactsCountActive + u.ContactsCountArchived
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

func (u *AppUserEntity) FixObsolete() error {
	fixContactsJson := func() error {
		if u.ContactsJson != "" {
			contacts := make([]UserContactJson, 0, u.ContactsCount)
			if err := ffjson.Unmarshal([]byte(u.ContactsJson), &contacts); err != nil {
				panic(errors.Wrap(err, "Failed to unmarshal user.ContactsJson").Error())
			}
			contacts = fixUserContacts(contacts, "")

			active, archived := userContactsByStatus(contacts)

			if u.ContactsCountActive = len(active); u.ContactsCountActive == 0 {
				u.ContactsJsonActive = ""
			} else if jsonBytes, err := ffjson.Marshal(active); err != nil {
				return err
			} else {
				u.ContactsJsonActive = string(jsonBytes)
			}

			if u.ContactsCountArchived = len(archived); u.ContactsCountArchived == 0 {
				u.ContactsJsonArchived = ""
			} else if jsonBytes, err := ffjson.Marshal(archived); err != nil {
				return err
			} else {
				u.ContactsJsonArchived = string(jsonBytes)
			}

			//panic(fmt.Sprintf("len(contacts): %v, contacts: %v, len(archived): %v, archived: %v, ContactsJsonActive: %v", len(contacts), contacts, len(archived), archived, u.ContactsJsonActive))

			u.ContactsJson = ""
			u.ContactsCount = 0
		}

		return nil
	}
	return fixContactsJson()
}

func (u *AppUserEntity) CounterpartiesIDs() (ids []int64) {
	counterparties := u.Contacts()
	ids = make([]int64, len(counterparties))
	for i, c := range counterparties {
		ids[i] = c.ID
	}
	return ids
}

func (u *AppUserEntity) RemoveContact(contactID int64) (changed bool) {
	contacts := u.Contacts()
	for i, contact := range contacts {
		if contact.ID == contactID {
			contacts = append(contacts[:i], contacts[i+1:]...)
			u.SetContacts(contacts)
			return true
		}
	}
	return false
}

func (u AppUser) AddOrUpdateContact(c Contact) (changed bool) {
	if c.ContactEntity == nil {
		panic("c.ContactEntity == nil")
	}
	if u.ID != c.UserID {
		panic(fmt.Sprintf("appUser.ID:%d != contact.UserID:%d", u.ID, c.UserID))
	}
	c2 := NewUserContactJson(c.ID, c.Status, c.FullName(), c.Balanced)
	c2.Transfers = c.GetTransfersInfo()
	c2.TgUserID = c.TelegramUserID
	contacts := u.Contacts()
	found := false
	for i, c1 := range contacts {
		if c1.ID == c.ID {
			found = true
			if !c1.Equal(c2) {
				contacts[i] = c2
				changed = true
			}
			break
		}
	}
	if !found {
		contacts = append(contacts, c2)
		changed = true
	}
	if changed {
		u.SetContacts(contacts)
	}
	return
}

func (u *AppUserEntity) SetContacts(contacts []UserContactJson) {

	active, archived := userContactsByStatus(contacts)

	if u.ContactsCountActive = len(active); u.ContactsCountActive == 0 {
		u.ContactsJsonActive = ""
	} else if jsonBytes, err := ffjson.Marshal(active); err != nil {
		panic(errors.Wrap(err, "Failed to marshal contacts").Error())
	} else {
		u.ContactsJsonActive = string(jsonBytes)
	}

	if u.ContactsCountArchived = len(archived); u.ContactsCountArchived == 0 {
		u.ContactsJsonArchived = ""
	} else if jsonBytes, err := ffjson.Marshal(archived); err != nil {
		panic(errors.Wrap(err, "Failed to marshal contacts").Error())
	} else {
		u.ContactsJsonArchived = string(jsonBytes)
	}

	u.ContactsJson = ""
	u.ContactsCount = 0
}

func (u *AppUserEntity) Contacts() (contacts []UserContactJson) {
	return append(u.ActiveContacts(), u.ArchivedContacts()...)
}

func (u *AppUserEntity) ContactsByID() (contactsByID map[int64]UserContactJson) {
	contacts := u.Contacts()
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

func (u *AppUserEntity) ActiveContacts() (contacts []UserContactJson) {
	if u.ContactsJsonActive != "" {
		contacts = make([]UserContactJson, 0, u.ContactsCountActive)
		if err := ffjson.Unmarshal([]byte(u.ContactsJsonActive), &contacts); err != nil {
			panic(errors.Wrap(err, "Failed to unmarshal user.ContactsJsonActive").Error())
		}
		contacts = fixUserContacts(contacts, STATUS_ACTIVE)
	}
	return
}

func (u *AppUserEntity) ArchivedContacts() (contacts []UserContactJson) {
	if u.ContactsJsonArchived != "" {
		contacts = make([]UserContactJson, 0, u.ContactsCountArchived)
		if err := ffjson.Unmarshal([]byte(u.ContactsJsonArchived), &contacts); err != nil {
			panic(errors.Wrap(err, "Failed to unmarshal user.ContactsJsonArchived").Error())
		}
		contacts = fixUserContacts(contacts, STATUS_ARCHIVED)
	}
	return
}

func (u *AppUserEntity) GetContactInfoByID(contactID int64) (UserContactJson, error) {
	for _, c := range u.Contacts() {
		if c.ID == contactID {
			return c, nil
		}
	}
	return UserContactJson{}, errors.New(fmt.Sprintf("Contact not found in user.ContactsJson by ID=%d", contactID))
}

func (u *AppUserEntity) LatestCounterparties(limit int) (contacts []UserContactJson) { //TODO: Need implement sorting
	allCounterparties := u.Contacts()
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

func (u *AppUserEntity) ActiveContactsWithBalance() (contacts []UserContactJson) {
	activeContacts := u.ActiveContacts()
	contacts = make([]UserContactJson, 0, len(activeContacts))
	for _, cp := range activeContacts {
		if cp.BalanceJson != nil {
			contacts = append(contacts, cp)
		}
	}
	return
}

func (u *AppUserEntity) AddGroup(group Group, tgBot string) (changed bool) {
	groups := u.ActiveGroups()
	for i, g := range groups {
		if g.ID == group.ID {
			if g.Name != group.Name || g.Note != group.Note || g.MembersCount != group.MembersCount {
				g.Name = group.Name
				g.Note = group.Note
				g.MembersCount = group.MembersCount
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
		Name:         group.Name,
		Note:         group.Note,
		MembersCount: group.MembersCount,
	}
	if tgBot != "" {
		g.TgBots = []string{tgBot}
	}
	groups = append(groups, g)
	u.SetActiveGroups(groups)
	changed = true
	return
}

func (u *AppUserEntity) ActiveGroups() (groups []UserGroupJson) {
	if u.GroupsJsonActive != "" {
		if err := ffjson.Unmarshal([]byte(u.GroupsJsonActive), &groups); err != nil {
			panic(errors.Wrap(err, "Failed to unmarhal user.ContactsJson").Error())
		}
	}
	return
}

func (u *AppUserEntity) SetActiveGroups(groups []UserGroupJson) {
	if len(groups) == 0 {
		u.GroupsJsonActive = ""
		u.GroupsCountActive = 0
	} else {
		if data, err := ffjson.Marshal(&groups); err != nil {
			panic(err.Error())
		} else {
			u.GroupsJsonActive = (string)(data)
			u.GroupsCountActive = len(groups)
		}
	}
}

var _ bots.BotAppUser = (*AppUserEntity)(nil)

func (u *AppUserEntity) GetCurrencies() []string {
	return u.LastCurrencies
}

func (u *AppUserEntity) SetBotUserID(platform, botID, botUserID string) {
	u.AddAccount(user.Account{
		Provider: platform,
		App:      botID,
		ID:       botUserID,
	})
}

func (u *AppUserEntity) PreferredLocale() string {
	if u.PreferredLanguage != "" {
		return u.PreferredLanguage
	} else {
		return strongo.LocaleEnUS.Code5
	}
}

func (u *AppUserEntity) SetPreferredLocale(code5 string) error {
	if len(code5) != 5 {
		return errors.New("code5 length should be 5")
	}
	u.PreferredLanguage = code5
	return nil
}

func (u *AppUserEntity) SetNames(first, last, user string) {
	u.FirstName = first
	u.LastName = last
	u.Username = user
}

func (u *AppUserEntity) Load(ps []datastore.Property) (err error) {
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
			continue
		case "FbmUserID":
			continue
		case "ViberUserID":
			continue
		case "ViberBotID":
			continue
		case "TelegramUserID":
			if telegramUserID, ok := p.Value.(int64); ok && telegramUserID != 0 {
				u.AddAccount(user.Account{
					Provider: "telegram",
					ID:       strconv.FormatInt(telegramUserID, 10),
				})
			}
		case "TelegramUserIDs":
			switch p.Value.(type) {
			case int64:
				if id := p.Value.(int64); id != 0 {
					u.AddAccount(user.Account{
						Provider: "telegram",
						ID:       strconv.FormatInt(id, 10),
					})
				}
			case []int64:
				for _, id := range p.Value.([]int64) {
					if id != 0 {
						u.AddAccount(user.Account{
							Provider: "telegram",
							ID:       strconv.FormatInt(id, 10),
						})
					}
				}
			default:
				err = fmt.Errorf("Unknown type of TelegramUserIDs value: %T", p.Value)
			}
		case "GoogleUniqueUserID":
			if v, ok := p.Value.(string); ok && v != "" {
				u.AddAccount(user.Account{
					Provider: "google",
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
					u.ContactsJson = contactsJson
					if err := u.FixObsolete(); err != nil {
						return err
					}
					//panic(fmt.Sprintf("contactsJson: %v\n ContactsJson: %v\n ContactsJsonActive: %v", contactsJson, u.ContactsJson, u.ContactsJsonActive))
					if u.ContactsCountActive > 0 {
						p.Name = "ContactsJsonActive"
						p.Value = u.ContactsJsonActive
						p2 = append(p2, p)
						//
						p.Name = "ContactsCountActive"
						p.Value = int64(u.ContactsCountActive)
						p2 = append(p2, p)
					}

					if u.ContactsCountArchived > 0 {
						p.Name = "ContactsJsonArchived"
						p.Value = u.ContactsJsonArchived
						p2 = append(p2, p)
						//
						p.Name = "ContactsCountArchived"
						p.Value = int64(u.ContactsCountArchived)
						p2 = append(p2, p)

					}
					continue
				}
			}
		}
		p2 = append(p2, p)
	}
	if err = datastore.LoadStruct(u, p2); err != nil {
		return
	}
	return
}

var userPropertiesToClean = map[string]gaedb.IsOkToRemove{
	"AA":              gaedb.IsObsolete,
	"FmbUserID":       gaedb.IsObsolete,
	"CounterpartyIDs": gaedb.IsObsolete,
	//
	"ContactsCount": gaedb.IsZeroInt,   // TODO: Obsolete
	"ContactsJson":  gaedb.IsEmptyJson, // TODO: Obsolete
	//
	"GroupsCountActive":                   gaedb.IsZeroInt,
	"GroupsJsonActive":                    gaedb.IsEmptyJson,
	"GroupsCountArchived":                 gaedb.IsZeroInt,
	"GroupsJsonArchived":                  gaedb.IsEmptyJson,
	"BillsCountActive":                    gaedb.IsZeroInt,
	"BillsJsonActive":                     gaedb.IsEmptyJson,
	"BillSchedulesCountActive":            gaedb.IsZeroInt,
	"BillSchedulesJsonActive":             gaedb.IsEmptyJson,
	"BalanceCount":                        gaedb.IsZeroInt,
	"BalanceJson":                         gaedb.IsEmptyString,
	"CountOfAckTransfersByCounterparties": gaedb.IsZeroInt,
	"CountOfAckTransfersByUser":           gaedb.IsZeroInt,
	"CountOfInvitesAccepted":              gaedb.IsZeroInt,
	"CountOfInvitesCreated":               gaedb.IsZeroInt,
	"CountOfReceiptsCreated":              gaedb.IsZeroInt,
	"CountOfTransfers":                    gaedb.IsZeroInt,
	"ContactsCountActive":                 gaedb.IsZeroInt,
	"ContactsJsonActive":                  gaedb.IsEmptyJson,
	"ContactsCountArchived":               gaedb.IsZeroInt,
	"ContactsJsonArchived":                gaedb.IsEmptyJson,
	"DtAccessGranted":                     gaedb.IsZeroTime,
	"EmailAddress":                        gaedb.IsEmptyString,
	"EmailAddressOriginal":                gaedb.IsEmptyString,
	"FirstName":                           gaedb.IsEmptyString,
	"HasDueTransfers":                     gaedb.IsFalse,
	"InvitedByUserID":                     gaedb.IsZeroInt,
	"IsAnonymous":                         gaedb.IsFalse,
	"LastName":                            gaedb.IsEmptyString,
	"LastTransferAt":                      gaedb.IsZeroTime,
	"LastTransferID":                      gaedb.IsZeroInt,
	"LastFeedbackAt":                      gaedb.IsZeroTime,
	"LastFeedbackRate":                    gaedb.IsEmptyString,
	"LastUserAgent":                       gaedb.IsEmptyString,
	"LastUserIpAddress":                   gaedb.IsEmptyString,
	"Nickname":                            gaedb.IsEmptyString,
	"PhoneNumber":                         gaedb.IsZeroInt,
	"PhoneNumberConfirmed":                gaedb.IsFalse, // TODO: Duplicate of PhoneNumberIsConfirmed
	"PhoneNumberIsConfirmed":              gaedb.IsFalse, // TODO: Duplicate of PhoneNumberConfirmed
	"EmailConfirmed":                      gaedb.IsFalse,
	"PreferredLanguage":                   gaedb.IsEmptyString,
	"PrimaryCurrency":                     gaedb.IsEmptyString,
	"ScreenName":                          gaedb.IsEmptyString,
	"SmsCost":                             gaedb.IsZeroFloat,
	"SmsCostUSD":                          gaedb.IsZeroInt,
	"SmsCount":                            gaedb.IsZeroInt,
	"Username":                            gaedb.IsEmptyString,
	"VkUserID":                            gaedb.IsZeroInt,
	"DtLastLogin":                         gaedb.IsZeroTime,
	"PasswordBcryptHash":                  gaedb.IsObsolete,
	"TransfersWithInterestCount":          gaedb.IsZeroInt,
	//
	"ViberBotID":         gaedb.IsObsolete,
	"ViberUserID":        gaedb.IsObsolete,
	"FbmUserID":          gaedb.IsObsolete,
	"FbUserID":           gaedb.IsObsolete,
	"FbUserIDs":          gaedb.IsObsolete,
	"GoogleUniqueUserID": gaedb.IsObsolete,
	"TelegramUserID":     gaedb.IsObsolete,
	"TelegramUserIDs":    gaedb.IsObsolete,
	//
}

func (u *AppUserEntity) cleanProps(properties []datastore.Property) ([]datastore.Property, error) {
	var err error
	if properties, err = gaedb.CleanProperties(properties, userPropertiesToClean); err != nil {
		return properties, err
	}
	if properties, err = u.UserRewardBalance.cleanProperties(properties); err != nil {
		return properties, err
	}
	return properties, err
}

func (u *AppUserEntity) Save() (properties []datastore.Property, err error) {
	if u.GroupsJsonActive != "" && u.GroupsCountActive == 0 {
		return nil, errors.New(`u.GroupsJsonActive != "" && u.GroupsCountActive == 0`)
	}

	contacts := u.Contacts()

	if len(contacts) != u.ContactsCountActive+u.ContactsCountArchived {
		panic("len(contacts) != u.ContactsCountActive + u.ContactsCountArchived")
	}

	contactsCount := len(contacts)

	contactsByName := make(map[string]int, contactsCount)
	contactsByTgUserID := make(map[int64]int, contactsCount)

	u.TransfersWithInterestCount = 0
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
				panic(fmt.Sprintf("User has at least 2 contacts with same name: `%d` and `%d`.", contacts[duplicateOf].ID, contact.ID))
			}
			contactsByName[contact.Name] = i
		}
		if contact.TgUserID != 0 {
			if duplicateOf, isDuplicate := contactsByTgUserID[contact.TgUserID]; isDuplicate {
				panic(fmt.Sprintf("User has at least 2 contacts with same TgUserID=%d: %d and %d", contact.TgUserID, contacts[duplicateOf].ID, contact.ID))
			}
			contactsByTgUserID[contact.TgUserID] = i
		}
		if contact.Transfers != nil {
			u.TransfersWithInterestCount += len(contact.Transfers.OutstandingWithInterest)
		}
	}

	u.SavedCounter += 1
	if properties, err = datastore.SaveStruct(u); err != nil {
		return
	}
	if properties, err = u.cleanProps(properties); err != nil {
		return
	}

	checkHasProperties(AppUserKind, properties)
	return properties, err
}

func (u *AppUserEntity) BalanceWithInterest(periodEnds time.Time) (balance Balance) {
	if u.TransfersWithInterestCount == 0 {
		balance = u.Balance()
	} else if u.TransfersWithInterestCount > 0 {
		var (
			userBalance Balance
		)
		userBalance = u.Balance()
		balance = make(Balance, u.BalanceCount)
		for _, contact := range u.Contacts() {
			for currency, value := range contact.BalanceWithInterest(nil, periodEnds) {
				balance[currency] += value
			}
		}
		if len(balance) != u.BalanceCount {
			panic(fmt.Sprintf("len(balance) != u.BalanceCount: %v != %v", len(balance), u.BalanceCount))
		}
		for c, v := range balance {
			if tv := userBalance[c]; v < tv {
				panic(fmt.Sprintf("For currency %v balance with interest is less then total balance: %v < %v", c, v, tv))
			}
		}
	} else {
		panic(fmt.Sprintf("TransfersWithInterestCount > 0: %v", u.TransfersWithInterestCount))
	}
	return
}
