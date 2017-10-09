package facade

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/pkg/errors"
	//"github.com/strongo/app"
	"github.com/strongo/app/log"
	"github.com/strongo/decimal"
	"golang.org/x/net/context"
	"math"
	"time"
	"strconv"
	"github.com/strongo/app/user"
	"github.com/strongo/app/db"
)

type billFacade struct {
}

var Bill = billFacade{}

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
	if billEntity.CreatorUserID == 0 {
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
		err = errors.WithMessage(ErrBadInput, fmt.Sprintf("billEntity.AmountTotal < 0", billEntity.AmountTotal))
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
	//	return bill, errors.New(fmt.Sprintf("len(members) == 0, MembersJson: %v", billEntity.MembersJson))
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

		if billEntity.SplitMode == models.SplitModeShare {
			shareAmount = decimal.NewDecimal64p2FromFloat64(
				math.Floor(billEntity.AmountTotal.AsFloat64()/float64(len(members))*100+0.5) / 100,
			)
		} else if billEntity.SplitMode == models.SplitModeEqually || billEntity.SplitMode == models.SplitModeAdjustment {
			amountToSplitEqually := billEntity.AmountTotal
			if billEntity.SplitMode == models.SplitModeAdjustment {
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
			}
			equalAmount = decimal.NewDecimal64p2FromFloat64(
				math.Floor(amountToSplitEqually.AsFloat64()/float64(len(members))*100+0.5) / 100,
			)
		}

		// Used to check equal split
		amountsCountByValue := make(map[decimal.Decimal64p2]int, 2)

		// Calculate totals & initial checks
		for i, member := range members {
			if member.Paid != 0 {
				payersCount += 1
				totalPaidByMembers += member.Paid
			}
			totalOwedByMembers += member.Owes
			totalSharesPerMembers += member.Shares

			if member.Adjustment != 0 && billEntity.SplitMode != models.SplitModeAdjustment {
				err = fmt.Errorf("members[%d].Adjustment != 0 && billEntity.SplitMode == %v", member.Adjustment, billEntity.SplitMode)
			}

			// Individual member checks - we can't move this checks down as it should fail first before deviation checks
			{
				if member.Owes < 0 {
					err = errors.Wrapf(ErrBadInput, "members[%d].Owes is negative: %v", i, member.Owes)
					return
				}
				if member.UserID != billEntity.CreatorUserID {
					if len(member.ContactByUser) == 0 {
						err = errors.New(fmt.Sprintf("len(members[i].ContactByUser) == 0: i==%v", i))
						return
					}
					if member.UserID == 0 {
						if len(member.ContactByUser) == 0 {
							err = errors.New("Bill member is missing ContactByUser ID.")
							return
						}

						for _, counterparty := range member.ContactByUser {
							if counterparty.ContactID == 0 {
								panic("counterparty.ContactID == 0")
							}
							var duplicateContactID bool
							for _, cID := range contactIDs {
								if cID == counterparty.ContactID {
									duplicateContactID = true
									break
								}
							}
							if !duplicateContactID {
								contactIDs = append(contactIDs, counterparty.ContactID)
							}
						}
					}
				}
			}
		}

		for i, member := range members {
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
				if totalOwedByMembers == 0 && totalOwedByMembers == 0 {
					return nil
				}
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
				ensureNoAdjustment()
				ensureEqualShare()
				if err = ensureMemberAmountDeviateWithin1cent(); err != nil {
					return
				}
				amountsCountByValue[member.Owes] += 1
			case models.SplitModeAdjustment: //TODO: Should we allow negative adjustments?
				ensureNoShare()
				if err = ensureMemberAmountDeviateWithin1cent(); err != nil {
					return
				}
			case models.SplitModeExactAmount:
				ensureNoAdjustment()
				ensureNoShare()
			case models.SplitModePercentage:
				ensureNoAdjustment()
			case models.SplitModeShare:
				if member.Shares == 0 {
					err = errors.Wrapf(ErrBadInput, "Member %d is missing Shares value", i)
					return
				}
				ensureNoAdjustment()
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
			if len(amountsCountByValue) > 2 {
				return bill, errors.Wrapf(ErrBadInput, "len(amountsCountByValue) > 2: %v", amountsCountByValue)

			}
		case models.SplitModePercentage:
			if int64(totalPercentageByMembers) != 100*100 {
				err = errors.WithMessage(ErrBadInput, fmt.Sprintf("Total percentage for all members should be 100%%, got %v", totalPercentageByMembers))
				return
			}
		case models.SplitModeShare:
			if billEntity.Shares == 0 {
				billEntity.Shares = totalSharesPerMembers
			} else if billEntity.Shares != totalSharesPerMembers {
				err = errors.WithMessage(ErrBadInput, fmt.Sprintf("billEntity.Shares != totalSharesPerMembers"))
				return
			}
		case models.SplitModeAdjustment:

		}

		if (totalOwedByMembers != 0 || totalPaidByMembers != 0) && totalOwedByMembers != billEntity.AmountTotal {
			err = fmt.Errorf("totalOwedByMembers != billEntity.AmountTotal: %v != %v", totalOwedByMembers, billEntity.AmountTotal)
			return
		}

		// Load counterparties so we can get respective userIDs
		var counterparties []models.Contact
		// Use non transactional context
		counterparties, err = dal.Contact.GetContactsByIDs(c, contactIDs)
		if err != nil {
			return bill, errors.Wrap(err, "Failed to get counterparties by ID.")
		}

		// Assign userIDs from counterparty to respective member
		sCreatorUserID := strconv.FormatInt(billEntity.CreatorUserID, 10)
		for _, member := range members {
			for _, counterparty := range counterparties {
				// TODO: assign not just for creator?
				if member.UserID == 0 && member.ContactByUser[sCreatorUserID].ContactID == counterparty.ID {
					member.UserID = counterparty.CounterpartyUserID
					break
				}
			}
		}
		billEntity.ContactIDs = contactIDs
	}

	if bill, err = dal.Bill.InsertBillEntity(tc, billEntity); err != nil {
		return
	}

	return
}

func (billFacade) CreateBillTransfers(c context.Context, billID string) error {
	bill, err := dal.Bill.GetBillByID(c, billID)
	if err != nil {
		return err
	}

	members := bill.GetBillMembers()

	{ // Verify payers count
		payersCount := 0
		for _, member := range members {
			if member.Paid != 0 {
				payersCount += 1
			}
		}
		if payersCount == 0 {
			return ErrBillHasNoPayer
		} else if payersCount > 1 {
			return ErrBillHasTooManyPayers
		}
	}

	sCreatorUserID := strconv.FormatInt(bill.CreatorUserID, 10)
	for _, member := range members {
		if member.Paid == 0 {
			creatorContactID := member.ContactByUser[sCreatorUserID].ContactID
			if err = Bill.createBillTransfer(c, billID, creatorContactID); err != nil {
				return errors.Wrapf(err, "Failed to create bill trasfer for %d", creatorContactID)
			}
		}
	}
	return nil
}

func (billFacade) createBillTransfer(c context.Context, billID string, creatorCounterpartyID int64) error {
	err := dal.DB.RunInTransaction(c, func(c context.Context) error {
		bill, err := dal.Bill.GetBillByID(c, billID)

		if err != nil {
			return err
		}
		members := bill.GetBillMembers()

		var (
			borrower *models.BillMemberJson
			payer    *models.BillMemberJson
		)
		sCreatorUserID := strconv.FormatInt(bill.CreatorUserID, 10)
		for _, member := range members {
			if member.Paid > 0 {
				if payer != nil {
					return ErrBillHasTooManyPayers
				}
				payer = &member
				if borrower != nil {
					break
				}
			} else if member.ContactByUser[sCreatorUserID].ContactID == creatorCounterpartyID {
				borrower = &member
				if payer != nil {
					break
				}
			}
		}
		if borrower == nil {
			return errors.New("Bill member not found by creatorCounterpartyID")
		}
		if payer == nil {
			return ErrBillHasNoPayer
		}
		//transferSource := dal.NewTransferSourceBot("api", "no-id", "0") // TODO: Needs refactoring! Move it out of DAL, do we really need an interface?

		from := models.TransferCounterpartyInfo{
			UserID:    payer.UserID,
			ContactID: payer.ContactByUser[sCreatorUserID].ContactID,
		}
		to := models.TransferCounterpartyInfo{
			UserID:    borrower.UserID,
			ContactID: payer.ContactByUser[sCreatorUserID].ContactID,
		}
		log.Debugf(c, "from: %v", from)
		log.Debugf(c, "to: %v", to)
		//_, _, _, _, _, _, err = CreateTransfer(
		//	c,
		//	strongo.EnvUnknown,
		//	transferSource,
		//	bill.CreatorUserID,
		//	billID,
		//	false,
		//	0,
		//	from, to,
		//	models.AmountTotal{Currency: models.Currency(bill.Currency), Value: bill.AmountTotal},
		//	time.Time{},
		//)
		//if err != nil {
		//	return err
		//}
		return nil
	}, dal.CrossGroupTransaction)
	return err
}

type BillMemberUserInfo struct {
	ContactID int64
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
			err = errors.New(fmt.Sprintf("Member  #%d does not have information for %v", i, sUserID))
			return
		}
		billMembersUserInfo[i] = BillMemberUserInfo{
			ContactID: billMemberContactJson.ContactID,
			Name:      billMemberContactJson.ContactName,
		}
	}
	return
}

func (billFacade billFacade) SplitBills(c context.Context, userID int64, groupID string, billIDs []string) (err error) {
	now := time.Now()

	var split models.Split

	err = dal.DB.RunInTransaction(c, func(tc context.Context) error {
		split, err = dal.Split.InsertSplit(tc, models.SplitEntity{
			OwnedByUser: user.OwnedByUser{
				AppUserIntID: userID,
				DtCreated:    now,
			},
			BillIDs: billIDs,
		})
		if err != nil {
			return err
		}

		bills, err := dal.Bill.GetBillsByIDs(tc, billIDs)
		if err != nil {
			return err
		}

		billEntityHolders := make([]db.EntityHolder, len(bills))
		for i, bill := range bills { // Assign splitID to bills and fails if a bill already assigned to another split
			if bill.SplitID == 0 {
				bill.SplitID = split.ID
			} else {
				return fmt.Errorf("bill %d already belongs to a split %d", bill.ID, bill.SplitID)
			}
			billEntityHolders[i] = &bill
		}

		if err = dal.DB.UpdateMulti(c, billEntityHolders); err != nil {
			return err
		}

		billFacade.createTransfers(c, split.ID)
		//var splitUser models.AppUser
		//
		//if groupID != 0 {
		//	splitUser, err = dal.User.GetUserByID(c, -1*groupID)
		//} else {
		//	splitUser, err = dal.User.GetUserByID(c, userID)
		//}

		return nil
	}, db.CrossGroupTransaction)

	return
}

func (billFacade billFacade) createTransfers(c context.Context, splitID int64) (error) {
	split, err := dal.Split.GetSplitByID(c, splitID)
	if err != nil {
		return err
	}
	bills, err := dal.Bill.GetBillsByIDs(c, split.BillIDs)

	balances := billFacade.getBalances(splitID, bills)
	balances = billFacade.cleanupBalances(balances)

	for currency, totalsByMember := range balances {
		for memberID, memberTotal := range totalsByMember {
			if memberTotal.Balance() > 0 { // TODO: Create delay task
				if err = billFacade.createTransfer(c, splitID, memberTotal.BillIDs, memberID, currency, memberTotal.Balance()); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (billFacade) createTransfer(c context.Context, splitID int64, billIDs []int64, memberID, currency string, amount decimal.Decimal64p2) error {
	return nil
}

type SplitMemberTotal struct {
	Paid    decimal.Decimal64p2
	Owes    decimal.Decimal64p2
	BillIDs []int64
}

func (t SplitMemberTotal) Balance() decimal.Decimal64p2 {
	return t.Paid - t.Owes
}

type SplitTotalsByMember map[string]SplitMemberTotal
type SplitTotalsByCurrency map[string]SplitTotalsByMember

func (billFacade) getBalances(splitID int64, bills []models.Bill) (balanceByCurrency SplitTotalsByCurrency) {
	balanceByCurrency = make(SplitTotalsByCurrency)
	for _, bill := range bills {
		balanceByMember := balanceByCurrency[bill.Currency]
		if balanceByMember == nil {
			balanceByMember = make(SplitTotalsByMember, bill.MembersCount)
			balanceByCurrency[bill.Currency] = balanceByMember
		}
		for _, member := range bill.GetBillMembers() {
			memberTotal := balanceByMember[member.ID]
			memberTotal.Paid += member.Paid
			memberTotal.Owes += member.Owes
			balanceByMember[member.ID] = memberTotal
		}
	}
	return
}

func (billFacade) cleanupBalances(balanceByCurrency SplitTotalsByCurrency) SplitTotalsByCurrency {
	return balanceByCurrency
}
