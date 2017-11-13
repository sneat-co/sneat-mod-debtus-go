package facade

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/pkg/errors"
	"github.com/strongo/app/db"
	"github.com/strongo/app/log"
	"golang.org/x/net/context"
	"reflect"
	"strconv"
	"github.com/sanity-io/litter"
)

func ChangeContactStatus(c context.Context, contactID int64, newStatus string) (contact models.Contact, err error) {
	err = dal.DB.RunInTransaction(c, func(c context.Context) error {
		if contact, err = dal.Contact.GetContactByID(c, contactID); err != nil {
			return err
		} else if contact.Status != newStatus {
			contact.Status = newStatus
			user, err := dal.User.GetUserByID(c, contact.UserID)
			if err != nil {
				return err
			}
			if userChanged := user.AddOrUpdateContact(contact); userChanged {
				err = dal.DB.UpdateMulti(c, []db.EntityHolder{&contact, &user})
			} else {
				err = dal.Contact.SaveContact(c, contact)
			}
			return err
		}
		return nil
	}, dal.CrossGroupTransaction)
	return
}

func CreateContactWithinTransaction(
	tc context.Context,
	appUser models.AppUser,
	counterpartyUserID int64,
	counterpartyContact models.Contact,
	contactDetails models.ContactDetails,
) (
	contact models.Contact,
	counterpartyContactOutput models.Contact,
	err error,
) {
	log.Debugf(tc, "CreateContactWithinTransaction(appUser.ID=%v, counterpartyDetails=%v)", appUser.ID, contactDetails)
	if !dal.DB.IsInTransaction(tc) {
		err = errors.New("CreateContactWithinTransaction is called outside of transaction")
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
			if counterpartyContact, err = dal.Contact.GetContactByID(tc, counterpartyContact.ID); err != nil {
				return
			}
		}
		if counterpartyContact.UserID != counterpartyUserID {
			if counterpartyUserID == 0 {
				counterpartyUserID = counterpartyContact.UserID
			} else {
				panic(fmt.Sprintf("counterpartyContact.UserID != counterpartyUserID: %v != %v", counterpartyContact.UserID, counterpartyUserID))
			}
		}
		contact.ContactEntity.CounterpartyUserID = counterpartyUserID
		contact.ContactEntity.CounterpartyCounterpartyID = counterpartyContact.ID
		contact.ContactEntity.TransfersJson = counterpartyContact.TransfersJson
		counterpartyContactOutput = counterpartyContact
		contact.ContactEntity.Balanced = models.Balanced{
			CountOfTransfers: counterpartyContact.CountOfTransfers,
			LastTransferID:   counterpartyContact.LastTransferID,
			LastTransferAt:   counterpartyContact.LastTransferAt,
		}
		invitedCounterpartyBalance := models.ReverseBalance(counterpartyContact.Balance())
		log.Debugf(tc, "invitedCounterpartyBalance: %v", invitedCounterpartyBalance)
		contact.SetBalance(invitedCounterpartyBalance)
		if contact.BalanceCount != counterpartyContact.BalanceCount {
			panic(fmt.Sprintf("contact.BalanceCount != counterpartyContact.BalanceCount:  %v != %v", contact.BalanceCount, counterpartyContact.BalanceCount))
		}
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
			if err = dal.Contact.SaveContact(tc, counterpartyContact); err != nil {
				return
			}
		} else if counterpartyContact.CounterpartyCounterpartyID != contact.ID {
			err = fmt.Errorf("inviter contact %v already has CounterpartyCounterpartyID=%v", counterpartyContact.ID, counterpartyContact.CounterpartyCounterpartyID)
			return
		}
	}

	appUser.AddOrUpdateContact(contact)

	{ // Verifications for data integrity
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
			if cBalance, cpBalance := contact.Balance(), counterpartyContact.Balance(); !cBalance.Equal(models.ReverseBalance(cpBalance)) {
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

func CreateContact(c context.Context, userID int64, contactDetails models.ContactDetails) (counterparty models.Contact, user models.AppUser, err error) {
	var contactIDs []int64
	if contactIDs, err = dal.Contact.GetContactIDsByTitle(c, userID, contactDetails.Username, false); err != nil {
		return
	}
	switch len(contactIDs) {
	case 0:
		err = dal.DB.RunInTransaction(c, func(tc context.Context) (err error) {
			if user, err = dal.User.GetUserByID(tc, userID); err != nil {
				return
			}
			if counterparty, _, err = CreateContactWithinTransaction(tc, user, 0, models.Contact{}, contactDetails); err != nil {
				err = errors.Wrap(err, "Failed to create counterparty within transaction")
				return
			}
			if err = dal.User.SaveUser(tc, user); err != nil {
				err = errors.Wrap(err, "Failed to save user entity to DB")
				return
			}
			return
		}, dal.CrossGroupTransaction)
		return
	case 1:
		if counterparty, err = dal.Contact.GetContactByID(c, contactIDs[0]); err != nil {
			return
		}
		user.ID = userID
		return
	default:
		err = errors.New(fmt.Sprintf("Too many counterparties (%d), IDs: %v", len(contactIDs), contactIDs))
		return
	}
}

func UpdateContact(c context.Context, contactID int64, values map[string]string) (contactEntity *models.ContactEntity, err error) {
	err = dal.DB.RunInTransaction(c, func(c context.Context) error {
		if contact, err := dal.Contact.GetContactByID(c, contactID); err != nil {
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
				if user, err := dal.User.GetUserByID(c, contact.UserID); err != nil {
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
		if contact, err = dal.Contact.GetContactByID(c, contactID); err != nil && !db.IsNotFound(err) {
			return
		}
		if contact.ContactEntity != nil && contact.CounterpartyUserID != 0 {
			return ErrContactNotDeletable
		}
		if user, err = dal.User.GetUserByID(c, contact.UserID); err != nil {
			return
		}
		if userContact := user.ContactByID(contactID); userContact != nil {
			userContactBalance := userContact.Balance()
			contactBalance := contact.Balance()
			if !reflect.DeepEqual(userContactBalance, contactBalance) {
				return errors.New(fmt.Sprintf("Data integrity issue: userContactBalance != contactBalance\n\tuserContactBalance: %v\n\tcontactBalance: %v", userContactBalance, contactBalance))
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
			if err = dal.User.SaveUser(c, user); err != nil {
				return err
			}
		}
		if err = dal.Contact.DeleteContact(c, contactID); err != nil {
			return err
		}
		return nil
	}, dal.CrossGroupTransaction)
	return
}
