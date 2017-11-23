package facade

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/pkg/errors"
	"github.com/strongo/log"
	"golang.org/x/net/context"
)

type usersLinker struct {
	// Groups methods for linking 2 users via Contact
	changes *usersLinkingDbChanges
}

func newUsersLinker(changes *usersLinkingDbChanges) usersLinker {
	return usersLinker{
		changes: changes,
	}
}

func (linker usersLinker) linkUsersWithinTransaction(
	tc context.Context, // 'tc' is transactional context, 'c' is not
) (
	err error,
) {
	changes := linker.changes
	inviterUser, invitedUser := changes.inviterUser, changes.invitedUser
	inviterContact, invitedContact := changes.inviterContact, changes.invitedContact
	if invitedContact == nil {
		invitedContact = new(models.Contact)
		changes.invitedContact = invitedContact
	}

	log.Debugf(tc, "usersLinker.linkUsersWithinTransaction(inviterUser.ID=%d, invitedUser.ID=%d, inviterContact=%d, inviterContact.UserID=%v)", inviterUser.ID, invitedUser.ID, inviterContact.ID, inviterContact.UserID)

	// First of all lets validate input
	if err = linker.validateInput(inviterUser, invitedUser, inviterContact); err != nil {
		return
	}

	if !dal.DB.IsInTransaction(tc) {
		err = errors.New("usersLinker.linkUsersWithinTransaction is called outside of transaction")
		return
	}

	// Update entities
	{
		if err = linker.getOrCreateInvitedContactByInviterUserAndInviterContact(tc, changes); err != nil {
			return
		} else {
			invitedContact = changes.invitedContact
		}

		if invitedContact.ContactEntity == nil {
			err = fmt.Errorf(
				"getOrCreateInvitedContactByInviterUserAndInviterContact() returned invitedContact.ContactEntity == nil, invitedContact.ID: %d",
				invitedContact.ID)
			return
		} else if invitedContact.UserID != invitedUser.ID {
			panic(fmt.Sprintf("invitedContact.UserID != invitedUser.ID: %v != %v", invitedContact.UserID, invitedUser.ID))
		}

		log.Debugf(tc, "getOrCreateInvitedContactByInviterUserAndInviterContact() => invitedContact.ID: %v", invitedContact.ID)

		if err = linker.updateInvitedUser(tc, *invitedUser, inviterUser.ID, *inviterContact); err != nil {
			return
		}

		if _, err = linker.updateInviterContact(tc, *inviterUser, *invitedUser, inviterContact, invitedContact); err != nil {
			return
		}
	}

	// verify
	{
		invitedContact.MustMatchCounterparty(*inviterContact)
		if !invitedUser.ContactByID(invitedContact.ID).Balance().Equal(inviterUser.ContactByID(inviterContact.ID).Balance().Reversed()) {
			panic(fmt.Sprintf("users contacts json balances are not equal (invited vs inviter): %v != %v",
				invitedUser.ContactByID(invitedContact.ID).Balance(),
				inviterUser.ContactByID(inviterContact.ID).Balance(),
			))
		}
	}
	return
}

func (linker usersLinker) validateInput(
	inviterUser, invitedUser *models.AppUser,
	inviterContact *models.Contact,
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
	if inviterUser.ID == invitedUser.ID {
		panic(fmt.Sprintf("inviterUser.ID == invitedUser.ID: %v", inviterUser.ID))
	}
	if inviterContact.UserID != inviterUser.ID {
		panic(fmt.Sprintf("usersLinker.validateInput(): inviterContact.UserID != inviterUser.ID: %v != %v", inviterContact.UserID, inviterUser.ID))
	}
	return nil
}

// Purpose of the function is an attempt to link existing counterparties
func (linker usersLinker) getOrCreateInvitedContactByInviterUserAndInviterContact(
	tc context.Context, changes *usersLinkingDbChanges,
) (err error) {
	inviterUser, invitedUser := *changes.inviterUser, *changes.invitedUser
	inviterContact := *changes.inviterContact
	log.Debugf(tc, "getOrCreateInvitedContactByInviterUserAndInviterContact()\n\tinviterContact.ID: %v", inviterContact.ID)
	if inviterUser.ID == invitedUser.ID {
		panic(fmt.Sprintf("inviterUser.ID == invitedUser.ID: %v", inviterUser.ID))
	}

	var invitedContact models.Contact
	if changes.invitedContact != nil && changes.invitedContact.ID != 0 {
		invitedContact = *changes.invitedContact
	} else {
		changes.invitedContact = &invitedContact
	}

	if invitedUser.ContactsCount > 0 {
		var invitedUserContacts []models.Contact
		// Use non transaction context
		invitedUserContacts, err = dal.Contact.GetContactsByIDs(tc, invitedUser.ContactIDs())
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
		log.Debugf(tc, "getOrCreateInvitedContactByInviterUserAndInviterContact(): creating new contact for invited user")
		invitedContactDetails := models.ContactDetails{
			FirstName:  inviterUser.FirstName,
			LastName:   inviterUser.LastName,
			Nickname:   inviterUser.Nickname,
			ScreenName: inviterUser.ScreenName,
			Username:   inviterUser.Username,
		}
		createContactDbChanges := &createContactDbChanges{
			user:                changes.invitedUser,
			counterpartyContact: changes.inviterContact,
		}
		if invitedContact, inviterContact, err = createContactWithinTransaction(tc, createContactDbChanges, inviterUser.ID, invitedContactDetails); err != nil {
			return
		}
		if changes.inviterContact == nil {
			changes.inviterContact = &inviterContact
			changes.FlagAsChanged(changes.inviterContact)
		}
		if invitedUser.LastTransferAt.Before(inviterContact.LastTransferAt) {
			invitedUser.LastTransferID = inviterContact.LastTransferID
			invitedUser.LastTransferAt = inviterContact.LastTransferAt
			changes.FlagAsChanged(changes.invitedUser)
		}
	} else {
		log.Debugf(tc, "getOrCreateInvitedContactByInviterUserAndInviterContact(): linking existing contact: %v", invitedContact)
		// TODO: How do we merge existing contacts?
		invitedContact.CountOfTransfers = inviterContact.CountOfTransfers
		invitedContact.LastTransferID = inviterContact.LastTransferID
		invitedContact.LastTransferAt = inviterContact.LastTransferAt
		if err = invitedContact.SetBalance(inviterContact.Balance().Reversed()); err != nil {
			return
		}
		changes.FlagAsChanged(changes.invitedContact)
	}
	invitedContact.MustMatchCounterparty(inviterContact)
	return
}

func (linker usersLinker) updateInvitedUser(c context.Context, invitedUser models.AppUser, inviterUserID int64, inviterContact models.Contact) (err error) {
	log.Debugf(c, "usersLinker.updateInvitedUser()")
	var invitedUserChanged bool

	if invitedUser.InvitedByUserID == 0 {
		invitedUser.InvitedByUserID = inviterUserID
		invitedUserChanged = true
	}

	if inviterContact.LastTransferAt.After(invitedUser.LastTransferAt) {
		invitedUser.LastTransferID = inviterContact.LastTransferID
		invitedUser.LastTransferAt = inviterContact.LastTransferAt
		invitedUserChanged = true
	}

	if invitedUserChanged {
		linker.changes.FlagAsChanged(linker.changes.invitedUser)
	}
	return
}

// Updates counterparty entity that belongs to inviter user (inviterContact.UserID == inviterUser.ID)
func (linker usersLinker) updateInviterContact(
	tc context.Context,
	inviterUser, invitedUser models.AppUser,
	inviterContact, invitedContact *models.Contact,
) (
	isJustConnected bool, err error,
) {
	log.Debugf(tc, "usersLinker.updateInviterContact(), inviterContact.CounterpartyUserID: %d, inviterContact.CountOfTransfers: %d", inviterContact.CounterpartyUserID, inviterContact.CountOfTransfers)
	// validate input
	{
		if inviterUser.ID == 0 {
			panic("inviterUser.ID == 0")
		}
		if invitedUser.ID == 0 {
			panic("invitedUser.ID == 0")
		}
		if inviterContact.UserID != inviterUser.ID {
			panic(fmt.Sprintf("usersLinker.updateInviterContact(): inviterContact.UserID != inviterUser.ID: %v != %v\ninvitedContact.UserID: %v, invitedUser.ID: %v",
				inviterContact.UserID, inviterUser.ID, invitedContact.UserID, invitedUser.ID))
		}
		if invitedContact.UserID != invitedUser.ID {
			panic(fmt.Sprintf("invitedContact.UserID != invitedUser.ID: %v != %v\ninviterContact.UserID: %v, inviterUser.ID: %v",
				invitedContact.UserID, invitedContact.ID, inviterContact.UserID, inviterUser.ID))
		}
		if invitedContact.ID == inviterContact.ID {
			panic(fmt.Sprintf("invitedContact.ID == inviterContact.ID: %v", invitedContact.ID))
		}
		if invitedUser.ID == inviterUser.ID {
			panic(fmt.Sprintf("invitedUser.ID == inviterUser.ID: %v", invitedUser.ID))
		}
	}
	var inviterContactChanged bool
	if inviterContact.FirstName == "" {
		inviterContact.FirstName = invitedUser.FirstName
		inviterContactChanged = true
	}
	if inviterContact.LastName == "" {
		inviterContact.LastName = invitedUser.LastName
		inviterContactChanged = true
	}
	//if inviterContactChanged {
	//	inviterContact.UpdateSearchName()
	//}
	if inviterContactChanged {
		linker.changes.FlagAsChanged(linker.changes.inviterContact)
	} else {
		defer func() {
			if inviterContactChanged {
				linker.changes.FlagAsChanged(linker.changes.inviterContact)
			}
		}()
	}
	switch inviterContact.CounterpartyUserID {
	case 0:
		log.Debugf(tc, "Updating inviterUser.Contact* fields...")
		isJustConnected = true
		inviterContactChanged = true
		inviterContact.CounterpartyUserID = invitedUser.ID
		inviterContact.CounterpartyCounterpartyID = invitedContact.ID
		inviterUserContacts := inviterUser.Contacts()
		for i, inviterUserContact := range inviterUserContacts {
			if inviterUserContact.ID == inviterContact.ID {
				if inviterUserContact.UserID == 0 {
					inviterUserContact.UserID = inviterContact.CounterpartyUserID
					inviterUserContacts[i] = inviterUserContact
					inviterUser.SetContacts(inviterUserContacts)
					linker.changes.FlagAsChanged(linker.changes.inviterUser)
				} else if inviterUserContact.UserID == inviterContact.CounterpartyUserID {
					// do nothing
				} else {
					err = fmt.Errorf(
						"data integrity issue for contact %v: inviterUserContact.UserID != inviterContact.CounterpartyUserID: %v != %v",
						inviterContact.ID, inviterUserContact.UserID, inviterContact.CounterpartyUserID)
					return
				}
				goto inviterUserContactFound
			}
		}
		if inviterUser.AddOrUpdateContact(*inviterContact) {
			linker.changes.FlagAsChanged(linker.changes.inviterUser)
		}
		inviterUserContactFound:
		// Queue task to update all existing transfers
		if inviterContact.CountOfTransfers > 0 {
			if err = dal.Transfer.DelayUpdateTransfersWithCounterparty(
				tc,
				invitedContact.ID,
				inviterContact.ID,
			); err != nil {
				err = errors.WithMessage(err, "Failed to enqueue delayUpdateTransfersWithCounterparty()")
				return
			}
		} else {
			log.Debugf(tc, "No need to update transfers of inviter as inviterContact.CountOfTransfers == 0")
		}
	case invitedUser.ID:
		log.Infof(tc, "inviterContact.CounterpartyUserID is already set, updateInviterContact() did nothing")
	default:
		err = fmt.Errorf("inviterContact.CounterpartyUserID is different from current user. inviterContact.CounterpartyUserID: %v, currentUserID: %v", inviterContact.CounterpartyUserID, invitedUser.ID)
		return
	}
	return
}
