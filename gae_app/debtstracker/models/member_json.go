package models

//go:generate ffjson $GOFILE

type MemberJson struct {
	ID            string
	Name          string                   `json:",omitempty"`
	UserID        int64                    `json:",omitempty"`
	TgUserID      int64                    `json:",omitempty"`
	ContactIDs    []int64                  `json:",omitempty"`
	ContactByUser MemberContactsJsonByUser `json:",omitempty"`
	Shares        int                      `json:",omitempty"`
}

var _ SplitMember = (*MemberJson)(nil)

func (m MemberJson) GetID() string {
	return m.ID
}

func (m MemberJson) GetName() string {
	return m.Name
}

func (m MemberJson) GetShares() int {
	return m.Shares
}
