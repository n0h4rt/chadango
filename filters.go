package chadango

import (
	"regexp"
	"strings"

	"github.com/n0h4rt/chadango/utils"
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
//
// It provides methods for checking if an event passes the filter conditions,
// combining filters using logical operators (AND, OR, XOR), and negating filters using logical NOT.
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

// CombineFilter is a struct that represents the logical combination of two filters.
//
// It supports AND, OR, and XOR operations.
type CombineFilter struct {
	Left  Filter // Left represents the first filter to be combined.
	Right Filter // Right represents the second filter to be combined.
	Mode  int    // Mode specifies the combination mode: 0 for AND, 1 for OR, and 2 for XOR.
}

// Check returns true if the combination of the left and right filters satisfies the specified mode.
//
// Args:
//   - event: The event to check against the filter conditions.
//
// Returns:
//   - bool: True if the event passes the filter conditions, false otherwise.
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

// And returns a new [CombineFilter] that combines the current filter with the provided filter using logical AND.
//
// Args:
//   - filter: The filter to combine with the current filter.
//
// Returns:
//   - Filter: A new [CombineFilter] representing the logical AND of the current filter and the provided filter.
func (f *CombineFilter) And(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterAnd}
}

// Or returns a new [CombineFilter] that combines the current filter with the provided filter using logical OR.
//
// Args:
//   - filter: The filter to combine with the current filter.
//
// Returns:
//   - Filter: A new [CombineFilter] representing the logical OR of the current filter and the provided filter.
func (f *CombineFilter) Or(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterOr}
}

// Xor returns a new [CombineFilter] that combines the current filter with the provided filter using logical XOR.
//
// Args:
//   - filter: The filter to combine with the current filter.
//
// Returns:
//   - Filter: A new [CombineFilter] representing the logical XOR of the current filter and the provided filter.
func (f *CombineFilter) Xor(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterXor}
}

// Not returns a new [NotFilter] negating the current filter.
//
// Returns:
//   - Filter: A new [NotFilter] representing the logical NOT of the current filter.
func (f *CombineFilter) Not() Filter {
	return &NotFilter{f}
}

// NotFilter is a struct that represents the logical NOT of a filter.
//
// It negates the result of the base filter.
type NotFilter struct {
	Base Filter // Base represents the filter to be negated.
}

// Check returns the logical negation of the base filter's result.
//
// Args:
//   - event: The event to check against the filter conditions.
//
// Returns:
//   - bool: True if the event does not pass the base filter conditions, false otherwise.
func (f *NotFilter) Check(event *Event) bool {
	return !f.Base.Check(event)
}

// And returns a new [CombineFilter] that combines the current filter with the provided filter using logical AND.
//
// Args:
//   - filter: The filter to combine with the current filter.
//
// Returns:
//   - Filter: A new [CombineFilter] representing the logical AND of the current filter and the provided filter.
func (f *NotFilter) And(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterAnd}
}

// Or returns a new [CombineFilter] that combines the current filter with the provided filter using logical OR.
//
// Args:
//   - filter: The filter to combine with the current filter.
//
// Returns:
//   - Filter: A new [CombineFilter] representing the logical OR of the current filter and the provided filter.
func (f *NotFilter) Or(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterOr}
}

// Xor returns a new [CombineFilter] that combines the current filter with the provided filter using logical XOR.
//
// Args:
//   - filter: The filter to combine with the current filter.
//
// Returns:
//   - Filter: A new [CombineFilter] representing the logical XOR of the current filter and the provided filter.
func (f *NotFilter) Xor(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterXor}
}

// Not returns a new [NotFilter] negating the current filter.
//
// Returns:
//   - Filter: A new [NotFilter] representing the logical NOT of the current filter.
func (f *NotFilter) Not() Filter {
	return &NotFilter{f}
}

// UserFilter represents a filter for users.
//
// It filters events based on the username of the user associated with the event.
type UserFilter struct {
	Users []string // Users is a list of usernames to filter events based on user information.
}

// Check checks if the event's user is in the filter's list of users.
//
// Args:
//   - event: The event to check against the filter conditions.
//
// Returns:
//   - bool: True if the event's user is in the filter's list of users, false otherwise.
func (f *UserFilter) Check(event *Event) bool {
	if event.User == nil {
		return false
	}
	return utils.Contains(f.Users, strings.ToLower(event.User.Name))
}

// And returns a new [CombineFilter] that combines the current filter with the provided filter using logical AND.
//
// Args:
//   - filter: The filter to combine with the current filter.
//
// Returns:
//   - Filter: A new [CombineFilter] representing the logical AND of the current filter and the provided filter.
func (f *UserFilter) And(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterAnd}
}

// Or returns a new [CombineFilter] that combines the current filter with the provided filter using logical OR.
//
// Args:
//   - filter: The filter to combine with the current filter.
//
// Returns:
//   - Filter: A new [CombineFilter] representing the logical OR of the current filter and the provided filter.
func (f *UserFilter) Or(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterOr}
}

// Xor returns a new [CombineFilter] that combines the current filter with the provided filter using logical XOR.
//
// Args:
//   - filter: The filter to combine with the current filter.
//
// Returns:
//   - Filter: A new [CombineFilter] representing the logical XOR of the current filter and the provided filter.
func (f *UserFilter) Xor(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterXor}
}

// Not returns a new [NotFilter] negating the current filter.
//
// Returns:
//   - Filter: A new [NotFilter] representing the logical NOT of the current filter.
func (f *UserFilter) Not() Filter {
	return &NotFilter{f}
}

// Add adds a user to the filter's list of users.
//
// Args:
//   - userName: The username to add to the filter's list.
func (f *UserFilter) Add(userName string) {
	f.Users = append(f.Users, strings.ToLower(userName))
}

// Remove removes a user from the filter's list of users.
//
// Args:
//   - userName: The username to remove from the filter's list.
func (f *UserFilter) Remove(userName string) {
	f.Users = utils.Remove(f.Users, strings.ToLower(userName))
}

// NewUserFilter returns a new [UserFilter].
//
// Args:
//   - usernames: A list of usernames to filter events based on user information.
//
// Returns:
//   - Filter: A new [UserFilter] initialized with the provided usernames.
func NewUserFilter(usernames ...string) Filter {
	return &UserFilter{Users: usernames}
}

// ChatFilter represents a filter for chats.
//
// It filters events based on the name of the chat associated with the event.
type ChatFilter struct {
	Chats []string // Chats is a list of chat names to filter events based on chat information.
}

// Check checks if the event's group is in the filter's list of chats.
//
// Args:
//   - event: The event to check against the filter conditions.
//
// Returns:
//   - bool: True if the event's group is in the filter's list of chats, false otherwise.
func (f *ChatFilter) Check(event *Event) bool {
	if event.Group == nil {
		return false
	}
	return utils.Contains(f.Chats, event.Group.Name)
}

// And returns a new [CombineFilter] that combines the current filter with the provided filter using logical AND.
//
// Args:
//   - filter: The filter to combine with the current filter.
//
// Returns:
//   - Filter: A new [CombineFilter] representing the logical AND of the current filter and the provided filter.
func (f *ChatFilter) And(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterAnd}
}

// Or returns a new [CombineFilter] that combines the current filter with the provided filter using logical OR.
//
// Args:
//   - filter: The filter to combine with the current filter.
//
// Returns:
//   - Filter: A new [CombineFilter] representing the logical OR of the current filter and the provided filter.
func (f *ChatFilter) Or(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterOr}
}

// Xor returns a new [CombineFilter] that combines the current filter with the provided filter using logical XOR.
//
// Args:
//   - filter: The filter to combine with the current filter.
//
// Returns:
//   - Filter: A new [CombineFilter] representing the logical XOR of the current filter and the provided filter.
func (f *ChatFilter) Xor(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterXor}
}

// Not returns a new [NotFilter] negating the current filter.
//
// Returns:
//   - Filter: A new [NotFilter] representing the logical NOT of the current filter.
func (f *ChatFilter) Not() Filter {
	return &NotFilter{f}
}

// Add adds a chat to the filter's list of chats.
//
// Args:
//   - chatName: The chat name to add to the filter's list.
func (f *ChatFilter) Add(chatName string) {
	f.Chats = append(f.Chats, strings.ToLower(chatName))
}

// Remove removes a chat from the filter's list of chats.
//
// Args:
//   - chatName: The chat name to remove from the filter's list.
func (f *ChatFilter) Remove(chatName string) {
	f.Chats = utils.Remove(f.Chats, strings.ToLower(chatName))
}

// NewChatFilter returns a new [ChatFilter].
//
// Args:
//   - groupnames: A list of chat names to filter events based on chat information.
//
// Returns:
//   - Filter: A new [ChatFilter] initialized with the provided chat names.
func NewChatFilter(groupnames ...string) Filter {
	return &ChatFilter{Chats: groupnames}
}

// RegexFilter represents a [Message] filter based on a regular expression pattern.
//
// It filters messages based on whether their text matches the specified regular expression pattern.
type RegexFilter struct {
	Pattern *regexp.Regexp // Pattern represents the regular expression pattern used for filtering messages.
}

// Check checks if the event matches the regular expression pattern.
//
// Args:
//   - event: The event to check against the filter conditions.
//
// Returns:
//   - bool: True if the event's message text matches the regular expression pattern, false otherwise.
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

// And returns a new [CombineFilter] that combines the current filter with the provided filter using logical AND.
//
// Args:
//   - filter: The filter to combine with the current filter.
//
// Returns:
//   - Filter: A new [CombineFilter] representing the logical AND of the current filter and the provided filter.
func (f *RegexFilter) And(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterAnd}
}

// Or returns a new [CombineFilter] that combines the current filter with the provided filter using logical OR.
//
// Args:
//   - filter: The filter to combine with the current filter.
//
// Returns:
//   - Filter: A new [CombineFilter] representing the logical OR of the current filter and the provided filter.
func (f *RegexFilter) Or(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterOr}
}

// Xor returns a new [CombineFilter] that combines the current filter with the provided filter using logical XOR.
//
// Args:
//   - filter: The filter to combine with the current filter.
//
// Returns:
//   - Filter: A new [CombineFilter] representing the logical XOR of the current filter and the provided filter.
func (f *RegexFilter) Xor(filter Filter) Filter {
	return &CombineFilter{f, filter, CombineFilterXor}
}

// Not returns a new [NotFilter] negating the current filter.
//
// Returns:
//   - Filter: A new [NotFilter] representing the logical NOT of the current filter.
func (f *RegexFilter) Not() Filter {
	return &NotFilter{f}
}

// NewRegexFilter returns a new [RegexFilter].
//
// Args:
//   - pattern: The regular expression pattern to use for filtering messages.
//
// Returns:
//   - Filter: A new [RegexFilter] initialized with the provided regular expression pattern.
func NewRegexFilter(pattern string) Filter {
	return &RegexFilter{Pattern: regexp.MustCompile(pattern)}
}
