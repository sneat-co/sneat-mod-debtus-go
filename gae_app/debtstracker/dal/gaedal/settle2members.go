package gaedal

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/decimal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"github.com/strongo/app/db"
	"github.com/strongo/app/log"
	"github.com/pkg/errors"
	"fmt"
)

func Settle2members(c context.Context, groupID, debtorID, sponsorID string, currency models.Currency, amount decimal.Decimal64p2) (err error) {
	log.Debugf(c, "Settle2members(groupID=%v, debtorID=%v, sponsorID=%v, currency=%v, amount=%v)", groupID, debtorID, sponsorID, currency, amount)
	query := datastore.NewQuery(models.BillKind)
	query = query.KeysOnly()
	query = query.Filter("UserGroupID=", groupID)
	query = query.Filter("Currency=", string(currency))
	query = query.Filter("DebtorIDs=", debtorID)
	query = query.Filter("SponsorIDs=", sponsorID)
	query = query.Order("DtCreated")
	query = query.Limit(20)

	keys, err := query.GetAll(c, nil)
	if len(keys) == 0 {
		log.Errorf(c, "No bills found to settle")
		return
	} else {
		log.Debugf(c, "keys: %v", keys)
	}

	err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		var (
			group models.Group
			groupDebtor, groupSponsor models.GroupMemberJson
		)
		if group, err = dal.Group.GetGroupByID(c, groupID); err != nil {
			return
		}
		if groupDebtor, err = group.GetGroupMemberByID(debtorID); err != nil {
			return errors.WithMessage(err, "unknown debtor ID="+debtorID)
		}
		if groupSponsor, err = group.GetGroupMemberByID(sponsorID); err != nil {
			return errors.WithMessage(err, "Unknown sponsor ID="+sponsorID)
		}

		if v, ok := groupDebtor.Balance[currency]; !ok {
			return fmt.Errorf("group debtor has no balance in currency=%v", currency)
		} else if -v < amount {
			log.Warningf(c, "Debtor balance is less then settling amount")
			amount = -v
		}
		if v, ok := groupSponsor.Balance[currency]; !ok {
			return fmt.Errorf("group sponsor has no balance in currency=%v", currency)
		} else if v < amount {
			log.Warningf(c, "sponsor balance is less then settling amount")
			amount = v
		}

		toSave := make([]db.EntityHolder, 1, len(keys) + 1)
		toSave[0] = &group

		for _, k := range keys {
			if amount == 0 {
				break
			} else if amount < 0 {
				panic(fmt.Sprintf("amount < 0: %v", amount))
			}
			var bill models.Bill
			if bill, err = dal.Bill.GetBillByID(c, k.StringID()); err != nil {
				return
			}
			billMembers := bill.GetBillMembers()
			var debtor, sponsor *models.BillMemberJson
			var debtorInvertedBalance, diff decimal.Decimal64p2
			for i := range billMembers {
				switch billMembers[i].ID {
				case debtorID:
					if debtor = &billMembers[i]; debtor.Balance() >= 0 {
						log.Warningf(c, "Got debtor %v with positive balance = %v", debtor.ID, debtor.Balance())
						goto nextBill
					}
					if sponsor != nil {
						break
					}
				case sponsorID:
					if sponsor = &billMembers[i]; sponsor.Balance() <= 0 {
						log.Warningf(c, "Got sponsor %v with negative balance = %v", sponsor.ID, sponsor.Balance())
						goto nextBill
					}
					if debtor != nil {
						break
					}
				}
			}
			if debtor == nil {
				log.Warningf(c, "Debtor not found by ID="+debtorID)
				goto nextBill
			}
			if sponsor == nil {
				log.Warningf(c, "Sponsor not found by ID="+sponsorID)
				goto nextBill
			}
			debtorInvertedBalance = -1 * debtor.Balance()
			if debtorInvertedBalance <= sponsor.Balance() {
				diff = debtorInvertedBalance
			} else {
				diff = sponsor.Balance()
			}

			if diff > amount {
				diff = amount
			}

			log.Debugf(c, "diff: %v", diff)
			amount -= diff

			debtor.Paid += diff
			sponsor.Paid -= diff
			groupDebtor.Balance[currency] += diff
			groupSponsor.Balance[currency] -= diff

			if err = bill.SetBillMembers(billMembers); err != nil {
				return
			}

			log.Debugf(c, "groupDebtor.Balance: %v", groupDebtor.Balance)
			log.Debugf(c, "groupSponsor.Balance: %v", groupSponsor.Balance)

			toSave = append(toSave, &*&bill)

			nextBill:
		}

		if len(toSave) > 1 {
			groupMembers := group.GetGroupMembers()
			for i, m := range groupMembers {
				switch m.ID {
				case debtorID:
					groupMembers[i] = groupDebtor
				case sponsorID:
					groupMembers[i] = groupSponsor
				}
			}
			if changed := group.SetGroupMembers(groupMembers); !changed {
				panic("Group members not changed - something wrong")
			}
			if err = dal.DB.UpdateMulti(c, toSave); err != nil {
				return
			}
		} else {
			log.Errorf(c, "No bills found to settle")
		}

		return
	}, db.CrossGroupTransaction)

	return
}
