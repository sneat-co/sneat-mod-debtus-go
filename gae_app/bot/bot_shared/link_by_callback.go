package bot_shared

type callbackLink struct {
}

var CallbackLink = callbackLink{}

func (callbackLink) ToGroup(groupID string, isEdit bool) string {
	s := GROUP_COMMAND + "?id=" + groupID
	if isEdit {
		s += "&edit=1"
	}
	return s
}
