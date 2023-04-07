package facade

import (
	"testing"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtmocks"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
)

func TestUsersLinker_LinkUsersWithinTransaction(t *testing.T) {
	c := context.Background()
	dtmocks.SetupMocks(c)

	usersLinker := usersLinker{}

	var (
		err                            error
		inviterUser, invitedUser       models.AppUser
		inviterContact, invitedContact models.Contact
	)

	if inviterUser, err = User.GetUserByID(c, 1); err != nil {
		t.Error("Failed to get inviter user", err)
		return
	}

	if invitedUser, err = User.GetUserByID(c, 3); err != nil {
		t.Error("Failed to get invited user", err)
		return
	}

	if inviterContact, err = GetContactByID(c, 6); err != nil {
		t.Error("Failed to get inviter user", err)
		return
	}

	if inviterContact.CounterpartyUserID != 0 {
		t.Error("inviterContact.CounterpartyUserID != 0")
	}

	if inviterContact.CounterpartyCounterpartyID != 0 {
		t.Error("inviterContact.CounterpartyCounterpartyID != 0")
	}

	err = dtdal.DB.RunInTransaction(c, func(tc context.Context) (err error) {
		usersLinker = newUsersLinker(&usersLinkingDbChanges{
			inviterUser:    &inviterUser,
			invitedUser:    &invitedUser,
			inviterContact: &inviterContact,
			invitedContact: &invitedContact,
		})
		if err = usersLinker.linkUsersWithinTransaction(tc, "unit-test:1"); err != nil {
			return err
		}
		return nil
	}, dtdal.CrossGroupTransaction)

	if err != nil {
		t.Error("Unexpected error:", err)
		return
	}

	if len(usersLinker.changes.EntityHolders()) == 0 {
		t.Error("len(usersLinker.changes.EntityHolders()) == 0")
		return
	}

	invitedContact = *usersLinker.changes.invitedContact
	inviterContact = *usersLinker.changes.inviterContact
	invitedUser = *usersLinker.changes.invitedUser
	inviterUser = *usersLinker.changes.inviterUser

	if invitedContact.ID == 0 {
		t.Error("invitedContact.ID == 0")
		return
	}

	if invitedContact.ID == inviterContact.ID {
		t.Errorf("invitedContact.ID == inviterContact.ID: %d", invitedContact.ID)
	}

	if invitedContact.ContactEntity == nil {
		t.Error("invitedContact.ContactEntity == nil")
		return
	}

	if invitedContact.UserID == 0 {
		t.Error("invitedContact.UserID == 0")
		return
	}

	if invitedContact.UserID != invitedUser.ID {
		t.Errorf("invitedContact.UserID == invitedUser.ID : %d != %d", invitedContact.UserID, invitedUser.ID)
		return
	}

	if invitedContact.CounterpartyUserID == 0 {
		t.Error("invitedContact.CounterpartyUserID == 0")
		return
	}

	if invitedContact.CounterpartyCounterpartyID == 0 {
		t.Error("invitedContact.CounterpartyCounterpartyID == 0")
		return
	}

	if invitedContact.CounterpartyUserID != inviterUser.ID {
		t.Errorf("invitedContact.CounterpartyUserID != inviterUser.ID : %d != %d", invitedContact.CounterpartyUserID, inviterUser.ID)
		return
	}

	if invitedContact.CounterpartyCounterpartyID != inviterContact.ID {
		t.Errorf("invitedContact.CounterpartyCounterpartyID != inviterContact.ID : %d != %d", invitedContact.CounterpartyCounterpartyID, inviterContact.ID)
		return
	}

	if inviterContact.CounterpartyUserID == 0 {
		t.Error("inviterContact.CounterpartyUserID == 0")
		return
	}

	if inviterContact.CounterpartyCounterpartyID == 0 {
		t.Error("inviterContact.CounterpartyCounterpartyID == 0")
		return
	}

	if inviterContact.CounterpartyUserID != invitedUser.ID {
		t.Errorf("inviterContact.CounterpartyUserID != invitedUser.ID : %d != %d", inviterContact.CounterpartyUserID, invitedUser.ID)
		return
	}

	if inviterContact.CounterpartyCounterpartyID != invitedContact.ID {
		t.Errorf("inviterContact.CounterpartyCounterpartyID != invitedContact.ID : %d != %d", inviterContact.CounterpartyCounterpartyID, invitedContact.ID)
		return
	}

	if invitedContact.Username != "" && invitedContact.Username == inviterContact.Username {
		t.Errorf("invitedContact.Username == inviterContact.Username: %v", invitedContact.Username)
		return
	}

	if invitedContact.FirstName != "" && invitedContact.FirstName == inviterContact.FirstName {
		t.Errorf("invitedContact.FirstName == inviterContact.FirstName: %v", invitedContact.FirstName)
		return
	}

	if invitedContact.LastName != "" && invitedContact.LastName == inviterContact.LastName {
		t.Errorf("invitedContact.LastName == inviterContact.LastName: %v", invitedContact.LastName)
		return
	}

	if invitedContact.Nickname != "" && invitedContact.Nickname == inviterContact.Nickname {
		t.Errorf("invitedContact.Nickname == inviterContact.Nickname: %v", invitedContact.Nickname)
		return
	}

	if invitedContact.ScreenName != "" && invitedContact.ScreenName == inviterContact.ScreenName {
		t.Errorf("invitedContact.ScreenName == inviterContact.ScreenName: %v", invitedContact.ScreenName)
		return
	}

	var isInvitedUserHasInvitedContact bool

	for _, invitedUserContact := range invitedUser.Contacts() {
		if invitedUserContact.ID == invitedContact.ID {
			isInvitedUserHasInvitedContact = true
			break
		}
	}

	if !isInvitedUserHasInvitedContact {
		t.Error("Invited user missing invited contact in the CounterpartiesJson")
		return
	}
}
