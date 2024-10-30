package models

import "net/url"

// GroupInfo represents a group info.
type GroupInfo struct {
	OwnerMessage string `json:"ownr_msg"`
	Title        string `json:"title"`
}

// GetDescription returns the description of the group.
func (gi *GroupInfo) GetMessage() string {
	unquoted, _ := url.QueryUnescape(gi.OwnerMessage)
	return unquoted
}

// GetTitle returns the title of the group.
func (gi *GroupInfo) GetTitle() string {
	unquoted, _ := url.QueryUnescape(gi.Title)
	return unquoted
}
