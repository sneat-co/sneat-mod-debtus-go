package api

import (
	"bitbucket.com/debtstracker/gae_app/debtstracker/auth"
	"golang.org/x/net/context"
	"net/http"
	"github.com/strongo/app/log"
	"strings"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
	"bytes"
	"io"
)

func panicUnknownStatus(status string) {
	panic("Unknown status: " + status)
}
func handleGetUserData(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	log.Debugf(c, "handleGetUserData(authInfo.UserID: %d)", authInfo.UserID)
	user, err := getApiUser(c, w, r, authInfo)
	if err != nil {
		return
	}
	markResponseAsJson(w.Header())

	rPath := r.URL.Path

	//getQueryValue := r.URL.Query().Get
	getQueryValue := func(name string) string {
		prefix := "/"+name+"-"
		start := strings.Index(rPath, prefix) + len(prefix)
		if start < 0 {
			return ""
		}
		end := strings.Index(rPath[start:], "/")
		if end < 0 {
			end = len(rPath)
		} else {
			end += start
		}
		return rPath[start:end]
	}

	status := getQueryValue("status")

	if status != "" && status != models.STATUS_ACTIVE && status != models.STATUS_ARCHIVED {
		BadRequestMessage(c, w, "Unknown status: " + status)
		return
	}

	dataCodes := strings.Split(getQueryValue("load"), ",")
	if len(dataCodes) == 0 {
		BadRequestMessage(c, w, "Missing `load` parameter value")
		return
	}

	//log.Debugf(c, "load: %v", dataCodes)

	dataResults := make([]*bytes.Buffer, len(dataCodes))

	hasContent := false
	for i, dataCode := range dataCodes {
		//log.Debugf(c, "i=%d, dataCode=%v", i, dataCode)
		dataResults[i] = &bytes.Buffer{}
		switch dataCode {
		case "Contacts":
			if status == models.STATUS_ACTIVE || status == models.STATUS_ARCHIVED {
				hasContent = writeUserContactsToJson(c, dataResults[i], status, user) || hasContent
			} else {
				panicUnknownStatus(status)
			}
		case "Groups":
			if status == models.STATUS_ACTIVE || status == models.STATUS_ARCHIVED {
				hasContent = writeUserGroupsToJson(c, dataResults[i], status, user) || hasContent
			} else {
				panicUnknownStatus(status)
			}
		case "Bills":
			switch status {
			case models.STATUS_ACTIVE:
				hasContent = writeUserActiveBillsToJson(c, dataResults[i], user) || hasContent
			default:
				panicUnknownStatus(status)
			}
		case "BillSchedules":
			switch status {
			case models.STATUS_ACTIVE:
				hasContent = writeUserActiveBillSchedulesToJson(c, dataResults[i], user) || hasContent
			default:
				panicUnknownStatus(status)
			}
		default:
			BadRequestMessage(c, w, "Unknown data code: "+dataCode)
			return
		}
	}

	if !hasContent {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Write(([]byte)("{"))
	needComma := false
	for _, dataResult := range dataResults {
		if dataResult.Len() > 0 {
			if needComma {
				w.Write([]byte(","))
			} else {
				needComma = true
			}
			w.Write([]byte("\n"))
			w.Write(dataResult.Bytes())
		}
	}
	w.Write(([]byte)("\n}"))
}

func writeUserGroupsToJson(_ context.Context, w io.Writer, status string, user models.AppUser) bool {
	//log.Debugf(c, "writeUserGroupsToJson(status=%v)", status)
	var jsonVal string
	switch status {
	case models.STATUS_ACTIVE:
		jsonVal = user.GroupsJsonActive
	case models.STATUS_ARCHIVED:
		jsonVal = user.GroupsJsonArchived
	default:
		panicUnknownStatus(status)
	}
	if jsonVal != "" {
		w.Write(([]byte)(`"Groups":`))
		w.Write([]byte(jsonVal))
		return true
	}
	return false
}

func writeUserContactsToJson(c context.Context, w io.Writer, status string, user models.AppUser) bool {
	//log.Debugf(c, "writeUserContactsToJson(status=%v)", status)
	var jsonVal string
	switch status {
	case models.STATUS_ACTIVE:
		jsonVal = user.ContactsJsonActive
	case models.STATUS_ARCHIVED:
		jsonVal = user.ContactsJsonArchived
	default:
		panicUnknownStatus(status)
	}

	if jsonVal != "" {
		w.Write(([]byte)(`"Contacts":`))
		w.Write([]byte(jsonVal))
		return true
	}
	return false
}

func writeUserActiveBillsToJson(c context.Context, w io.Writer, user models.AppUser) bool {
	if user.BillsJsonActive != "" {
		log.Debugf(c, "User has BillsJsonActive")
		if user.BillsCountActive == 0 {
			log.Warningf(c, "User(id=%d).BillsJsonActive is not empty && BillsCountActive == 0", user.ID)
		}
		w.Write(([]byte)(`"Bills":`))
		w.Write([]byte(user.BillsJsonActive))
		return true
	}
	return false
}

func writeUserActiveBillSchedulesToJson(c context.Context, w io.Writer, user models.AppUser) bool {
	if user.BillSchedulesJsonActive != "" {
		log.Debugf(c, "User has BillSchedulesJsonActive")
		if user.BillSchedulesCountActive == 0 {
			log.Warningf(c, "User(id=%d).BillSchedulesJsonActive is not empty && BillSchedulesCountActive == 0", user.ID)
		}
		w.Write(([]byte)(`"BillSchedules":`))
		w.Write([]byte(user.BillSchedulesJsonActive))
		return true
	}
	return false
}
