package chadango

import (
	"regexp"
	"strings"
)

// This approach aims to simplify the syntax of combining filters.
// For example:
//   filter := Filter.And(Filter.And(Filter)).Or(Filter.Not())
// Instead of:
//   filter := Or(And(Filter, And(Filter, Filter)), Not(Filter)) (excluding the package name)
//
// Since Go does not support type inheritance nor method declaration with multiple receivers,
// everything needs to be explicitly declared.

// Filter is an interface that defines the methods for filtering events.
type Filter interface {
	Check(*Event) bool // Check evaluates if the given event passes the filter conditions.
	And(Filter) Filter // And returns a new filter that combines the current filter with another using logical AND.
	Or(Filter) Filter  // Or returns a new filter that combines the current filter with another using logical OR.
	Xor(Filter) Filter // Xor returns a new filter that combines the current filter with another using logical XOR.
	Not() Filter       // Not returns a new filter that negates the current filter using logical NOT.
}

const (
	// CombineFilterAnd combines filter using logical AND.
	CombineFilterAnd int = iota
	// CombineFilterAnd combines filter using logical OR.
	CombineFilterOr
	// CombineFilterAnd combines filter using logical XOR.
	CombineFilterXor
)

// CombineFilter is a struct that represents the logical AND of two filters.
type CombineFilter struct {
	Left  Filter // Left represents the first filter to be combined using logical AND.
	Right Filter // Right represents the second filter to be combined using logical AND.
	Mode  int    // Mode specifies the combination mode: 0 for AND, 1 for OR, and 2 for XOR.
}

// Check returns true if both the left and right filters return true.
func (f *CombineFilter) Check(event *Event) bool {
	switch f.Mode {
	case CombineFilterAnd:
		return f.Left.Check(event) && f.Right.Check(event)
	case CombineFilterOr:
		return f.Left.Check(event) || f.Right.Check(event)
	case CombineFilterXor:
		return (f.Left.Check(event) || f.Right.Check(event)) && !(f.Left.Check(event) && f.Right.Check(event))
	default:
		return false
	}
}

// And returns a new CombineFilter that combines the current filter with the provided filter using logical AND.
func (f *CombineFilter) And(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterAnd}
}

// Or returns a new CombineFilter that combines the current filter with the provided filter using logical OR.
func (f *CombineFilter) Or(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterOr}
}

// Xor returns a new CombineFilter that combines the current filter with the provided filter using logical XOR.
func (f *CombineFilter) Xor(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterXor}
}

// Not returns a new NotFilter negating the current filter.
func (f *CombineFilter) Not() Filter {
	return &NotFilter{f}
}

// NotFilter is a struct that represents the logical NOT of a filter.
type NotFilter struct {
	Base Filter // Base represents the filter to be negated using logical NOT.
}

// Check returns the logical negation of the filter's result.
func (f *NotFilter) Check(event *Event) bool {
	return !f.Base.Check(event)
}

// And returns a new CombineFilter that combines the current filter with the provided filter using logical AND.
func (f *NotFilter) And(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterAnd}
}

// Or returns a new CombineFilter that combines the current filter with the provided filter using logical OR.
func (f *NotFilter) Or(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterOr}
}

// Xor returns a new CombineFilter that combines the current filter with the provided filter using logical XOR.
func (f *NotFilter) Xor(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterXor}
}

// Not returns a new NotFilter negating the current filter.
func (f *NotFilter) Not() Filter {
	return &NotFilter{f}
}

// UserFilter represents a filter for users.
type UserFilter struct {
	Users []string // Users is a list of usernames to filter events based on user information.
}

// Check checks if the event's user is in the filter's list of users.
func (f *UserFilter) Check(event *Event) bool {
	if event.User == nil {
		return false
	}
	return Contains(f.Users, strings.ToLower(event.User.Name))
}

// And returns a new CombineFilter that combines the current filter with the provided filter using logical AND.
func (f *UserFilter) And(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterAnd}
}

// Or returns a new CombineFilter that combines the current filter with the provided filter using logical OR.
func (f *UserFilter) Or(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterOr}
}

// Xor returns a new CombineFilter that combines the current filter with the provided filter using logical XOR.
func (f *UserFilter) Xor(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterXor}
}

// Not returns a new NotFilter negating the current filter.
func (f *UserFilter) Not() Filter {
	return &NotFilter{f}
}

// Add adds a user to the filter's list of users.
func (f *UserFilter) Add(userName string) {
	f.Users = append(f.Users, strings.ToLower(userName))
}

// Remove removes a user from the filter's list of users.
func (f *UserFilter) Remove(userName string) {
	f.Users = Remove(f.Users, strings.ToLower(userName))
}

// NewUserFilter returns a new `UserFilter`.
func NewUserFilter(usernames ...string) Filter {
	return &UserFilter{Users: usernames}
}

// ChatFilter represents a filter for chats.
type ChatFilter struct {
	Chats []string // Chats is a list of chat names to filter events based on chat information.
}

// Check checks if the event's group is in the filter's list of chats.
func (f *ChatFilter) Check(event *Event) bool {
	if event.Group == nil {
		return false
	}
	return Contains(f.Chats, event.Group.Name)
}

// And returns a new CombineFilter that combines the current filter with the provided filter using logical AND.
func (f *ChatFilter) And(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterAnd}
}

// Or returns a new CombineFilter that combines the current filter with the provided filter using logical OR.
func (f *ChatFilter) Or(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterOr}
}

// Xor returns a new CombineFilter that combines the current filter with the provided filter using logical XOR.
func (f *ChatFilter) Xor(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterXor}
}

// Not returns a new NotFilter negating the current filter.
func (f *ChatFilter) Not() Filter {
	return &NotFilter{f}
}

// Add adds a chat to the filter's list of chats.
func (f *ChatFilter) Add(chatName string) {
	f.Chats = append(f.Chats, strings.ToLower(chatName))
}

// Remove removes a chat from the filter's list of chats.
func (f *ChatFilter) Remove(chatName string) {
	f.Chats = Remove(f.Chats, strings.ToLower(chatName))
}

// NewChatFilter returns a new `ChatFilter`.
func NewChatFilter(groupnames ...string) Filter {
	return &ChatFilter{Chats: groupnames}
}

// RegexFilter represents a Message filter based on a regular expression pattern.
type RegexFilter struct {
	Pattern *regexp.Regexp // Pattern represents the regular expression pattern used for filtering messages.
}

// Check checks if the event matches the regular expression pattern.
func (f *RegexFilter) Check(event *Event) bool {
	if f.Pattern == nil {
		return false
	}
	switch event.Type {
	case OnMessage:
		return f.Pattern.MatchString(event.Message.Text)
	}
	return false
}

// And returns a new CombineFilter that combines the current filter with the provided filter using logical AND.
func (f *RegexFilter) And(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterAnd}
}

// Or returns a new CombineFilter that combines the current filter with the provided filter using logical OR.
func (f *RegexFilter) Or(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterOr}
}

// Xor returns a new CombineFilter that combines the current filter with the provided filter using logical XOR.
func (f *RegexFilter) Xor(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterXor}
}

// Not returns a new NotFilter negating the current filter.
func (f *RegexFilter) Not() Filter {
	return &NotFilter{f}
}

// NewRegexFilter returns a new `RegexFilter`.
func NewRegexFilter(pattern string) Filter {
	return &RegexFilter{Pattern: regexp.MustCompile(pattern)}
}
