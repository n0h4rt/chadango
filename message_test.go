package chadango

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseGroupMessage(t *testing.T) {
	group := &Group{
		LoginName: "testuser",
		UserID:    12345,
	}
	tests := []struct {
		name     string
		data     string
		want     *Message
		wantUser *User
	}{
		{
			name: "NormalMessage",
			data: "1688488704:testuser:12345:12345:moderationID:messageID:userIP:0:rawText",
			want: &Message{
				Time:         time.Unix(1688488704, 0),
				UserID:       12345,
				ModerationID: "moderationID",
				ID:           "messageID",
				UserIP:       "userIP",
				RawText:      "rawText",
				Text:         "rawText",
				User:         &User{Name: "testuser", IsSelf: true},
				FromSelf:     true,
			},
			wantUser: &User{Name: "testuser", IsSelf: true},
		},
		{
			name: "AnonymousMessage",
			data: "1688488704::anonName:12345:moderationID:messageID:userIP:0:rawText",
			want: &Message{
				Time:         time.Unix(1688488704, 0),
				UserID:       12345,
				ModerationID: "moderationID",
				ID:           "messageID",
				UserIP:       "userIP",
				RawText:      "rawText",
				Text:         "rawText",
				User:         &User{Name: "anonName", IsAnon: true},
				FromAnon:     true,
			},
			wantUser: &User{Name: "anonName", IsAnon: true},
		},
		{
			name: "AnonymousMessageWithSeed",
			data: "1688488704::12345:12345:moderationID:messageID:userIP:0:rawText:anonSeed=1234",
			want: &Message{
				Time:         time.Unix(1688488704, 0),
				UserID:       12345,
				ModerationID: "moderationID",
				ID:           "messageID",
				UserIP:       "userIP",
				RawText:      "rawText:anonSeed=1234",
				Text:         "rawText",
				User:         &User{Name: GetAnonName(1234, 12345), IsAnon: true},
				FromAnon:     true,
				AnonSeed:     1234,
			},
			wantUser: &User{Name: GetAnonName(1234, 12345), IsAnon: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseGroupMessage(tt.data, group)
			assert.Equal(t, tt.want.Time, got.Time)
			assert.Equal(t, tt.want.UserID, got.UserID)
			assert.Equal(t, tt.want.ModerationID, got.ModerationID)
			assert.Equal(t, tt.want.ID, got.ID)
			assert.Equal(t, tt.want.UserIP, got.UserIP)
			assert.Equal(t, tt.want.RawText, got.RawText)
			assert.Equal(t, tt.want.Text, got.Text)
			assert.Equal(t, tt.wantUser.Name, got.User.Name)
			assert.Equal(t, tt.wantUser.IsAnon, got.User.IsAnon)
			assert.Equal(t, tt.wantUser.IsSelf, got.User.IsSelf)
			assert.Equal(t, tt.want.FromAnon, got.FromAnon)
			assert.Equal(t, tt.want.FromSelf, got.FromSelf)
		})
	}
}

func TestParsePrivateMessage(t *testing.T) {
	private := &Private{
		LoginName: "testuser",
	}
	tests := []struct {
		name     string
		data     string
		want     *Message
		wantUser *User
	}{
		{
			name: "NormalMessage",
			data: "testuser:1688488704:12345.6789:0:rawText",
			want: &Message{
				Time:      time.Unix(1688488704, 0),
				ID:        "12345",
				RawText:   "rawText",
				Text:      "rawText",
				User:      &User{Name: "testuser"},
				IsPrivate: true,
			},
			wantUser: &User{Name: "testuser"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParsePrivateMessage(tt.data, private)
			assert.Equal(t, tt.want.Time, got.Time)
			assert.Equal(t, tt.want.ID, got.ID)
			assert.Equal(t, tt.want.RawText, got.RawText)
			assert.Equal(t, tt.want.Text, got.Text)
			assert.Equal(t, tt.wantUser.Name, got.User.Name)
			assert.Equal(t, tt.want.IsPrivate, got.IsPrivate)
		})
	}
}

func TestParseAnnouncement(t *testing.T) {
	group := &Group{
		LoginName: "testgroup",
	}
	tests := []struct {
		name     string
		data     string
		want     *Message
		wantText string
	}{
		{
			name:     "NormalAnnouncement",
			data:     "testgroup:1688488704:This is an announcement",
			want:     &Message{RawText: "This is an announcement", Text: "This is an announcement"},
			wantText: "This is an announcement",
		},
		{
			name:     "AnnouncementWithHTML",
			data:     "testgroup:1688488704:<font color=\"#FF0000\">This is an announcement</font>",
			want:     &Message{RawText: "<font color=\"#FF0000\">This is an announcement</font>", Text: "This is an announcement"},
			wantText: "This is an announcement",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseAnnouncement(tt.data, group)
			assert.Equal(t, tt.want.RawText, got.RawText)
			assert.Equal(t, tt.want.Text, got.Text)
		})
	}
}
