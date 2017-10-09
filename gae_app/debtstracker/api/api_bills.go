package api

//go:generate ffjson $GOFILE

import (
	"bitbucket.com/debtstracker/gae_app/debtstracker/auth"
	"bitbucket.com/debtstracker/gae_app/debtstracker/dal"
	"bitbucket.com/debtstracker/gae_app/debtstracker/facade"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
	"fmt"
	"github.com/pkg/errors"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/strongo/decimal"
	"golang.org/x/net/context"
	"net/http"
	"strconv"
)

type BillDto struct {
	// TODO: Generate ffjson
	ID      int64
	Name    string
	Amount  models.Amount
	Members []BillMemberDto
}

func handleGetBill(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	billID, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		BadRequestError(c, w, err)
		return
	}
	bill, err := dal.Bill.GetBillByID(c, billID)
	if err != nil {
		InternalError(c, w, err)
		return
	}
	billToResponse(c, w, authInfo.UserID, bill)
}

type BillMemberDto struct {
	UserID     int64               `json:",omitempty"`
	ContactID  int64               `json:",omitempty"`
	Amount     decimal.Decimal64p2
	Paid       decimal.Decimal64p2 `json:",omitempty"`
	Share      int                 `json:",omitempty"`
	Adjustment decimal.Decimal64p2 `json:",omitempty"`
}

func handleCreateBill(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	splitMode := models.SplitMode(r.PostFormValue("split"))
	if !models.IsValidBillSplit(splitMode) {
		BadRequestMessage(c, w, fmt.Sprintf("Split parameter has unkown value: %v", splitMode))
		return
	}
	amountStr := r.PostFormValue("amount")
	if amountStr == "" {
		BadRequestMessage(c, w, fmt.Sprintf("Missing required parameter: amount. %v", r.PostForm))
		return
	}
	amount, err := decimal.ParseDecimal64p2(amountStr)
	if err != nil {
		BadRequestError(c, w, err)
		return
	}
	var members []BillMemberDto
	if err = ffjson.Unmarshal([]byte(r.PostFormValue("members")), &members); err != nil {
		BadRequestError(c, w, err)
		return
	}
	if len(members) == 0 {
		BadRequestMessage(c, w, "No members has been provided")
		return
	}
	billEntity := models.NewBillEntity(models.BillCommon{
		Status: models.STATUS_DRAFT,
		SplitMode:     splitMode,
		CreatorUserID: authInfo.UserID,
		Name:          r.PostFormValue("name"),
		Currency:      r.PostFormValue("currency"),
		AmountTotal:   amount,
	})

	var (
		totalByMembers decimal.Decimal64p2
	)
	billMembers := make([]models.BillMemberJson, len(members))

	contactIDs := make([]int64, 0, len(members))
	memberUserIDs := make([]int64, 0, len(members))

	for i, member := range members {
		if member.ContactID == 0 && member.UserID == 0 {
			BadRequestMessage(c, w, fmt.Sprintf("members[%d]: ContactID == 0 && UserID == 0", i))
			return
		}
		if member.ContactID != 0 {
			contactIDs = append(contactIDs, member.ContactID)
		}
		if member.UserID != 0 {
			memberUserIDs = append(memberUserIDs, member.UserID)
		}
	}

	var contacts []models.Contact
	if len(contactIDs) > 0 {
		if contacts, err = dal.Contact.GetContactsByIDs(c, contactIDs); err != nil {
			InternalError(c, w, err)
			return
		}
	}

	var memberUsers []models.AppUser
	if len(memberUserIDs) > 0 {
		if memberUsers, err = dal.User.GetUsersByIDs(c, memberUserIDs); err != nil {
			InternalError(c, w, err)
			return
		}
	}

	for i, member := range members {
		if member.UserID != 0 && member.ContactID != 0 {
			BadRequestMessage(c, w, fmt.Sprintf("Member has both UserID and ContactID: %v, %v", member.UserID, member.ContactID))
			return
		}
		totalByMembers += member.Amount
		billMembers[i] = models.BillMemberJson{
			MemberJson: models.MemberJson{
				UserID: member.UserID,
				Shares: member.Share,
			},
			Owes:       member.Amount,
			Adjustment: member.Adjustment,
		}
		if member.ContactID != 0 {
			for _, contact := range contacts {
				if contact.ID == member.ContactID {
					billMembers[i].ContactByUser = models.MemberContactsJsonByUser{
						strconv.FormatInt(contact.UserID, 10): {
							ContactID:   contact.ID,
							ContactName: contact.FullName(),
						},
					}
					goto contactFound
				}
			}
			panic(fmt.Sprintf("Contact not found by ID=%d", member.ContactID))
		contactFound:
		}
		if member.UserID != 0 {
			for _, u := range memberUsers {
				if u.ID == member.UserID {
					billMembers[i].Name = u.FullName()
					break
				}
			}
		}
	}
	if totalByMembers != amount {
		BadRequestMessage(c, w, fmt.Sprintf("Total amount is not equal to sum of member's amounts: %v != %v", amount, totalByMembers))
		return
	}
	billEntity.SetBillMembers(billMembers)
	billEntity.SplitMode = models.SplitModePercentage

	var bill models.Bill
	err = dal.DB.RunInTransaction(c, func(tc context.Context) (err error) {
		bill, err = facade.Bill.CreateBill(c, tc, billEntity)
		return
	}, dal.CrossGroupTransaction)

	if err != nil {
		InternalError(c, w, err)
		return
	}
	billToResponse(c, w, authInfo.UserID, bill)
}

func billToResponse(c context.Context, w http.ResponseWriter, userID int64, bill models.Bill) {
	if userID == 0 {
		InternalError(c, w, errors.New("Required parameter userID == 0."))
		return
	}
	if bill.ID == 0 {
		InternalError(c, w, errors.New("Required parameter bill.ID == 0."))
		return
	}
	if bill.BillEntity == nil {
		InternalError(c, w, errors.New("Required parameter bill.BillEntity is nil."))
		return
	}
	billDto := BillDto{
		ID:   bill.ID,
		Name: bill.Name,
		Amount: models.Amount{
			Currency: models.Currency(bill.Currency),
			Value:    decimal.Decimal64p2(bill.AmountTotal),
		},
	}
	billMembers := bill.GetBillMembers()
	members := make([]BillMemberDto, len(billMembers))
	sUserID := strconv.FormatInt(userID, 10)
	for i, billMember := range billMembers {
		members[i] = BillMemberDto{
			UserID:     billMember.UserID,
			ContactID:  billMember.ContactByUser[sUserID].ContactID,
			Amount:     billMember.Owes,
			Adjustment: billMember.Adjustment,
			Share:      billMember.Shares,
		}
	}
	billDto.Members = members
	jsonToResponse(c, w, map[string]BillDto{"Bill": billDto}) // TODO: Define DTO as need to clean BillMember.ContactByUser
}
