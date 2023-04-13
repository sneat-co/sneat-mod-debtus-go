package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/api/dto"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/log"
)

func getUserID(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) (userID int64) {
	userID = authInfo.UserID

	if stringID := r.URL.Query().Get("user"); stringID != "" {
		var err error
		userID, err = strconv.ParseInt(stringID, 10, 64)
		if err != nil {
			BadRequestError(c, w, err)
			return
		}
		if !authInfo.IsAdmin && userID != authInfo.UserID {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}
	return
}

type UserCounterpartiesResponse struct {
	UserID         int64
	Counterparties []dto.ContactListDto
}

func handleCreateCounterparty(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	name := strings.TrimSpace(r.PostForm.Get("name"))
	email := strings.TrimSpace(r.PostForm.Get("email"))
	tel := strings.TrimSpace(r.PostForm.Get("tel"))

	contactDetails := models.ContactDetails{
		Username: name,
	}
	if len(email) > 0 {
		contactDetails.EmailAddressOriginal = email
	}
	if len(tel) > 0 {
		telNumber, err := strconv.ParseInt(tel, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		contactDetails.PhoneNumber = telNumber
	}
	counterparty, _, err := facade.CreateContact(c, authInfo.UserID, contactDetails)
	if err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}
	w.Write([]byte(strconv.FormatInt(counterparty.ID, 10)))
}

func getContactID(w http.ResponseWriter, query url.Values) (int64, error) {
	counterpartyID, err := strconv.ParseInt(query.Get("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}
	return counterpartyID, err
}

func handleGetContact(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	query := r.URL.Query()
	counterpartyID, err := getContactID(w, query)
	if err != nil {
		return
	}
	counterparty, err := facade.GetContactByID(c, tx, counterpartyID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	contactToResponse(c, w, authInfo, counterparty)
}

func contactToResponse(c context.Context, w http.ResponseWriter, authInfo auth.AuthInfo, contact models.Contact) {
	if !authInfo.IsAdmin && contact.UserID != authInfo.UserID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	transfers, hasMoreTransfers, err := dtdal.Transfer.LoadTransfersByContactID(c, contact.ID, 0, 100)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	counterpartyJson := dto.ContactDetailsDto{
		ContactListDto: dto.ContactListDto{
			Status: contact.Status,
			ContactDto: dto.ContactDto{
				ID:     strconv.FormatInt(contact.ID, 10),
				Name:   contact.FullName(),
				UserID: strconv.FormatInt(contact.UserID, 10),
			},
		},
		TransfersResultDto: dto.TransfersResultDto{
			HasMoreTransfers: hasMoreTransfers,
			Transfers:        dto.TransfersToDto(authInfo.UserID, transfers),
		},
	}
	if contact.BalanceJson != "" {
		balance := json.RawMessage(contact.BalanceJson)
		counterpartyJson.Balance = &balance
	}
	if contact.EmailAddressOriginal != "" {
		counterpartyJson.Email = &dto.EmailInfo{
			Address:     contact.EmailAddressOriginal,
			IsConfirmed: contact.EmailConfirmed,
		}
	}
	if contact.PhoneNumber != 0 {
		counterpartyJson.Phone = &dto.PhoneInfo{
			Number:      contact.PhoneNumber,
			IsConfirmed: contact.PhoneNumberConfirmed,
		}
	}
	if len(contact.GroupIDs) > 0 {
		for _, groupID := range contact.GroupIDs {
			var group models.Group
			if group, err = dtdal.Group.GetGroupByID(c, groupID); err != nil {
				ErrorAsJson(c, w, http.StatusInternalServerError, err)
				return
			}
			for _, member := range group.GetGroupMembers() {
				for _, memberContactID := range member.ContactIDs {
					if memberContactID == strconv.FormatInt(contact.ID, 10) {
						counterpartyJson.Groups = append(counterpartyJson.Groups, dto.ContactGroupDto{
							ID:           group.ID,
							Name:         group.Name,
							MemberID:     memberContactID,
							MembersCount: group.MembersCount,
						})
					}
				}
			}
		}
	}

	jsonToResponse(c, w, counterpartyJson)
}

//type CounterpartyTransfer struct {
//
//}

func handleDeleteContact(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	log.Debugf(c, "handleDeleteContact()")
	//err := r.ParseForm()
	//if err != nil {
	//	BadRequestError(c, hashedWriter, err)
	//	return
	//}
	contactID, err := getContactID(w, r.URL.Query())
	if err != nil {
		BadRequestError(c, w, err)
		return
	}
	log.Debugf(c, "contactID: %v", contactID)
	if _, err := facade.DeleteContact(c, contactID); err != nil {
		InternalError(c, w, err)
		return
	}
	log.Infof(c, "Contact deleted: %v", contactID)
}

func handleArchiveCounterparty(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	//err := r.ParseForm()
	//if err != nil {
	//	BadRequestError(c, hashedWriter, err)
	//	return
	//}
	contactID, err := getContactID(w, r.URL.Query())
	if err != nil {
		BadRequestError(c, w, err)
		return
	}
	if contact, err := facade.ChangeContactStatus(c, contactID, models.STATUS_ARCHIVED); err != nil {
		InternalError(c, w, err)
		return
	} else {
		contactToResponse(c, w, authInfo, contact)
	}
}

func handleActivateCounterparty(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	//err := r.ParseForm()
	//if err != nil {
	//	BadRequestError(c, hashedWriter, err)
	//	return
	//}

	contactID, err := getContactID(w, r.URL.Query())
	if err != nil {
		BadRequestError(c, w, err)
		return
	}
	if contact, err := facade.ChangeContactStatus(c, contactID, models.STATUS_ACTIVE); err != nil {
		InternalError(c, w, err)
		return
	} else {
		contactToResponse(c, w, authInfo, contact)
	}
}

func handleUpdateCounterparty(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	counterpartyID, err := getContactID(w, r.URL.Query())
	if err != nil {
		return
	}
	if err := r.ParseForm(); err != nil {
		BadRequestError(c, w, err)
		return
	}
	values := make(map[string]string, len(r.PostForm))
	for k, vals := range r.PostForm {
		switch len(vals) {
		case 1:
			values[k] = vals[0]
		case 0:
			values[k] = vals[0]
		default:
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Too many values for '%v'.", k)))
			return
		}
	}

	if counterpartyEntity, err := facade.UpdateContact(c, counterpartyID, values); err != nil {
		InternalError(c, w, err)
		return
	} else {
		contactToResponse(c, w, authInfo, models.NewContact(counterpartyID, counterpartyEntity))
	}
}
