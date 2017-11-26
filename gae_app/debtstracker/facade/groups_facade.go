package facade

import (
	"strconv"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/pkg/errors"
	"github.com/strongo/app/gae"
	"github.com/strongo/app/slices"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/db"
	"github.com/strongo/log"
	"golang.org/x/net/context"
	"google.golang.org/appengine/delay"
	"google.golang.org/appengine/taskqueue"
)

type groupFacade struct {
}

var Group = groupFacade{}

func (groupFacade groupFacade) CreateGroup(c context.Context,
	groupEntity *models.GroupEntity,
	tgBotCode string,
	beforeGroupInsert func(tc context.Context, groupEntity *models.GroupEntity) (group models.Group, err error),
	afterGroupInsert func(c context.Context, group models.Group, user models.AppUser) (err error),
) (group models.Group, groupMember models.GroupMember, err error) {
	if err = dal.DB.RunInTransaction(c, func(c context.Context) error {
		var intUserID int64
		if intUserID, err = strconv.ParseInt(groupEntity.CreatorUserID, 10, 64); err != nil {
			return err
		}
		user, err := dal.User.GetUserByID(c, intUserID)
		if err != nil {
			return err
		}
		existingGroups := user.ActiveGroups()

		if beforeGroupInsert != nil {
			if group, err = beforeGroupInsert(c, groupEntity); err != nil {
				return err
			}
		}

		var groupMembersChanged bool
		groupMembersChanged, _, memberIndex, member, members := groupEntity.AddOrGetMember(groupEntity.CreatorUserID, "", user.FullName())
		member.Shares = 1
		members[memberIndex] = member
		groupEntity.SetGroupMembers(members)

		if group.ID == "" {
			for _, existingGroup := range existingGroups {
				if existingGroup.Name == groupEntity.Name {
					return errors.New("Duplicate group name")
				}
			}
			if group, err = dal.Group.InsertGroup(c, groupEntity); err != nil {
				return err
			}
		} else if groupMembersChanged {
			if err = dal.Group.SaveGroup(c, group); err != nil {
				return err
			}
		}

		groupJson := models.UserGroupJson{
			ID:           group.ID,
			Name:         group.Name,
			Note:         group.Note,
			MembersCount: group.MembersCount,
		}

		if tgBotCode != "" {
			for _, tgGroupBot := range groupJson.TgBots {
				if tgGroupBot == tgBotCode {
					goto botFound
				}
			}
			groupJson.TgBots = append(groupJson.TgBots, tgBotCode)
		botFound:
		}

		user.SetActiveGroups(append(existingGroups, groupJson))

		if afterGroupInsert != nil {
			if err = afterGroupInsert(c, group, user); err != nil {
				return err
			}
		}

		if err = dal.User.SaveUser(c, user); err != nil {
			return err
		}
		if err = groupFacade.DelayUpdateGroupUsers(c, group.ID); err != nil {
			return err
		}
		return err
	}, dal.CrossGroupTransaction); err != nil {
		return
	}
	log.Infof(c, "Group created, ID=%v", group.ID)
	return
}

type NewUser struct {
	Name string
	bots.BotUser
	ChatMember bots.WebhookActor
}

func (groupFacade) AddUsersToTheGroupAndOutstandingBills(c context.Context, groupID string, newUsers []NewUser) (models.Group, []NewUser, error) {
	if groupID == "" {
		panic("groupID is empty string")
	}
	if len(newUsers) == 0 {
		panic("len(newUsers) == 0")
	}
	var group models.Group
	if err := dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		changed := false
		if group, err = dal.Group.GetGroupByID(c, groupID); err != nil {
			return
		}
		j := 0
		for _, newUser := range newUsers {
			_, isChanged, _, _, groupMembers := group.AddOrGetMember(strconv.FormatInt(newUser.GetAppUserIntID(), 10), "", newUser.Name)
			changed = changed || isChanged
			if isChanged {
				group.SetGroupMembers(groupMembers)
				newUsers[j] = newUser
				j += 1
			}
		}
		newUsers = newUsers[:j]
		if changed {
			if err = dal.Group.SaveGroup(c, group); err != nil {
				return
			}
			if err = Group.DelayUpdateGroupUsers(c, group.ID); err != nil {
				return err
			}
		}
		return
	}, db.SingleGroupTransaction); err != nil {
		return group, newUsers, err
	}
	return group, newUsers, nil
}

var delayUpdateGroupUsers = delay.Func("updateGroupUsers", updateGroupUsers)

func (groupFacade) DelayUpdateGroupUsers(c context.Context, groupID string) error { // TODO: Move to DAL?
	if groupID == "" {
		panic("groupID is empty string")
	}
	return gae.CallDelayFunc(c, common.QUEUE_USERS, "update-group-users", delayUpdateGroupUsers, groupID)
}

func updateGroupUsers(c context.Context, groupID string) error {
	if groupID == "" {
		log.Criticalf(c, "groupID is empty string")
		return nil
	}

	log.Debugf(c, "updateGroupUsers(groupID=%v)", groupID)
	group, err := dal.Group.GetGroupByID(c, groupID)
	if err != nil {
		return err
	}
	var tasks []*taskqueue.Task
	for _, member := range group.GetGroupMembers() {
		if member.UserID != "" {
			task, err := gae.CreateDelayTask(common.QUEUE_USERS, "update-user-with-groups", delayUpdateUserWithGroups, member.UserID, []string{groupID}, []string{})
			if err != nil {
				return err
			}
			tasks = append(tasks, task)
		}
	}
	_, err = taskqueue.AddMulti(c, tasks, common.QUEUE_USERS)
	return err
}

var delayUpdateUserWithGroups = delay.Func("UpdateUserWithGroups", delayedUpdateUserWithGroups)

func delayedUpdateUserWithGroups(c context.Context, userID string, groupIDs2add, groupIDs2remove []string) (err error) {
	log.Debugf(c, "delayedUpdateUserWithGroups(userID=%d, groupIDs2add=%v, groupIDs2remove=%v)", userID, groupIDs2add, groupIDs2remove)
	groups2add := make([]models.Group, len(groupIDs2add))
	for i, groupID := range groupIDs2add {
		if groups2add[i], err = dal.Group.GetGroupByID(c, groupID); err != nil {
			return
		}
	}
	if err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		var user models.AppUser
		if user, err = dal.User.GetUserByStrID(c, userID); err != nil {
			return
		}
		return User.UpdateUserWithGroups(c, user, groups2add, groupIDs2remove)
	}, dal.SingleGroupTransaction); err != nil {
		return err
	}
	return err
}

func (userFacade) UpdateUserWithGroups(c context.Context, user models.AppUser, groups2add []models.Group, groups2remove []string) (err error) {
	log.Debugf(c, "updateUserWithGroup(user.ID=%d, len(groups2add)=%d, groups2remove=%v)", user.ID, len(groups2add), groups2remove)
	groups := user.ActiveGroups()
	updated := false
	if groups2add != nil {
		for _, group2add := range groups2add {
			updated = user.AddGroup(group2add, "") || updated
		}
	}
	if groups2remove != nil {
		for _, group2remove := range groups2remove {
			for i, group := range groups {
				if group.ID == group2remove {
					groups = append(groups[:i], groups[i+1:]...)
					updated = true
					continue
				}
			}
		}
	}
	if !updated {
		log.Debugf(c, "User is not update with groups")
		return
	}
	user.SetActiveGroups(groups)
	if err = dal.User.SaveUser(c, user); err != nil {
		return
	}
	return
}

var delayUpdateContactWithGroups = delay.Func("UpdateContactWithGroups", delayedUpdateContactWithGroup)

func (userFacade) DelayUpdateContactWithGroups(c context.Context, contactID int64, addGroupIDs, removeGroupIDs []string) error {
	return gae.CallDelayFunc(c, common.QUEUE_USERS, "update-contact-groups", delayUpdateContactWithGroups, contactID, addGroupIDs, removeGroupIDs)
}

func delayedUpdateContactWithGroup(c context.Context, contactID int64, addGroupIDs, removeGroupIDs []string) (err error) {
	log.Debugf(c, "delayedUpdateContactWithGroup(contactID=%d, addGroupIDs=%v, removeGroupIDs=%v)", contactID, addGroupIDs, removeGroupIDs)
	if _, err = dal.Contact.GetContactByID(c, contactID); err != nil {
		return
	}
	if err = dal.DB.RunInTransaction(c, func(c context.Context) error {
		return User.UpdateContactWithGroups(c, contactID, addGroupIDs, removeGroupIDs)
	}, dal.SingleGroupTransaction); err != nil {
		return
	}
	return
}

func (userFacade) UpdateContactWithGroups(c context.Context, contactID int64, addGroupIDs, removeGroupIDs []string) error {
	log.Debugf(c, "UpdateContactWithGroups(contactID=%d, addGroupIDs=%v, removeGroupIDs=%v)", contactID, addGroupIDs, removeGroupIDs)
	if contact, err := dal.Contact.GetContactByID(c, contactID); err != nil {
		return err
	} else {
		var isAdded, isRemoved bool
		contact.GroupIDs, isAdded = slices.MergeStrings(contact.GroupIDs, addGroupIDs)
		contact.GroupIDs, isRemoved = slices.FilterStrings(contact.GroupIDs, removeGroupIDs)
		if isAdded || isRemoved {
			return dal.Contact.SaveContact(c, contact)
		}
		return nil
	}
}

func (groupFacade) LeaveGroup(c context.Context, groupID string, userID string) (group models.Group, user models.AppUser, err error) {
	if err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		if group, err = dal.Group.GetGroupByID(c, groupID); err != nil {
			return
		}
		members := group.GetGroupMembers()
		groupChanged := false
		for i, m := range members {
			if m.UserID == userID {
				members = append(members[:i], members[i+1:]...)
				groupChanged = true
				break
			}
		}
		if groupChanged {
			group.SetGroupMembers(members)
			if err = dal.Group.SaveGroup(c, group); err != nil {
				return
			}
		}
		if user, err = dal.User.GetUserByStrID(c, userID); err != nil {
			return
		}
		groups := user.ActiveGroups()
		userChanged := false
		for i, g := range groups {
			if g.ID == groupID {
				groups = append(groups[:i], groups[i+1:]...)
				userChanged = true
				break
			}
		}
		if userChanged {
			user.SetActiveGroups(groups)
			if err = dal.User.SaveUser(c, user); err != nil {
				return
			}
		}
		if err = Group.DelayUpdateGroupUsers(c, groupID); err != nil {
			return
		}
		return
	}, db.CrossGroupTransaction); err != nil {
		return
	}
	return
}
