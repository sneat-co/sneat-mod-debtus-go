package gaedal

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/pkg/errors"
	"github.com/strongo/app/gae"
	"github.com/strongo/db"
	"github.com/strongo/decimal"
	"github.com/strongo/log"
	"google.golang.org/appengine/delay"
	"sync"
)

const updateUsersWithBillKeyName = "update-users-with-bill"

func DelayUpdateUsersWithBill(c context.Context, billID string, userIDs []string) (err error) {
	return gae.CallDelayFunc(c, common.QUEUE_BILLS, updateUsersWithBillKeyName, delayUpdateUsersWithBill, billID, userIDs)
}

var delayUpdateUsersWithBill = delay.Func(updateUsersWithBillKeyName, updateUsersWithBill)

func updateUsersWithBill(c context.Context, billID string, userIDs []string) (err error) {
	wg := new(sync.WaitGroup)
	wg.Add(len(userIDs))
	for i := range userIDs {
		go func(i int) {
			defer wg.Done()
			if err2 := gae.CallDelayFunc(c, common.QUEUE_BILLS, updateUserWithBillKeyName, delayUpdateUserWithBill, billID, userIDs[i]); err != nil {
				err = err2
			}
		}(i)
	}
	wg.Wait()
	return
}

const updateUserWithBillKeyName = "update-user-with-bill"

var delayUpdateUserWithBill = delay.Func(updateUserWithBillKeyName, updateUserWithBill)

func updateUserWithBill(c context.Context, billID, userID string) (err error) {
	log.Debugf(c, "updateUserWithBill(billID=%v, userID=%v)", billID, userID)
	var (
		bill             models.Bill
		wg               sync.WaitGroup
		billErr          error
		userChanged      bool
		userIsBillMember bool
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if bill, billErr = facade.GetBillByID(c, billID); err != nil {
			return
		}
	}()
	if err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		var user models.AppUser
		if user, err = dal.User.GetUserByStrID(c, userID); err != nil {
			return
		}
		wg.Wait()
		if billErr != nil {
			return errors.WithMessage(billErr, "failed to get bill")
		} else if bill.BillEntity == nil {
			return errors.New("bill.BillEntity == nil")
		}
		var userBillBalance decimal.Decimal64p2
		if bill.Status != models.BillStatusDeleted {
			for _, billMember := range bill.GetBillMembers() {
				if billMember.UserID == userID {
					userBillBalance = billMember.Balance()
					userIsBillMember = true
					log.Debugf(c, "userBillBalance: %v; billMember.Owes: %v; billMember.Paid: %v",
						userBillBalance, billMember.Owes, billMember.Paid)
					break
				}
			}
		}

		log.Debugf(c, "userIsBillMember: %v", userIsBillMember)

		shouldBeInOutstanding := userIsBillMember && (bill.Status == models.BillStatusOutstanding || bill.Status == models.BillStatusDraft)
		userOutstandingBills := user.GetOutstandingBills()
		for i, userOutstandingBill := range userOutstandingBills {
			if userOutstandingBill.ID == billID {
				if !shouldBeInOutstanding {
					// Remove bill info from the user
					userOutstandingBills = append(userOutstandingBills[:i], userOutstandingBills[i+1:]...)
				} else {
					if billUserGroupID := bill.UserGroupID(); userOutstandingBill.GroupID != billUserGroupID {
						userOutstandingBill.GroupID = billUserGroupID
						userChanged = true
					}
					if userOutstandingBill.UserBalance != userBillBalance {
						userOutstandingBill.UserBalance = userBillBalance
						userChanged = true
					}
					if userOutstandingBill.Total != bill.AmountTotal {
						userOutstandingBill.Total = bill.AmountTotal
						userChanged = true
					}
					if userOutstandingBill.Currency != bill.Currency {
						userOutstandingBill.Currency = bill.Currency
						userChanged = true
					}
					if userOutstandingBill.Name != bill.Name {
						userOutstandingBill.Name = bill.Name
						userChanged = true
					}
					userOutstandingBills[i] = userOutstandingBill
				}
				goto doneWithChanges
			}
		}
		if shouldBeInOutstanding {
			userOutstandingBills = append(userOutstandingBills, models.BillJson{
				ID:           bill.ID,
				Name:         bill.Name,
				MembersCount: bill.MembersCount,
				Total:        bill.AmountTotal,
				Currency:     bill.Currency,
				UserBalance:  userBillBalance,
				GroupID:      bill.UserGroupID(),
			})
			userChanged = true
		}
	doneWithChanges:
		if userChanged {
			if _, err = user.SetOutstandingBills(userOutstandingBills); err != nil {
				return
			}
			if err = facade.User.SaveUser(c, user); err != nil {
				return
			}
		} else {
			log.Debugf(c, "User not changed, ID: %v", user.ID)
		}
		return
	}, db.SingleGroupTransaction); err != nil {
		if db.IsNotFound(err) {
			log.Errorf(c, err.Error())
			err = nil
		}
		return
	}
	if userChanged {
		log.Infof(c, "User %v updated with info for bill %v", userID, billID)
	}
	return
}
