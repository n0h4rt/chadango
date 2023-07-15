package chadango

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserFilter_Check(t *testing.T) {
	filter := NewUserFilter("user1", "user2", "user3")
	event1 := &Event{User: &User{Name: "user1"}}
	event2 := &Event{User: &User{Name: "user4"}}

	// Check if the filter matches event1
	result1 := filter.Check(event1)
	assert.True(t, result1, "UserFilter should match event1")

	// Check if the filter matches event2
	result2 := filter.Check(event2)
	assert.False(t, result2, "UserFilter should not match event2")
}

func TestUserFilter_Add_Remove(t *testing.T) {
	filter := NewUserFilter()
	filter.(*UserFilter).Add("user1")
	filter.(*UserFilter).Add("user2")

	// Check if the filter matches "user1"
	event1 := &Event{User: &User{Name: "user1"}}
	result1 := filter.Check(event1)
	assert.True(t, result1, "UserFilter should match event1")

	// Check if the filter matches "user2"
	event2 := &Event{User: &User{Name: "user2"}}
	result2 := filter.Check(event2)
	assert.True(t, result2, "UserFilter should match event2")

	filter.(*UserFilter).Remove("user2")

	// Check if the filter still matches "user2" after removal
	result3 := filter.Check(event2)
	assert.False(t, result3, "UserFilter should not match event2 after removal")
}

func TestChatFilter_Check(t *testing.T) {
	filter := NewChatFilter("chat1", "chat2", "chat3")
	event1 := &Event{Group: &Group{Name: "chat1"}}
	event2 := &Event{Group: &Group{Name: "chat4"}}

	// Check if the filter matches event1
	result1 := filter.Check(event1)
	assert.True(t, result1, "ChatFilter should match event1")

	// Check if the filter matches event2
	result2 := filter.Check(event2)
	assert.False(t, result2, "ChatFilter should not match event2")
}

func TestChatFilter_Add_Remove(t *testing.T) {
	filter := NewChatFilter()
	filter.(*ChatFilter).Add("chat1")
	filter.(*ChatFilter).Add("chat2")

	// Check if the filter matches "chat1"
	event1 := &Event{Group: &Group{Name: "chat1"}}
	result1 := filter.Check(event1)
	assert.True(t, result1, "ChatFilter should match event1")

	// Check if the filter matches "chat2"
	event2 := &Event{Group: &Group{Name: "chat2"}}
	result2 := filter.Check(event2)
	assert.True(t, result2, "ChatFilter should match event2")

	filter.(*ChatFilter).Remove("chat2")

	// Check if the filter still matches "chat2" after removal
	result3 := filter.Check(event2)
	assert.False(t, result3, "ChatFilter should not match event2 after removal")
}

func TestRegexFilter_Check(t *testing.T) {
	pattern := `https?://\S+\.(?:png|jpe?g|gif|bmp|svg)`
	filter := NewRegexFilter(pattern)
	event1 := &Event{Type: OnMessage, Message: &Message{Text: "Lorem ipsum dolor sit amet https://i.imgur.com/Ag8wg1F.png"}}
	event2 := &Event{Type: OnMessage, Message: &Message{Text: "Dolor sit amet"}}

	// Check if the filter matches event1
	result1 := filter.Check(event1)
	assert.True(t, result1, "RegexFilter should match event1")

	// Check if the filter matches event2
	result2 := filter.Check(event2)
	assert.False(t, result2, "RegexFilter should not match event2")
}

func TestCombineFilter_And(t *testing.T) {
	userFilter := NewUserFilter("user1")
	chatFilter := NewChatFilter("chat1")

	// Create an AND combination of the filters
	filter := userFilter.And(chatFilter)

	event1 := &Event{
		User:  &User{Name: "user1"},
		Group: &Group{Name: "chat1"},
	}

	event2 := &Event{
		User:  &User{Name: "user2"},
		Group: &Group{Name: "chat2"},
	}

	// Check if the combined filter matches event1
	result1 := filter.Check(event1)
	assert.True(t, result1, "Combined filter should match event1")

	// Check if the combined filter matches event2
	result2 := filter.Check(event2)
	assert.False(t, result2, "Combined filter should not match event2")
}

func TestCombineFilter_Or(t *testing.T) {
	userFilter := NewUserFilter("user1")
	chatFilter := NewChatFilter("chat1")

	// Create an OR combination of the filters
	filter := userFilter.Or(chatFilter)

	event1 := &Event{
		Type:  OnMessage,
		User:  &User{Name: "user1"},
		Group: &Group{Name: "chat1"},
	}

	event2 := &Event{
		Type:  OnMessage,
		User:  &User{Name: "user2"},
		Group: &Group{Name: "chat1"},
	}

	// Check if the combined filter matches event1
	result1 := filter.Check(event1)
	assert.True(t, result1, "Combined filter should match event1")

	// Check if the combined filter matches event2
	result2 := filter.Check(event2)
	assert.True(t, result2, "Combined filter should match event2")
}

func TestCombineFilter_Xor(t *testing.T) {
	userFilter := NewUserFilter("user1")
	chatFilter := NewChatFilter("chat1")

	// Create an XOR combination of the filters
	filter := userFilter.Xor(chatFilter)

	event1 := &Event{
		Type:  OnMessage,
		User:  &User{Name: "user1"},
		Group: &Group{Name: "chat1"},
	}

	event2 := &Event{
		Type:  OnMessage,
		User:  &User{Name: "user2"},
		Group: &Group{Name: "chat1"},
	}

	// Check if the combined filter matches event1
	result1 := filter.Check(event1)
	assert.False(t, result1, "Combined filter should not match event1")

	// Check if the combined filter matches event2
	result2 := filter.Check(event2)
	assert.True(t, result2, "Combined filter should match event2")
}

func TestCombineFilter_Not(t *testing.T) {
	userFilter := NewUserFilter("user1")

	// Create a NOT filter
	notFilter := userFilter.Not()

	event1 := &Event{User: &User{Name: "user1"}}
	event2 := &Event{User: &User{Name: "user2"}}

	// Check if the NOT filter matches event1
	result1 := notFilter.Check(event1)
	assert.False(t, result1, "NOT filter should not match event1")

	// Check if the NOT filter matches event2
	result2 := notFilter.Check(event2)
	assert.True(t, result2, "NOT filter should match event2")
}
