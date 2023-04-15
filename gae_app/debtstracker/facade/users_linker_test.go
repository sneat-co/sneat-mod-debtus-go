package facade

import (
	"github.com/dal-go/dalgo/dal"
	"testing"

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

	if inviterUser, err = User.GetUserByID(c, nil, 1); err != nil {
		t.Error("Failed to get inviter user", err)
		return
	}

	if invitedUser, err = User.GetUserByID(c, nil, 3); err != nil {
		t.Error("Failed to get invited user", err)
		return
	}

	if inviterContact, err = GetContactByID(c, nil, 6); err != nil {
		t.Error("Failed to get inviter user", err)
		return
	}

	if inviterContact.Data.CounterpartyUserID != 0 {
		t.Error("inviterContact.CounterpartyUserID != 0")
	}

	if inviterContact.Data.CounterpartyCounterpartyID != 0 {
		t.Error("inviterContact.CounterpartyCounterpartyID != 0")
	}

	var db dal.Database
	if db, err = GetDatabase(c); err != nil {
		return
	}
	err = db.RunReadwriteTransaction(c, func(tc context.Context, tx dal.ReadwriteTransaction) (err error) {
		usersLinker = newUsersLinker(&usersLinkingDbChanges{
			inviterUser:    &inviterUser,
			invitedUser:    &invitedUser,
			inviterContact: &inviterContact,
			invitedContact: &invitedContact,
		})
		if err = usersLinker.linkUsersWithinTransaction(tc, tx, "unit-test:1"); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		t.Error("Unexpected error:", err)
		return
	}

	if len(usersLinker.changes.Records()) == 0 {
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

	if invitedContact.Data == nil {
		t.Error("invitedContact.ContactData == nil")
		return
	}

	if invitedContact.Data.UserID == 0 {
		t.Error("invitedContact.UserID == 0")
		return
	}

	if invitedContact.Data.UserID != invitedUser.ID {
		t.Errorf("invitedContact.UserID == invitedUser.ID : %d != %d", invitedContact.Data.UserID, invitedUser.ID)
		return
	}

	if invitedContact.Data.CounterpartyUserID == 0 {
		t.Error("invitedContact.CounterpartyUserID == 0")
		return
	}

	if invitedContact.Data.CounterpartyCounterpartyID == 0 {
		t.Error("invitedContact.CounterpartyCounterpartyID == 0")
		return
	}

	if invitedContact.Data.CounterpartyUserID != inviterUser.ID {
		t.Errorf("invitedContact.CounterpartyUserID != inviterUser.ID : %d != %d", invitedContact.Data.CounterpartyUserID, inviterUser.ID)
		return
	}

	if invitedContact.Data.CounterpartyCounterpartyID != inviterContact.ID {
		t.Errorf("invitedContact.CounterpartyCounterpartyID != inviterContact.ID : %d != %d", invitedContact.Data.CounterpartyCounterpartyID, inviterContact.ID)
		return
	}

	if inviterContact.Data.CounterpartyUserID == 0 {
		t.Error("inviterContact.CounterpartyUserID == 0")
		return
	}

	if inviterContact.Data.CounterpartyCounterpartyID == 0 {
		t.Error("inviterContact.CounterpartyCounterpartyID == 0")
		return
	}

	if inviterContact.Data.CounterpartyUserID != invitedUser.ID {
		t.Errorf("inviterContact.CounterpartyUserID != invitedUser.ID : %d != %d", inviterContact.Data.CounterpartyUserID, invitedUser.ID)
		return
	}

	if inviterContact.Data.CounterpartyCounterpartyID != invitedContact.ID {
		t.Errorf("inviterContact.CounterpartyCounterpartyID != invitedContact.ID : %d != %d", inviterContact.Data.CounterpartyCounterpartyID, invitedContact.ID)
		return
	}

	if invitedContact.Data.Username != "" && invitedContact.Data.Username == inviterContact.Data.Username {
		t.Errorf("invitedContact.Username == inviterContact.Username: %v", invitedContact.Data.Username)
		return
	}

	if invitedContact.Data.FirstName != "" && invitedContact.Data.FirstName == inviterContact.Data.FirstName {
		t.Errorf("invitedContact.FirstName == inviterContact.FirstName: %v", invitedContact.Data.FirstName)
		return
	}

	if invitedContact.Data.LastName != "" && invitedContact.Data.LastName == inviterContact.Data.LastName {
		t.Errorf("invitedContact.LastName == inviterContact.LastName: %v", invitedContact.Data.LastName)
		return
	}

	if invitedContact.Data.Nickname != "" && invitedContact.Data.Nickname == inviterContact.Data.Nickname {
		t.Errorf("invitedContact.Nickname == inviterContact.Nickname: %v", invitedContact.Data.Nickname)
		return
	}

	if invitedContact.Data.ScreenName != "" && invitedContact.Data.ScreenName == inviterContact.Data.ScreenName {
		t.Errorf("invitedContact.ScreenName == inviterContact.ScreenName: %v", invitedContact.Data.ScreenName)
		return
	}

	var isInvitedUserHasInvitedContact bool

	for _, invitedUserContact := range invitedUser.Data.Contacts() {
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
