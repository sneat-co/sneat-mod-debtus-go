package facade

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/db"
	"github.com/strongo/db/gaedb"
	"golang.org/x/net/context"
	"testing"
)

func TestUsersLinker_LinkUsersWithinTransaction(t *testing.T) {
	c := context.Background()
	gaedb.SetupNdsMock()
	mockDB := SetupMocks(c)

	usersLinker := usersLinker{}

	var (
		err                            error
		entitiesToSave                 []db.EntityHolder
		inviterUser, invitedUser       models.AppUser
		inviterContact, invitedContact models.Contact
	)

	if inviterUser, err = dal.User.GetUserByID(c, 1); err != nil {
		t.Error("Failed to get inviter user", err)
		return
	}

	if invitedUser, err = dal.User.GetUserByID(c, 3); err != nil {
		t.Error("Failed to get invited user", err)
		return
	}

	if inviterContact, err = dal.Contact.GetContactByID(c, 6); err != nil {
		t.Error("Failed to get inviter user", err)
		return
	}

	if inviterContact.CounterpartyUserID != 0 {
		t.Error("inviterContact.CounterpartyUserID != 0")
	}

	if inviterContact.CounterpartyCounterpartyID != 0 {
		t.Error("inviterContact.CounterpartyCounterpartyID != 0")
	}

	err = mockDB.RunInTransaction(c, func(tc context.Context) (err error) {
		usersLinker = newUsersLinker(new(usersLinkingDbChanges))
		if err = usersLinker.linkUsersWithinTransaction(tc); err != nil {
			return err
		}
		return nil
	}, dal.CrossGroupTransaction)

	if err != nil {
		t.Error("Unexpected error:", err)
		return
	}

	if len(entitiesToSave) == 0 {
		t.Error("len(entitiesToSave) == 0")
		return
	}

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
