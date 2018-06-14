package facade

import (
	"fmt"
	"reflect"
	"strconv"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/pkg/errors"
	"github.com/sanity-io/litter"
	"github.com/strongo/db"
	"github.com/strongo/log"
	"github.com/crediterra/money"
)

func ChangeContactStatus(c context.Context, contactID int64, newStatus string) (contact models.Contact, err error) {
	err = dal.DB.RunInTransaction(c, func(c context.Context) error {
		if contact, err = GetContactByID(c, contactID); err != nil {
			return err
		}
		if contact.Status != newStatus {
			contact.Status = newStatus
			var user models.AppUser
			if user, err = User.GetUserByID(c, contact.UserID); err != nil {
				return err
			}
			if _, userChanged := user.AddOrUpdateContact(contact); userChanged {
				err = dal.DB.UpdateMulti(c, []db.EntityHolder{&contact, &user})
			} else {
				err = dal.DB.Update(c, &contact)
			}
		}
		return err
	}, dal.CrossGroupTransaction)
	return
}

func createContactWithinTransaction(
	tc context.Context,
	changes *createContactDbChanges,
	counterpartyUserID int64,
	contactDetails models.ContactDetails,
) (
	contact models.Contact,
	counterpartyContact models.Contact,
	err error,
) {
	appUser := *changes.user
	if changes.counterpartyContact != nil {
		counterpartyContact = *changes.counterpartyContact
	}

	log.Debugf(tc, "createContactWithinTransaction(appUser.ID=%v, counterpartyDetails=%v)", appUser.ID, contactDetails)
	if !dal.DB.IsInTransaction(tc) {
		err = errors.New("createContactWithinTransaction is called outside of transaction")
		return
	}
	if appUser.ID == 0 {
		err = errors.New("appUser.ID == 0")
		return
	}
	if appUser.AppUserEntity == nil {
		err = errors.New("appUser.AppUserEntity == nil")
		return
	}
	if appUser.ID == counterpartyUserID {
		panic(fmt.Sprintf("appUser.ID == counterpartyUserID: %v", counterpartyUserID))
	}
	if counterpartyContact.ContactEntity != nil && counterpartyContact.ID == 0 {
		panic(fmt.Sprintf("counterpartyContact.ContactEntity != nil && counterpartyContact.ID == 0: %v", litter.Sdump(counterpartyContact)))
	}

	contact.ContactEntity = models.NewContactEntity(appUser.ID, contactDetails)
	if counterpartyContact.ID != 0 {
		if counterpartyContact.ContactEntity == nil {
			if counterpartyContact, err = GetContactByID(tc, counterpartyContact.ID); err != nil {
				return
			}
			changes.counterpartyContact = &counterpartyContact
		}
		if counterpartyContact.UserID != counterpartyUserID {
			if counterpartyUserID == 0 {
				counterpartyUserID = counterpartyContact.UserID
			} else {
				panic(fmt.Sprintf("counterpartyContact.UserID != counterpartyUserID: %v != %v", counterpartyContact.UserID, counterpartyUserID))
			}
		}
		contact.CounterpartyUserID = counterpartyUserID
		contact.CounterpartyCounterpartyID = counterpartyContact.ID
		contact.TransfersJson = counterpartyContact.TransfersJson
		contact.Balanced = money.Balanced{
			CountOfTransfers: counterpartyContact.CountOfTransfers,
			LastTransferID:   counterpartyContact.LastTransferID,
			LastTransferAt:   counterpartyContact.LastTransferAt,
		}
		invitedCounterpartyBalance := counterpartyContact.Balance().Reversed()
		log.Debugf(tc, "invitedCounterpartyBalance: %v", invitedCounterpartyBalance)
		contact.SetBalance(invitedCounterpartyBalance)
		contact.MustMatchCounterparty(counterpartyContact)
	}

	if contact, err = dal.Contact.InsertContact(tc, contact.ContactEntity); err != nil {
		return
	}

	if counterpartyContact.ID != 0 {
		if counterpartyContact.CounterpartyCounterpartyID == 0 {
			counterpartyContact.CounterpartyCounterpartyID = contact.ID
			if counterpartyContact.CounterpartyUserID == 0 {
				counterpartyContact.CounterpartyUserID = contact.UserID
			} else {
				err = fmt.Errorf("inviter contact %v already has CounterpartyUserID=%v", counterpartyContact.ID, counterpartyContact.CounterpartyUserID)
				return
			}
			changes.FlagAsChanged(changes.counterpartyContact)
		} else if counterpartyContact.CounterpartyCounterpartyID != contact.ID {
			err = fmt.Errorf("inviter contact %v already has CounterpartyCounterpartyID=%v", counterpartyContact.ID, counterpartyContact.CounterpartyCounterpartyID)
			return
		}
	}

	if _, changed := appUser.AddOrUpdateContact(contact); changed {
		changes.FlagAsChanged(changes.user)
	}

	{ // Verifications for data integrity
		if counterpartyContact.ContactEntity != nil {
			contact.MustMatchCounterparty(counterpartyContact)
		}
		if contact.UserID != appUser.ID {
			panic(fmt.Sprintf("contact.UserID != appUser.ID: %v != %v", contact.UserID, appUser.ID))
		}
		if counterpartyContact.ContactEntity != nil {
			if counterpartyContact.UserID != counterpartyUserID {
				panic(fmt.Sprintf("counterpartyContact.UserID != counterpartyUserID: %v != %v", counterpartyContact.UserID, counterpartyUserID))
			}
			if contact.ID == counterpartyContact.ID {
				panic(fmt.Sprintf("contact.ID == counterpartyContact.ID: %v", contact.ID))
			}
			if contact.UserID == counterpartyContact.UserID {
				panic(fmt.Sprintf("contact.UserID == counterpartyContact.UserID: %v", contact.UserID))
			}
			if contact.TransfersJson != counterpartyContact.TransfersJson {
				log.Errorf(tc, "contact.TransfersJson != counterpartyContact.TransfersJson\n contact: %v\n counterpartyContact: %v", contact.TransfersJson, counterpartyContact.TransfersJson)
			}
			if contact.BalanceCount != counterpartyContact.BalanceCount {
				panic(fmt.Sprintf("contact.BalanceCount != counterpartyContact.BalanceCount: %v != %v", contact.BalanceCount, counterpartyContact.BalanceCount))
			}
			if cBalance, cpBalance := contact.Balance(), counterpartyContact.Balance(); !cBalance.Equal(cpBalance.Reversed()) {
				panic(fmt.Sprintf("!contact.Balance().Equal(counterpartyContact.Balance())\ncontact.Balance(): %v\n counterpartyContact.Balance(): %v", cBalance, cpBalance))
			}
		}
		appUserContactJson := appUser.ContactByID(contact.ID)
		if ucBalance, cBalance := appUserContactJson.Balance(), contact.Balance(); !ucBalance.Equal(cBalance) {
			panic(fmt.Sprintf("appUserContactJson.Balance().Equal(contact.Balance())\nappUser.ContactByID(contact.ID).Balance(): %v\ncontact.Balance(): %v", ucBalance, cBalance))
		}
	}
	return
}

type createContactDbChanges struct {
	db.Changes
	user                *models.AppUser
	counterpartyContact *models.Contact
}

func CreateContact(c context.Context, userID int64, contactDetails models.ContactDetails) (contact models.Contact, user models.AppUser, err error) {
	var contactIDs []int64
	if contactIDs, err = dal.Contact.GetContactIDsByTitle(c, userID, contactDetails.Username, false); err != nil {
		return
	}
	switch len(contactIDs) {
	case 0:
		err = dal.DB.RunInTransaction(c, func(tc context.Context) (err error) {
			if user, err = User.GetUserByID(tc, userID); err != nil {
				return
			}
			changes := &createContactDbChanges{
				user:                &user,
				counterpartyContact: new(models.Contact),
			}
			if contact, _, err = createContactWithinTransaction(tc, changes, 0, contactDetails); err != nil {
				err = errors.WithMessage(err, "failed to create contact within transaction")
				return
			}

			if changes.HasChanges() {
				if err = dal.DB.UpdateMulti(tc, changes.EntityHolders()); err != nil {
					err = errors.WithMessage(err, "failed to save entity related to new contact")
					return
				}
				// TODO: move calls of delays to createContactWithinTransaction() ?
				if err = dal.User.DelayUpdateUserWithContact(tc, userID, contact.ID); err != nil { // Just in case
					return
				}
				if changes.counterpartyContact != nil && changes.counterpartyContact.ID > 0 {
					counterpartyContact := *changes.counterpartyContact
					if err = dal.User.DelayUpdateUserWithContact(tc, counterpartyContact.UserID, counterpartyContact.ID); err != nil { // Just in case
						return
					}
				}
			}
			return
		}, dal.CrossGroupTransaction)
		if err != nil {
			dal.User.DelayUpdateUserWithContact(c, contact.UserID, contact.ID)
			return
		}
		return
	case 1:
		if contact, err = GetContactByID(c, contactIDs[0]); err != nil {
			return
		}
		user.ID = userID
		return
	default:
		err = fmt.Errorf("too many counterparties (%d), IDs: %v", len(contactIDs), contactIDs)
		return
	}
}

func UpdateContact(c context.Context, contactID int64, values map[string]string) (contactEntity *models.ContactEntity, err error) {
	err = dal.DB.RunInTransaction(c, func(c context.Context) error {
		if contact, err := GetContactByID(c, contactID); err != nil {
			return err
		} else {
			contactEntity = contact.ContactEntity
			var changed bool
			for name, value := range values {
				switch name {
				case "Username":
					if contact.Username != value {
						contact.Username = value
						changed = true
					}
				case "FirstName":
					if contact.FirstName != value {
						contact.FirstName = value
						changed = true
					}
				case "LastName":
					if contact.LastName != value {
						contact.LastName = value
						changed = true
					}
				case "ScreenName":
					if contact.ScreenName != value {
						contact.ScreenName = value
						changed = true
					}
				case "EmailAddress":
					if contact.EmailAddressOriginal != value {
						contact.EmailAddressOriginal = value
						changed = true
					}
				case "PhoneNumber":
					if phoneNumber, err := strconv.ParseInt(value, 10, 64); err != nil {
						return err
					} else if contact.PhoneNumber != phoneNumber {
						contact.PhoneNumber = phoneNumber
						contact.PhoneNumberConfirmed = false
						changed = true
					}
				default:
					log.Debugf(c, "Unknown field: %v", name)
				}
			}
			if changed {
				if user, err := User.GetUserByID(c, contact.UserID); err != nil {
					return errors.Wrapf(err, "Failed to get user by ID=%v", contact.UserID)
				} else {
					user.AddOrUpdateContact(contact)
					return dal.DB.UpdateMulti(c, []db.EntityHolder{&contact, &user})
				}
			}
		}
		return nil
	}, dal.CrossGroupTransaction)
	return
}

var ErrContactNotDeletable = errors.New("Contact is not deletable")

func DeleteContact(c context.Context, contactID int64) (user models.AppUser, err error) {
	log.Warningf(c, "ContactDalGae.DeleteContact(%d)", contactID)
	var contact models.Contact
	err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		if contact, err = GetContactByID(c, contactID); err != nil {
			if db.IsNotFound(err) {
				log.Warningf(c, "Contact not found by ID: %v", contactID)
				err = nil
			}
			return
		}
		if contact.ContactEntity != nil && contact.CounterpartyUserID != 0 {
			return ErrContactNotDeletable
		}
		if user, err = User.GetUserByID(c, contact.UserID); err != nil {
			return
		}
		if userContact := user.ContactByID(contactID); userContact != nil {
			userContactBalance := userContact.Balance()
			contactBalance := contact.Balance()
			if !reflect.DeepEqual(userContactBalance, contactBalance) {
				return fmt.Errorf("Data integrity issue: userContactBalance != contactBalance\n\tuserContactBalance: %v\n\tcontactBalance: %v", userContactBalance, contactBalance)
			}
			if !user.RemoveContact(contactID) {
				return errors.New("Implementation error - user not changed on removing contact")
			}
			if contact.BalanceCount > 0 {
				userBalance := user.Balance()
				for k, v := range contactBalance {
					userBalance[k] -= v
				}
				if err = user.SetBalance(userBalance); err != nil {
					return err
				}
			}
			if err = User.SaveUser(c, user); err != nil {
				return err
			}
		}
		if err = dal.DB.Delete(c, &models.Contact{IntegerID: db.IntegerID{ID: contactID}}); err != nil {
			return err
		}
		return nil
	}, dal.CrossGroupTransaction)
	return
}

func SaveContact(c context.Context, contact models.Contact) error {
	return dal.DB.Update(c, &contact)
}

func GetContactsByIDs(c context.Context, contactsIDs []int64) (contacts []models.Contact, err error) {
	count := len(contactsIDs)
	entityHolders := make([]db.EntityHolder, count)
	for i, contactID := range contactsIDs {
		if contactID == 0 {
			panic(fmt.Sprintf("contactsIDs[%d] == 0", i))
		}
		entityHolders[i] = &models.Contact{IntegerID: db.NewIntID(contactID)}
	}
	if err = dal.DB.GetMulti(c, entityHolders); err != nil {
		return
	}
	contacts = make([]models.Contact, count)
	for i, entityHolder := range entityHolders {
		contacts[i] = *(entityHolder.(*models.Contact))
	}
	return contacts, err
}

func GetContactByID(c context.Context, contactID int64) (contact models.Contact, err error) {
	contact.ID = contactID
	err = dal.DB.Get(c, &contact)
	return
}
