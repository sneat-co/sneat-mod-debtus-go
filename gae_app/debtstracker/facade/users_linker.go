package facade

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"golang.org/x/net/context"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"github.com/pkg/errors"
	"github.com/strongo/app/log"
	"fmt"
	"github.com/strongo/app/db"
)

type UsersLinker struct {
	// Groups methods for linking 2 users via Contact
}

func (l UsersLinker) LinkUsersWithinTransaction(
	c, tc context.Context, // 'tc' is transactional context, 'c' is not
	inviterUser, invitedUser models.AppUser,
	inviterContact models.Contact,
) (
	entitiesToSave [] db.EntityHolder,
	invitedContact models.Contact,
	err error,
) {
	log.Debugf(c, "UsersLinker.LinkUsersWithinTransaction(inviterUser.ID=%d, invitedUser.ID=%d, inviterContact=%d)", inviterUser.ID, invitedUser.ID, inviterContact.ID)
	// First of all lets validate input
	if err = l.validateInput(inviterUser, invitedUser, inviterContact); err != nil {
		return
	}

	// Update entities
	{
		var inviterContactChanged, invitedUserChanged bool
		if invitedContact, inviterContactChanged, err = l.getOrCreateInvitedContactByInviterUserAndInviterContact(c, tc, invitedUser, inviterUser, inviterContact); err != nil {
			return
		}
		if invitedContact.ContactEntity == nil {
			err = errors.New(
				fmt.Sprintf(
					"getOrCreateInvitedContactByInviterUserAndInviterContact() returned invitedContact.ContactEntity == nil, invitedContact.ID: %d",
					invitedContact.ID,
				))
			return
		}
		if invitedUserChanged, err = l.updateInvitedUser(invitedUser, inviterUser.ID, inviterContact); err != nil {
			return
		} else if invitedUserChanged {
			entitiesToSave = toSave(entitiesToSave, &invitedUser)
		}

		var isJustConnected bool
		if isJustConnected, inviterContactChanged, err = l.updateInviterContact(tc, inviterUser, invitedUser, &inviterContact, &invitedContact); err != nil {
			return
		}
		if inviterContactChanged {
			log.Debugf(tc, "isJustConnected: %v", isJustConnected)
			entitiesToSave = toSave(entitiesToSave, &inviterContact)
		}
	}
	return
}

func (l UsersLinker) validateInput(
	inviterUser, invitedUser models.AppUser,
	inviterContact models.Contact,
) error {
	if inviterUser.ID == 0 {
		panic("inviterUser.ID == 0")
	}
	if invitedUser.ID == 0 {
		panic("invitedUser.ID == 0")
	}
	if inviterContact.ID == 0 {
		panic("inviterContact.ID == 0")
	}

	return nil
}

// Purpose of the function is an attempt to link existing counterparties
func (_ UsersLinker) getOrCreateInvitedContactByInviterUserAndInviterContact(
	c, tc context.Context,
	invitedUser models.AppUser,
	inviterUser models.AppUser,
	inviterContact models.Contact,
) (
	invitedContact models.Contact,
	invitedContactChanged bool,
	err error,
) { // TODO: Can this be re-used for invites as well?
	log.Debugf(c, "getOrCreateInvitedContactByInviterUserAndInviterContact()")
	if invitedUser.ContactsCount > 0 {
		var invitedUserContacts []models.Contact
		// Use non transaction context
		invitedUserContacts, err = dal.Contact.GetContactsByIDs(c, invitedUser.CounterpartiesIDs())
		if err != nil {
			err = errors.Wrap(err, "Failed to call dal.Contact.GetContactsByIDs()")
			return
		}
		for _, invitedUserContact := range invitedUserContacts {
			if invitedUserContact.CounterpartyUserID == inviterUser.ID {
				// We re-get the entity of the found invitedContact using transactional context
				// and store it to output var
				if invitedContact, err = dal.Contact.GetContactByID(tc, invitedUserContact.ID); err != nil {
					err = errors.Wrapf(err, "Failed to call dal.Contact.GetContactByID(%d)", invitedUserContact.ID)
					return
				}
				if invitedContact.FirstName == "" {
					invitedContact.FirstName = inviterUser.FirstName
				}
				if invitedContact.LastName == "" {
					invitedContact.LastName = inviterUser.LastName
				}
				break
			}
		}
	}

	if invitedContact.ID == 0 {
		log.Debugf(c, "getOrCreateInvitedContactByInviterUserAndInviterContact(): creating new contact for invited user")
		var invitedCounterpartyBalance models.Balance
		if invitedCounterpartyBalance, err = models.ReverseBalance(inviterContact.Balanced); err != nil {
			return
		}
		balanced := models.Balanced{
			CountOfTransfers: inviterContact.CountOfTransfers,
			LastTransferID:   inviterContact.LastTransferID,
			LastTransferAt:   inviterContact.LastTransferAt,
		}
		balanced.SetBalance(invitedCounterpartyBalance)
		invitedContacts := models.ContactDetails{
			FirstName: inviterUser.FirstName,
			LastName: inviterUser.LastName,
			Nickname: inviterUser.Nickname,
			ScreenName: inviterUser.ScreenName,
			Username: inviterUser.Username,
		}
		if invitedContact, err = CreateContactWithinTransaction(
			tc, invitedUser, inviterUser.ID, inviterContact.ID, invitedContacts, balanced,
		); err != nil {
			return
		}
		log.Debugf(c, "getOrCreateInvitedContactByInviterUserAndInviterContact(): created contact: %v", invitedContact)
	} else {
		log.Debugf(c, "getOrCreateInvitedContactByInviterUserAndInviterContact(): linking existing contact: %v", invitedContact)
		// TODO: How do we merge existing contacts?
		invitedContact.CountOfTransfers = inviterContact.CountOfTransfers
		invitedContact.LastTransferID = inviterContact.LastTransferID
		invitedContact.LastTransferAt = inviterContact.LastTransferAt
		var creatorCounterpartyBalance models.Balance
		creatorCounterpartyBalance, err = inviterContact.Balance()
		invitedContact.SetBalance(creatorCounterpartyBalance)
		invitedContactChanged = true
	}
	return
}

func (_ UsersLinker) updateInvitedUser(invitedUser models.AppUser, inviterUserID int64, inviterContact models.Contact) (invitedUserChanged bool, err error) {
	if invitedUser.InvitedByUserID == 0 {
		invitedUser.InvitedByUserID = inviterUserID
		invitedUserChanged = true
	}
	var inviterContactBalance models.Balance
	if inviterContactBalance, err = inviterContact.Balance(); err != nil {
		err = errors.Wrap(err, "Failed to get inviterContact.Balance()")
		return
	} else if len(inviterContactBalance) > 0 {
		for currency, value := range inviterContactBalance {
			invitedUser.Add2Balance(currency, -1*value)
		}
		invitedUserChanged = true
	}
	if inviterContact.LastTransferAt.After(invitedUser.LastTransferAt) {
		invitedUser.LastTransferID = inviterContact.LastTransferID
		invitedUser.LastTransferAt = inviterContact.LastTransferAt
		invitedUserChanged = true
	}
	return
}

// Updates counterparty entity that belongs to inviter user (inviterContact.UserID == inviterUser.ID)
func (_ UsersLinker) updateInviterContact(
	tc context.Context,
	inviterUser, invitedUser models.AppUser,
	inviterContact, invitedContact *models.Contact,
) (
	isJustConnected, inviterContactChange bool, err error,
) {
	log.Debugf(tc, "UsersLinker.updateInviterContact(), inviterContact.CounterpartyUserID: %d, inviterContact.CountOfTransfers: %d", inviterContact.CounterpartyUserID, inviterContact.CountOfTransfers)
	if inviterUser.ID == 0 {
		panic("inviterUser.ID == 0")
	}
	if invitedUser.ID == 0 {
		panic("invitedUser.ID == 0")
	}
	if inviterContact.UserID != inviterUser.ID {
		panic("inviterContact.UserID != inviterUser.ID")
	}

	if inviterContact.FirstName == "" {
		inviterContact.FirstName = invitedUser.FirstName
		inviterContactChange = true
	}
	if inviterContact.LastName == "" {
		inviterContact.LastName = invitedUser.LastName
		inviterContactChange = true
	}
	//if inviterContactChange {
	//	inviterContact.UpdateSearchName()
	//}
	switch inviterContact.CounterpartyUserID {
	case 0:
		log.Debugf(tc, "Updating inviterUser.Contact* fields...")
		isJustConnected = true
		inviterContactChange = true
		inviterContact.CounterpartyUserID = invitedUser.ID
		inviterContact.CounterpartyCounterpartyID = invitedContact.ID
		// Queue task to update all existing transfers
		if inviterContact.CountOfTransfers > 0 {
			if err = dal.Transfer.DelayUpdateTransfersWithCounterparty(
				tc,
				invitedContact.ID,
				inviterContact.ID,
			); err != nil {
				err = errors.Wrap(err, "Failed to enqueue delayUpdateTransfersWithCounterparty()")
				return
			}
		} else {
			log.Debugf(tc, "No need to update transfers of inviter as inviterContact.CountOfTransfers == 0")
		}
	case invitedUser.ID:
		log.Infof(tc, "inviterContact.CounterpartyUserID already set")
	default:
		err = fmt.Errorf("inviterContact.CounterpartyUserID is different from current user. inviterContact.CounterpartyUserID: %v, currentUserID: %v", inviterContact.CounterpartyUserID, invitedUser.ID)
		return
	}
	return
}
