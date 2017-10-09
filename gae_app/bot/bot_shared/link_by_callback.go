package bot_shared

import "strconv"

type callbackLink struct {
}

var CallbackLink = callbackLink{}

func (callbackLink) ToGroup(groupID string, isEdit bool) string {
	s := GROUP_COMMAND + "?id=" + strconv.FormatInt(groupID, 10)
	if isEdit {
		s += "&edit=1"
	}
	return s
}
