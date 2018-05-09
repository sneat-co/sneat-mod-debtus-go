package api

import (
	"fmt"
	"net/http"
	"strconv"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/api/dto"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/pkg/errors"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/strongo/decimal"
)

func handleGetBill(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	billID := r.URL.Query().Get("id")
	if billID == "" {
		BadRequestError(c, w, errors.New("Missing id parameter"))
		return
	}
	bill, err := facade.GetBillByID(c, billID)
	if err != nil {
		InternalError(c, w, err)
		return
	}
	billToResponse(c, w, authInfo.UserID, bill)
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
	var members []dto.BillMemberDto
	{
		membersJSON := r.PostFormValue("members")
		if err = ffjson.Unmarshal([]byte(membersJSON), &members); err != nil {
			BadRequestError(c, w, err)
			return
		}

	}
	if len(members) == 0 {
		BadRequestMessage(c, w, "No members has been provided")
		return
	}
	billEntity := models.NewBillEntity(models.BillCommon{
		Status:        models.BillStatusDraft,
		SplitMode:     splitMode,
		CreatorUserID: strconv.FormatInt(authInfo.UserID, 10),
		Name:          r.PostFormValue("name"),
		Currency:      models.Currency(r.PostFormValue("currency")),
		AmountTotal:   amount,
	})

	var (
		totalByMembers decimal.Decimal64p2
	)

	contactIDs := make([]int64, 0, len(members))
	memberUserIDs := make([]int64, 0, len(members))

	for i, member := range members {
		if member.ContactID == "" && member.UserID == "" {
			BadRequestMessage(c, w, fmt.Sprintf("members[%d]: ContactID == 0 && UserID == 0", i))
			return
		}
		if member.ContactID != "" {
			contactID, err := strconv.ParseInt(member.ContactID, 10, 64)
			if err != nil {
				BadRequestError(c, w, errors.WithMessage(err, "ContactID is not an integer"))
				return
			}
			contactIDs = append(contactIDs, contactID)
		}
		if member.UserID != "" {
			memberUserID, err := strconv.ParseInt(member.UserID, 10, 64)
			if err != nil {
				BadRequestError(c, w, errors.WithMessage(err, "memberUserID is not an integer"))
				return
			}
			memberUserIDs = append(memberUserIDs, memberUserID)
		}
	}

	var contacts []models.Contact
	if len(contactIDs) > 0 {
		if contacts, err = facade.GetContactsByIDs(c, contactIDs); err != nil {
			InternalError(c, w, err)
			return
		}
	}

	var memberUsers []*models.AppUser
	if len(memberUserIDs) > 0 {
		if memberUsers, err = facade.User.GetUsersByIDs(c, memberUserIDs); err != nil {
			InternalError(c, w, err)
			return
		}
	}

	billMembers := make([]models.BillMemberJson, len(members))
	for i, member := range members {
		if member.UserID != "" && member.ContactID != "" {
			BadRequestMessage(c, w, fmt.Sprintf("Member has both UserID and ContactID: %v, %v", member.UserID, member.ContactID))
			return
		}
		totalByMembers += member.Amount
		billMembers[i] = models.BillMemberJson{
			MemberJson: models.MemberJson{
				UserID: member.UserID,
				Shares: member.Share,
			},
			Percent: member.Percent,
			Owes:       member.Amount,
			Adjustment: member.Adjustment,
		}
		if member.ContactID != "" {
			for _, contact := range contacts {
				if strconv.FormatInt(contact.ID, 10) == member.ContactID {
					contactName := contact.FullName()
					billMembers[i].ContactByUser = models.MemberContactsJsonByUser{
						strconv.FormatInt(contact.UserID, 10): {
							ContactID:   member.ContactID,
							ContactName: contactName,
						},
					}
					if billMembers[i].Name == "" {
						billMembers[i].Name = contactName
					}
					goto contactFound
				}
			}
			BadRequestError(c, w, fmt.Errorf("contact not found by ID=%v", member.ContactID))
			return
		contactFound:
		}
		if member.UserID != "" {
			for _, u := range memberUsers {
				if strconv.FormatInt(u.ID, 10) == member.UserID {
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

	billEntity.SplitMode = models.SplitModePercentage

	if err = billEntity.SetBillMembers(billMembers); err != nil {
		InternalError(c, w, err)
		return
	}

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
	if bill.ID == "" {
		InternalError(c, w, errors.New("Required parameter bill.ID is empty string."))
		return
	}
	if bill.BillEntity == nil {
		InternalError(c, w, errors.New("Required parameter bill.BillEntity is nil."))
		return
	}
	billDto := dto.BillDto{
		ID:   bill.ID,
		Name: bill.Name,
		Amount: models.Amount{
			Currency: models.Currency(bill.Currency),
			Value:    decimal.Decimal64p2(bill.AmountTotal),
		},
	}
	billMembers := bill.GetBillMembers()
	members := make([]dto.BillMemberDto, len(billMembers))
	sUserID := strconv.FormatInt(userID, 10)
	for i, billMember := range billMembers {
		members[i] = dto.BillMemberDto{
			UserID:     billMember.UserID,
			ContactID:  billMember.ContactByUser[sUserID].ContactID,
			Amount:     billMember.Owes,
			Adjustment: billMember.Adjustment,
			Share:      billMember.Shares,
		}
	}
	billDto.Members = members
	jsonToResponse(c, w, map[string]dto.BillDto{"Bill": billDto}) // TODO: Define DTO as need to clean BillMember.ContactByUser
}
