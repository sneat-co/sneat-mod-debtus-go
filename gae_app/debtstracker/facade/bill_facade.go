package facade

import (
	"fmt"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/pkg/errors"
	//"github.com/strongo/app"
	"math"
	"strconv"
	"time"

	"context"
	"github.com/strongo/db"
	"github.com/strongo/decimal"
	"github.com/strongo/log"
)

type billFacade struct {
}

var Bill = billFacade{}

func (billFacade) AssignBillToGroup(c context.Context, inBill models.Bill, groupID, userID string) (bill models.Bill, group models.Group, err error) {
	bill = inBill
	if err = bill.AssignToGroup(groupID); err != nil {
		return
	}
	if bill.MembersCount == 0 {
		{ // Get group,
			var gc context.Context
			if bill.Currency == models.Currency("") {
				// we don't need to get it in transaction if no currency as balance will not be changed
				gc = dal.DB.NonTransactionalContext(c)
			} else {
				gc = c
			}
			if group, err = dal.Group.GetGroupByID(gc, groupID); err != nil {
				return
			}
		}
		if group.MembersCount > 0 {
			groupMembers := group.GetGroupMembers()

			billMembers := make([]models.BillMemberJson, len(groupMembers))
			paidIsSet := false
			for i, gm := range groupMembers {
				billMembers[i] = models.BillMemberJson{
					MemberJson: gm.MemberJson,
				}
				billMembers[i].AddedByUserID = userID
				if gm.UserID == bill.CreatorUserID {
					billMembers[i].Paid = bill.AmountTotal
					paidIsSet = true
				}
			}
			if !paidIsSet {
				for i, bm := range billMembers {
					if bm.UserID == userID {
						billMembers[i].Paid = bill.AmountTotal
						paidIsSet = true
						break
					}
				}
				if !paidIsSet { // current user is not members of the bill
					//group.AddOrGetMember(userID, "", )
					var user models.AppUser
					if user.ID, err = strconv.ParseInt(userID, 10, 64); err != nil {
						return
					}
					if user, err = User.GetUserByID(dal.DB.NonTransactionalContext(c), user.ID); err != nil {
						return
					}
					_, _, _, groupMember, _ := group.AddOrGetMember(userID, "", user.FullName())

					billMembers = append(billMembers, models.BillMemberJson{
						MemberJson: groupMember.MemberJson,
						Paid:       bill.AmountTotal,
					})
				}
			}
			if err = bill.SetBillMembers(billMembers); err != nil {
				return
			}
			if bill.Currency != models.Currency("") {
				if _, err = group.ApplyBillBalanceDifference(bill.Currency, bill.GetBalance().BillBalanceDifference(models.BillBalanceByMember{})); err != nil {
					return
				}
				if err = dal.Group.SaveGroup(c, group); err != nil {
					return
				}
			}
			log.Debugf(c, "bill.GetBillMembers(): %+v", bill.GetBillMembers())
		}
	}
	return
}

func (billFacade) CreateBill(c, tc context.Context, billEntity *models.BillEntity) (bill models.Bill, err error) {
	if c == nil {
		panic("Parameter c context.Context is required")
	}
	log.Debugf(c, "billFacade.CreateBill(%v)", billEntity)
	if tc == nil {
		panic("Parameter tc context.Context is required")
	}
	if billEntity == nil {
		panic("Parameter billEntity *models.BillEntity is required")
	}
	if !models.IsValidBillSplit(billEntity.SplitMode) {
		panic(fmt.Sprintf("billEntity.SplitMode has unknown value: %v", billEntity.SplitMode))
	}
	if billEntity.CreatorUserID == "" {
		err = errors.Wrap(ErrBadInput, "billEntity.CreatorUserID == 0")
		return
	}
	if billEntity.SplitMode == "" {
		err = errors.Wrap(ErrBadInput, "Missing required property SplitMode")
		return
	}
	if billEntity.AmountTotal == 0 {
		err = errors.Wrap(ErrBadInput, "billEntity.AmountTotal == 0")
		return
	}
	if billEntity.AmountTotal < 0 {
		err = errors.WithMessage(ErrBadInput, fmt.Sprintf("billEntity.AmountTotal < 0: %v", billEntity.AmountTotal))
		return
	}
	if billEntity.Status == "" {
		err = errors.Wrap(ErrBadInput, "billEntity.Status property is required")
		return
	}
	if !models.IsValidBillStatus(billEntity.Status) {
		err = errors.Wrapf(ErrBadInput, "Invalid status: %v, expected one of %v", billEntity.Status, models.BillStatuses)
		return
	}

	billEntity.DtCreated = time.Now()

	members := billEntity.GetBillMembers()
	//if len(members) == 0 {
	//	return bill, fmt.Errorf("len(members) == 0, MembersJson: %v", billEntity.MembersJson)
	//}

	if len(members) == 0 {
		billEntity.SplitMode = models.SplitModeEqually
	} else {
		contactIDs := make([]int64, 0, len(members)-1)

		var (
			totalPercentageByMembers decimal.Decimal64p2
			totalSharesPerMembers    int
			totalPaidByMembers       decimal.Decimal64p2
			totalOwedByMembers       decimal.Decimal64p2
			payersCount              int
			equalAmount              decimal.Decimal64p2
			shareAmount              decimal.Decimal64p2
		)

		switch billEntity.SplitMode {
		case models.SplitModeShare:
			shareAmount = decimal.NewDecimal64p2FromFloat64(
				math.Floor(billEntity.AmountTotal.AsFloat64()/float64(len(members))*100+0.5) / 100,
			)
		case models.SplitModeEqually:
			amountToSplitEqually := billEntity.AmountTotal
			var totalAdjustmentByMembers decimal.Decimal64p2
			for i, member := range members {
				if member.Adjustment > billEntity.AmountTotal {
					return bill, errors.WithMessage(ErrBadInput, fmt.Sprintf("members[%d].Adjustment > billEntity.AmountTotal", i))
				} else if member.Adjustment < 0 && member.Adjustment < -1*billEntity.AmountTotal {
					err = errors.WithMessage(ErrBadInput,
						fmt.Sprintf("members[%d].AdjustmentInCents < 0 && AdjustmentInCents < -1*billEntity.AmountTotal", i))
					return
				}
				totalAdjustmentByMembers += member.Adjustment
			}
			if totalAdjustmentByMembers > billEntity.AmountTotal {
				return bill, errors.Wrap(ErrBadInput, "totalAdjustmentByMembers > billEntity.AmountTotal")
			}
			amountToSplitEqually -= totalAdjustmentByMembers
			equalAmount = decimal.NewDecimal64p2FromFloat64(
				math.Floor(amountToSplitEqually.AsFloat64()/float64(len(members))*100+0.5) / 100,
			)
		}

		// We use it to check equal split
		amountsCountByValue := make(map[decimal.Decimal64p2]int)

		// Calculate totals & initial checks
		for i, member := range members {
			if member.Paid != 0 {
				payersCount += 1
				totalPaidByMembers += member.Paid
			}
			totalOwedByMembers += member.Owes
			totalSharesPerMembers += member.Shares

			// Individual member checks - we can't move this checks down as it should fail first before deviation checks
			{
				if member.Owes < 0 {
					err = errors.Wrapf(ErrBadInput, "members[%d].Owes is negative: %v", i, member.Owes)
					return
				}
				if member.UserID != billEntity.CreatorUserID {
					if len(member.ContactByUser) == 0 {
						err = fmt.Errorf("len(members[i].ContactByUser) == 0: i==%v", i)
						return
					}
					if member.UserID == "" {
						if len(member.ContactByUser) == 0 {
							err = errors.New("Bill member is missing ContactByUser ID.")
							return
						}

						for _, counterparty := range member.ContactByUser {
							if counterparty.ContactID == "" {
								panic("counterparty.ContactID == 0")
							}
							var counterpartyContactID int64
							counterpartyContactID, err = strconv.ParseInt(counterparty.ContactID, 10, 64)
							if err != nil {
								return
							}
							var duplicateContactID bool
							for _, cID := range contactIDs {
								if cID == counterpartyContactID {
									duplicateContactID = true
									break
								}
							}
							if !duplicateContactID {

								contactIDs = append(contactIDs, counterpartyContactID)
							}
						}
					}
				}
			}
		}

		adjustmentsCount := 0
		for i, member := range members {
			if member.Adjustment != 0 {
				adjustmentsCount++
			}
			ensureNoAdjustment := func() {
				if member.Adjustment != 0 {
					panic(fmt.Sprintf("Member #%d has Adjustment property not allowed with split mode %v", i, billEntity.SplitMode))
				}
			}
			ensureNoShare := func() {
				if member.Shares != 0 {
					panic(fmt.Sprintf("Member #%d has Shares property not allowed with split mode %v", i, billEntity.SplitMode))
				}
			}
			ensureEqualShare := func() {
				if member.Shares != members[0].Shares {
					panic(fmt.Sprintf("members[%d] has Shares not equal to members[0].Shares: %d != %d", i, member.Shares, members[i].Shares))
				}
			}

			ensureMemberAmountDeviateWithin1cent := func() error {
				//if totalOwedByMembers == 0 && totalOwedByMembers == 0 {
				//	return nil
				//}
				switch billEntity.SplitMode {
				case models.SplitModeShare:
					expectedAmount := int64(shareAmount) * int64(member.Shares)
					deviation := expectedAmount - int64(member.Owes)
					if deviation > 1 || deviation < -1 {
						return errors.Wrapf(ErrBadInput, "Member #%d has amount %v deviated too much (for %v) from expected %v.", i, member.Owes, decimal.Decimal64p2(deviation), decimal.Decimal64p2(expectedAmount))
					}
				default:
					deviation := int64(member.Owes - member.Adjustment - equalAmount)
					if deviation > 1 || deviation < -1 {
						return errors.Wrapf(ErrBadInput, "Member #%d has amount %v deviated too much (for %v) from equal %v.", i, member.Owes, decimal.Decimal64p2(deviation), equalAmount)
					}
				}
				return nil
			}
			switch billEntity.SplitMode {
			case models.SplitModeEqually:
				// ensureNoAdjustment()
				ensureEqualShare()
				if err = ensureMemberAmountDeviateWithin1cent(); err != nil {
					return
				}
				amountsCountByValue[member.Owes]++
			case models.SplitModeExactAmount:
				ensureNoAdjustment()
				ensureNoShare()
			case models.SplitModePercentage:
				totalPercentageByMembers += member.Percent
				// ensureNoAdjustment()
			case models.SplitModeShare:
				if member.Shares == 0 {
					err = errors.Wrapf(ErrBadInput, "Member %d is missing Shares value", i)
					return
				}
				// ensureNoAdjustment()
			}
		}

		if payersCount > 1 {
			return bill, ErrBillHasTooManyPayers
		}

		if !(billEntity.Status == models.STATUS_DRAFT && totalPaidByMembers == 0) && totalPaidByMembers != billEntity.AmountTotal {
			err = errors.WithMessage(ErrBadInput, fmt.Sprintf("Total paid for all members should be equal to billEntity amount (%v), got %v", billEntity.AmountTotal, totalPaidByMembers))
			return
		}
		switch billEntity.SplitMode {
		case models.SplitModeEqually:
			if len(amountsCountByValue) > 2+adjustmentsCount {
				return bill, errors.Wrapf(ErrBadInput, "len(amountsCountByValue):%v > 2 + adjustmentsCount:%v", amountsCountByValue, adjustmentsCount)
			}
		case models.SplitModePercentage:
			if totalPercentageByMembers != decimal.FromInt(100) {
				err = errors.WithMessage(ErrBadInput, fmt.Sprintf("Total percentage for all members should be 100%%, got %v%%", totalPercentageByMembers))
				return
			}
		case models.SplitModeShare:
			if billEntity.Shares == 0 {
				billEntity.Shares = totalSharesPerMembers
			} else if billEntity.Shares != totalSharesPerMembers {
				err = errors.WithMessage(ErrBadInput, fmt.Sprintf("billEntity.Shares != totalSharesPerMembers"))
				return
			}
		}

		if (totalOwedByMembers != 0 || totalPaidByMembers != 0) && totalOwedByMembers != billEntity.AmountTotal {
			err = fmt.Errorf("totalOwedByMembers != billEntity.AmountTotal: %v != %v", totalOwedByMembers, billEntity.AmountTotal)
			return
		}

		// Load counterparties so we can get respective userIDs
		var counterparties []models.Contact
		// Use non transactional context
		counterparties, err = GetContactsByIDs(c, contactIDs)
		if err != nil {
			return bill, errors.Wrap(err, "Failed to get counterparties by ID.")
		}

		// Assign userIDs from counterparty to respective member
		for _, member := range members {
			for _, counterparty := range counterparties {
				// TODO: assign not just for creator?
				if member.UserID == "" && member.ContactByUser[billEntity.CreatorUserID].ContactID == strconv.FormatInt(counterparty.ID, 10) {
					member.UserID = strconv.FormatInt(counterparty.CounterpartyUserID, 10)
					break
				}
			}
		}

		billEntity.ContactIDs = make([]string, len(contactIDs))
		for i, contactID := range contactIDs {
			billEntity.ContactIDs[i] = strconv.FormatInt(contactID, 10)
		}

	}

	if bill, err = InsertBillEntity(tc, billEntity); err != nil {
		return
	}

	billHistoryRecord := models.NewBillHistoryBillCreated(bill, nil)
	if err = dal.InsertWithRandomStringID(c, &billHistoryRecord, models.BillsHistoryIdLen); err != nil {
		return
	}
	return
}

//func (billFacade) CreateBillTransfers(c context.Context, billID string) error {
//	bill, err := facade.GetBillByID(c, billID)
//	if err != nil {
//		return err
//	}
//
//	members := bill.GetBillMembers()
//
//	{ // Verify payers count
//		payersCount := 0
//		for _, member := range members {
//			if member.Paid != 0 {
//				payersCount += 1
//			}
//		}
//		if payersCount == 0 {
//			return ErrBillHasNoPayer
//		} else if payersCount > 1 {
//			return ErrBillHasTooManyPayers
//		}
//	}
//
//	for _, member := range members {
//		if member.Paid == 0 {
//			creatorContactID := member.ContactByUser[bill.CreatorUserID].ContactID
//			if err = Bill.createBillTransfer(c, billID, strconv.FormatInt(bill.CreatorUserID, 10)); err != nil {
//				return errors.Wrapf(err, "Failed to create bill trasfer for %d", creatorContactID)
//			}
//		}
//	}
//	return nil
//}
//
//func (billFacade) createBillTransfer(c context.Context, billID string, creatorCounterpartyID int64) error {
//	err := dal.DB.RunInTransaction(c, func(c context.Context) error {
//		bill, err := facade.GetBillByID(c, billID)
//
//		if err != nil {
//			return err
//		}
//		members := bill.GetBillMembers()
//
//		var (
//			borrower *models.BillMemberJson
//			payer    *models.BillMemberJson
//		)
//		sCreatorUserID := strconv.FormatInt(bill.CreatorUserID, 10)
//		for _, member := range members {
//			if member.Paid > 0 {
//				if payer != nil {
//					return ErrBillHasTooManyPayers
//				}
//				payer = &member
//				if borrower != nil {
//					break
//				}
//			} else if member.ContactByUser[sCreatorUserID].ContactID == creatorCounterpartyID {
//				borrower = &member
//				if payer != nil {
//					break
//				}
//			}
//		}
//		if borrower == nil {
//			return errors.New("Bill member not found by creatorCounterpartyID")
//		}
//		if payer == nil {
//			return ErrBillHasNoPayer
//		}
//		//transferSource := dal.NewTransferSourceBot("api", "no-id", "0") // TODO: Needs refactoring! Move it out of DAL, do we really need an interface?
//
//		from := models.TransferCounterpartyInfo{
//			UserID:    payer.UserID,
//			ContactID: payer.ContactByUser[sCreatorUserID].ContactID,
//		}
//		to := models.TransferCounterpartyInfo{
//			UserID:    borrower.UserID,
//			ContactID: payer.ContactByUser[sCreatorUserID].ContactID,
//		}
//		log.Debugf(c, "from: %v", from)
//		log.Debugf(c, "to: %v", to)
//		//_, _, _, _, _, _, err = CreateTransfer(
//		//	c,
//		//	strongo.EnvUnknown,
//		//	transferSource,
//		//	bill.CreatorUserID,
//		//	billID,
//		//	false,
//		//	0,
//		//	from, to,
//		//	models.AmountTotal{Currency: models.Currency(bill.Currency), Value: bill.AmountTotal},
//		//	time.Time{},
//		//)
//		//if err != nil {
//		//	return err
//		//}
//		return nil
//	}, dal.CrossGroupTransaction)
//	return err
//}

type BillMemberUserInfo struct {
	ContactID string
	Name      string
}

func (billFacade) GetBillMembersUserInfo(c context.Context, bill models.Bill, forUserID int64) (billMembersUserInfo []BillMemberUserInfo, err error) {
	sUserID := strconv.FormatInt(forUserID, 10)

	for i, member := range bill.GetBillMembers() {
		var (
			billMemberContactJson models.MemberContactJson
			ok                    bool
		)
		if billMemberContactJson, ok = member.ContactByUser[sUserID]; !ok {
			err = fmt.Errorf("Member  #%d does not have information for %v", i, sUserID)
			return
		}
		billMembersUserInfo[i] = BillMemberUserInfo{
			ContactID: billMemberContactJson.ContactID,
			Name:      billMemberContactJson.ContactName,
		}
	}
	return
}

func (billFacade) AddBillMember(
	c context.Context, userID string, inBill models.Bill, memberID, memberUserID string, memberUserName string, paid decimal.Decimal64p2,
) (
	bill models.Bill, group models.Group, changed, isJoined bool, err error,
) {
	log.Debugf(c, "billFacade.AddBillMember(bill.ID=%v, memberID=%v, memberUserID=%v, memberUserName=%v, paid=%v)", bill.ID, memberID, memberUserID, memberUserName, paid)
	if paid < 0 {
		panic("paid < 0")
	}
	bill = inBill
	if bill.ID == "" {
		panic("bill.ID is empty string")
	}
	if !dal.DB.IsInTransaction(c) {
		panic("This method should be called within transaction")
	}

	// TODO: Verify bill was obtained within transaction

	previousBalance := bill.GetBalance()

	var (
		//isNew bool
		index                  int
		groupChanged           bool
		groupMember            models.GroupMemberJson
		billMember             models.BillMemberJson
		billMembers            []models.BillMemberJson
		groupMembers           []models.GroupMemberJson
		groupMembersJsonBefore string
	)

	totalAboutBefore := bill.AmountTotal

	if userGroupID := bill.UserGroupID(); userGroupID != "" {
		if group, err = dal.Group.GetGroupByID(c, userGroupID); err != nil {
			return
		}

		groupMembersJsonBefore = group.MembersJson

		if _, groupChanged, _, groupMember, groupMembers = group.AddOrGetMember(memberUserID, "", memberUserName); groupChanged {
			group.SetGroupMembers(groupMembers)
		} else {
			log.Debugf(c, "Group billMembers not changed, groupMember.ID: "+groupMember.ID)
		}
	}

	_, changed, index, billMember, billMembers = bill.AddOrGetMember(groupMember.ID, memberUserID, "", memberUserName)

	log.Debugf(c, "billMember.ID: "+billMember.ID)

	if paid > 0 {
		if billMember.Paid == paid {
			// Already set
		} else if paid == bill.AmountTotal {
			for i := range billMembers {
				billMembers[i].Paid = 0
			}
			billMember.Paid = paid
			changed = true
		} else {
			paidTotal := paid
			for _, bm := range billMembers {
				paidTotal += bm.Paid
			}
			if paidTotal <= bill.AmountTotal {
				billMember.Paid = paid
				changed = true
			} else {
				err = errors.New("Total paid by members exceeds bill amount")
				return
			}
		}
	}
	if !changed {
		return
	}

	billMembers[index] = billMember

	log.Debugf(c, "billMembers: %+v", billMembers)
	if err = bill.SetBillMembers(billMembers); err != nil {
		return
	}
	log.Debugf(c, "bill.GetBillMembers(): %+v", bill.GetBillMembers())

	if err = dal.Bill.SaveBill(c, bill); err != nil {
		return
	}

	log.Debugf(c, "bill.GetBillMembers() after save: %v", bill.GetBillMembers())

	currentBalance := bill.GetBalance()

	if balanceDifference := currentBalance.BillBalanceDifference(previousBalance); balanceDifference.IsNoDifference() {
		log.Debugf(c, "Bill balanceDifference: %v", balanceDifference)
		if groupChanged, err = group.ApplyBillBalanceDifference(bill.Currency, balanceDifference); err != nil {
			err = errors.WithMessage(err, "Failed to apply bill difference")
			return
		}
		if groupChanged {
			if err = dal.Group.SaveGroup(c, group); err != nil {
				return
			}
		}
	}

	log.Debugf(c, "group: %+v", group)
	var groupMembersJsonAfter string
	if group.GroupEntity != nil {
		groupMembersJsonAfter = group.MembersJson
	}
	billHistoryRecord := models.NewBillHistoryMemberAdded(userID, bill, totalAboutBefore, groupMembersJsonBefore, groupMembersJsonAfter)
	if err = dal.InsertWithRandomStringID(c, &billHistoryRecord, models.BillsHistoryIdLen); err != nil {
		return
	}

	isJoined = true
	return
}

var (
	ErrSettledBillsCanNotBeDeleted   = errors.New("settled bills can't be deleted")
	ErrOnlyDeletedBillsCanBeRestored = errors.New("only deleted bills can be restored")
)

func (billFacade) DeleteBill(c context.Context, billID string, userID int64) (bill models.Bill, err error) {
	if err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		if bill, err = GetBillByID(c, billID); err != nil {
			return
		}
		if bill.Status == models.BillStatusSettled {
			err = ErrSettledBillsCanNotBeDeleted
			return
		}
		if bill.Status == models.BillStatusDraft || bill.Status == models.BillStatusOutstanding {
			billHistoryRecord := models.NewBillHistoryBillDeleted(strconv.FormatInt(userID, 10), bill)
			if err = dal.InsertWithRandomStringID(c, &billHistoryRecord, models.BillsHistoryIdLen); err != nil {
				return
			}
			bill.Status = models.BillStatusDeleted
			if err = dal.Bill.SaveBill(c, bill); err != nil {
				return
			}
		}
		if groupID := bill.UserGroupID(); groupID != "" {
			var group models.Group
			if group, err = dal.Group.GetGroupByID(c, groupID); err != nil {
				return
			}
			outstandingBills := group.GetOutstandingBills()
			for i, billJson := range outstandingBills {
				if billJson.ID == billID {
					outstandingBills = append(outstandingBills[:i], outstandingBills[i+1:]...)
					group.SetOutstandingBills(outstandingBills)
					groupMembers := group.GetGroupMembers()
					billMembers := bill.GetBillMembers()
					for j, groupMember := range groupMembers {
						for _, billMember := range billMembers {
							if billMember.ID == groupMember.ID {
								groupMember.Balance[bill.Currency] -= billMember.Balance()
								groupMembers[j] = groupMember
								break
							}
						}
					}
					group.SetGroupMembers(groupMembers)
					if err = dal.Group.SaveGroup(c, group); err != nil {
						return
					}
					break
				}
			}
		}
		return
	}, db.CrossGroupTransaction); err != nil {
		return
	}
	return
}

func (billFacade) RestoreBill(c context.Context, billID string, userID int64) (bill models.Bill, err error) {
	if err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		if bill, err = GetBillByID(c, billID); err != nil {
			return
		}
		if bill.Status != models.BillStatusDeleted {
			err = ErrOnlyDeletedBillsCanBeRestored
			return
		}

		if bill.MembersCount > 1 {
			bill.Status = models.BillStatusOutstanding
		} else {
			bill.Status = models.BillStatusDraft
		}
		billHistoryRecord := models.NewBillHistoryBillRestored(strconv.FormatInt(userID, 10), bill)
		if err = dal.InsertWithRandomStringID(c, &billHistoryRecord, models.BillsHistoryIdLen); err != nil {
			return
		}
		if err = dal.Bill.SaveBill(c, bill); err != nil {
			return
		}
		if groupID := bill.UserGroupID(); groupID != "" {
			var group models.Group
			if group, err = dal.Group.GetGroupByID(c, groupID); err != nil {
				return
			}
			var groupChanged bool
			if groupChanged, err = group.AddBill(bill); err != nil {
				return
			} else if groupChanged {
				if err = dal.Group.SaveGroup(c, group); err != nil {
					return
				}
			}
		}
		return
	}, db.CrossGroupTransaction); err != nil {
		return
	}
	return
}

func GetBillByID(c context.Context, billID string) (bill models.Bill, err error) {
	bill.ID = billID
	err = dal.DB.Get(c, &bill)
	return
}

func InsertBillEntity(c context.Context, billEntity *models.BillEntity) (bill models.Bill, err error) {
	if billEntity == nil {
		panic("billEntity == nil")
	}
	if billEntity.CreatorUserID == "" {
		panic("CreatorUserID == 0")
	}
	if billEntity.AmountTotal == 0 {
		panic("AmountTotal == 0")
	}

	billEntity.DtCreated = time.Now()
	bill.BillEntity = billEntity

	err = dal.InsertWithRandomStringID(c, &bill, models.BillIdLen)
	return
}

//func (billFacade billFacade) createTransfers(c context.Context, splitID int64) error {
//	split, err := dal.Split.GetSplitByID(c, splitID)
//	if err != nil {
//		return err
//	}
//	bills, err := dal.Bill.GetBillsByIDs(c, split.BillIDs)
//
//	balances := billFacade.getBalances(splitID, bills)
//	balances = billFacade.cleanupBalances(balances)
//
//	for currency, totalsByMember := range balances {
//		for memberID, memberTotal := range totalsByMember {
//			if memberTotal.Balance() > 0 { // TODO: Create delay task
//				if err = billFacade.createTransfer(c, splitID, memberTotal.BillIDs, memberID, currency, memberTotal.Balance()); err != nil {
//					return err
//				}
//			}
//		}
//	}
//	return nil
//}
//
//func (billFacade) createTransfer(c context.Context, splitID int64, billIDs []int64, memberID, currency string, amount decimal.Decimal64p2) error {
//	return nil
//}
