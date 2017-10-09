package models

//go:generate ffjson $GOFILE

import (
	"github.com/pquerna/ffjson/ffjson"
)

type GroupMemberJson struct {
	MemberJson
}

var _ SplitMember = (*GroupMemberJson)(nil)

func (m *GroupMemberJson) String() string {
	buffer, _ := ffjson.MarshalFast(m)
	return string(buffer)
}
