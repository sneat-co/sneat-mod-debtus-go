package api

//go:generate ffjson $GOFILE

import (
	"bitbucket.com/debtstracker/gae_app/debtstracker/auth"
	"bitbucket.com/debtstracker/gae_app/debtstracker/dal"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/strongo/app/log"
	"golang.org/x/net/context"
	"net/http"
	"strconv"
	"strings"
	"bitbucket.com/debtstracker/gae_app/debtstracker/facade"
	"fmt"
	"io/ioutil"
	"sync"
	"github.com/strongo/app/db"
)

type GroupDto struct {
	ID           int64
	Name         string
	Status       string
	Note         string           `json:",omitempty"`
	MembersCount int              `json:",omitempty"`
	Members      []GroupMemberDto `json:",omitempty"`
}


type GroupMemberDto struct {
	ID        string
	UserID    int64  `json:",omitempty"`
	ContactID int64  `json:",omitempty"`
	Name      string `json:",omitempty"`
}

func handlerCreateGroup(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo, user models.AppUser) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	name := strings.TrimSpace(r.PostForm.Get("name"))
	note := strings.TrimSpace(r.PostForm.Get("note"))

	groupEntity := models.GroupEntity{
		CreatorUserID: authInfo.UserID,
		Name:          name,
	}
	if len(note) > 0 {
		groupEntity.Note = note
	}

	group, _, err := facade.Group.CreateGroup(c, &groupEntity, "", nil, nil)
	if err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}
	log.Infof(c, "Group created, ID: %d", group.ID)
	groupToResponse(c, w, group, user)
}

func handlerGetGroup(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo, user models.AppUser) {
	groupID, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		BadRequestError(c, w, err)
		return
	}
	group, err := dal.Group.GetGroupByID(c, groupID)
	if err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}
	if err = groupToResponse(c, w, group, user); err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}
}

func groupToResponse(c context.Context, w http.ResponseWriter, group models.Group, user models.AppUser) (error) {
	if jsons, err := groupsToJson([]models.Group{group}, user); err != nil {
		return err
	} else {
		markResponseAsJson(w.Header())
		w.Write(jsons[0])
		return nil
	}
}

func groupsToJson(groups []models.Group, user models.AppUser) (result [][]byte, err error) {
	result = make([][]byte, len(groups))

	groupStatuses := make(map[int64]string, len(groups))

	for _, group := range user.ActiveGroups() {
		groupStatuses[group.ID] = models.STATUS_ACTIVE
	}

	for i, group := range groups {
		groupDto := GroupDto{
			ID:           group.ID,
			Name:         group.Name,
			Note:         group.Note,
			MembersCount: group.MembersCount,
		}
		if status, ok := groupStatuses[group.ID]; ok {
			groupDto.Status = status
		} else {
			groupDto.Status = models.STATUS_ARCHIVED
		}
		contactsByID := user.ContactsByID()
		if group.MembersJson != "" {
			for _, member := range group.GetGroupMembers() {
				memberDto := GroupMemberDto{
					ID:   member.ID,
					Name: member.Name,
				}
				if member.UserID == user.ID {
					memberDto.Name = ""
					memberDto.UserID = user.ID
				} else if member.Name == "" {
					err = fmt.Errorf("Group(%d) has member(id=%d) without UserID and without Name", group.ID, member.ID)
					return
				}
				for _, contactID := range member.ContactIDs {
					if _, ok := contactsByID[contactID]; ok {
						memberDto.ContactID = contactID
						break
					}
				}
				groupDto.Members = append(groupDto.Members, memberDto)
			}
		}
		if result[i], err = ffjson.MarshalFast(&groupDto); err != nil {
			return
		}
	}
	return
}

func handleJoinGroups(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	defer r.Body.Close()

	var groupIDs []int64
	if body, err := ioutil.ReadAll(r.Body); err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
	} else if groupIDs, err = StringToInt64s(string(body), ","); err != nil {
		BadRequestError(c, w, err)
		return
	}

	groups := make([]models.Group, len(groupIDs))
	var user models.AppUser

	err := dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		if user, err = dal.User.GetUserByID(c, authInfo.UserID); err != nil {
			return
		}
		var waitGroup sync.WaitGroup
		waitGroup.Add(len(groupIDs))

		errs := make([]error, len(groupIDs))
		for i, groupID := range groupIDs {
			go func(i int) {
				var group models.Group
				if group, errs[i] = dal.Group.GetGroupByID(c, groupID); errs[i] != nil {
					waitGroup.Done()
					return
				}
				groups[i] = group
				userName := user.FullName()
				if userName == models.NO_NAME {
					userName = ""
				}
				if _, changed, _, _, members := group.AddOrGetMember(authInfo.UserID, 0, userName); changed {
					group.SetGroupMembers(members)
					if errs[i] = dal.Group.SaveGroup(c, group); errs[i] != nil {
						waitGroup.Done()
						return
					}
				}
				if errs[i] = facade.Group.DelayUpdateGroupUsers(c, groupID); errs[i] != nil {
					waitGroup.Done()
					return
				}
				waitGroup.Done()
			}(i)
		}
		waitGroup.Wait()
		for _, err = range errs {
			if err != nil {
				return
			}
		}

		if err = facade.User.UpdateUserWithGroups(c, user, groups, []int64{}); err != nil {
			return
		}

		return
	}, dal.CrossGroupTransaction)

	if err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}

	jsons, err := groupsToJson(groups, user)
	if err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
	}
	w.Write(([]byte)("["))
	lastJsonIndex := len(jsons) - 1
	for i, json := range jsons {
		w.Write(json)
		if i < lastJsonIndex {
			w.Write([]byte(","))
		}
	}
	w.Write(([]byte)("]"))
}

func handlerDeleteGroup(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {

}

func handlerUpdateGroup(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	log.Debugf(c, "handlerUpdateGroup()")
	idParam := r.URL.Query().Get("id")

	var (
		user  models.AppUser
		group models.Group
		err   error
	)

	if group.ID, err = strconv.ParseInt(idParam, 10, 64); err != nil {
		BadRequestError(c, w, err)
		return
	}

	groupName := strings.TrimSpace(r.FormValue("name"))
	groupNote := strings.TrimSpace(r.FormValue("note"))

	err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		if group, err = dal.Group.GetGroupByID(c, group.ID); err != nil {
			return
		}

		if group.CreatorUserID != authInfo.UserID {
			err = fmt.Errorf("User is not authrized to edit this group")
			return
		}

		changed := false
		if groupName != "" && group.Name != groupName {
			group.Name = groupName
			changed = true
		}
		if group.Note != groupNote {
			group.Note = groupNote
			changed = true
		}
		if changed {
			if err = dal.Group.SaveGroup(c, group); err != nil {
				return
			}
		}
		if user, err = dal.User.GetUserByID(c, authInfo.UserID); err != nil {
			return
		}

		if err = facade.User.UpdateUserWithGroups(c, user, []models.Group{group}, nil); err != nil {
			return
		}

		if err = facade.Group.DelayUpdateGroupUsers(c, group.ID); err != nil {
			return
		}

		return
	}, dal.CrossGroupTransaction)

	if err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}

	if err = groupToResponse(c, w, group, user); err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}
}

func handlerSetContactsToGroup(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo, user models.AppUser) {
	log.Debugf(c, "handlerSetContactsToGroup()")
	idParam := r.URL.Query().Get("id")

	var (
		groupID string
		group   models.Group
		err     error
	)

	if groupID, err = strconv.ParseInt(idParam, 10, 64); err != nil {
		BadRequestError(c, w, err)
		return
	}

	var (
		addContactIDs   []int64
		removeMemberIDs []string
	)
	if addContactIDs, err = StringToInt64s(r.FormValue("addContactIDs"), ","); err != nil {
		BadRequestError(c, w, err)
		return
	}
	removeMemberIDs = strings.Split(r.FormValue("removeMemberIDs"), ",")

	var contacts2add []models.Contact
	if contacts2add, err = dal.Contact.GetContactsByIDs(c, addContactIDs); err != nil {
		BadRequestError(c, w, err)
		return
	}

	for _, contact := range contacts2add {
		if contact.UserID != authInfo.UserID {
			BadRequestError(c, w, fmt.Errorf("Contact %d does not belong to the user %d", contact.ID, authInfo.UserID))
			return
		}
	}

	if err = dal.DB.RunInTransaction(c, func(c context.Context) error {
		if group, err = dal.Group.GetGroupByID(c, groupID); err != nil {
			return err
		}
		members := group.GetGroupMembers()
		changed := false
		changedContactIDs := make([]int64, 0, len(addContactIDs)+len(removeMemberIDs))

		var groupUserIDs []int64

		addGroupUserID := func(member models.GroupMemberJson) {
			if member.UserID != 0 && member.UserID != user.ID {
				groupUserIDs = append(groupUserIDs, member.UserID)
			}
		}

		for _, contact2add := range contacts2add {
			var (
				isChanged bool
			)
			for _, member := range members {
				for _, mContactID := range member.ContactIDs {
					if mContactID == contact2add.ID {
						goto found
					}
				}
			}
			_, isChanged, _, _, members = group.AddOrGetMember(contact2add.CounterpartyUserID, contact2add.ID, contact2add.FullName())
			if isChanged {
				changed = true
				changedContactIDs = append(changedContactIDs, contact2add.ID)
			}
		found:
		}

		for _, memberID := range removeMemberIDs {
			for i, member := range members {
				if member.ID == memberID {
					members = append(members[:i], members[i+1:]...)
					changed = true
					addGroupUserID(member)
					for _, contactID := range member.ContactIDs {
						for _, changedContactID := range changedContactIDs {
							if changedContactID == contactID {
								goto alreadyChanged
							}
						}
						changedContactIDs = append(changedContactIDs, contactID)
					alreadyChanged:
					}
				}
			}
		}
		if changed || len(changedContactIDs) > 0 { // Check for len(changedContactIDs) is excessive but just in case.
			group.SetGroupMembers(members)
			if err = dal.Group.SaveGroup(c, group); err != nil {
				return err
			}
		}

		{ // Executing this block outside of IF just in case for self-healing.
			if user, err = dal.User.GetUserByID(c, user.ID); err != nil {
				return err
			}
			if err = facade.User.UpdateUserWithGroups(c, user, []models.Group{group}, []int64{}); err != nil {
				return err
			}

			for _, member := range members {
				addGroupUserID(member)
			}

			if len(groupUserIDs) > 0 {
				if err = facade.Group.DelayUpdateGroupUsers(c, groupID); err != nil {
					return err
				}
			}

			if len(changedContactIDs) == 1 {
				err = facade.User.UpdateContactWithGroups(c, changedContactIDs[0], []int64{groupID}, []int64{})
			} else {
				for _, contactID := range changedContactIDs {
					if err = facade.User.DelayUpdateContactWithGroups(c, contactID, []int64{groupID}, []int64{}); err != nil {
						return err
					}
				}
			}
		}
		return err
	}, dal.CrossGroupTransaction); err != nil {
		if db.IsNotFound(err) {
			BadRequestError(c, w, err)
			return
		}
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}
	if err = groupToResponse(c, w, group, user); err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}
}


func StringToInt64s(s, sep string) (result []int64, err error) {
	if s == "" {
		return
	}
	vals := strings.Split(s, sep)
	result = make([]int64, len(vals))
	for i, val := range vals {
		if result[i], err = strconv.ParseInt(val, 10, 64); err != nil {
			return
		}
	}
	return
}

