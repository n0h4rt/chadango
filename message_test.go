package chadango

import (
	"testing"
	"time"

	"github.com/n0h4rt/chadango/models"
	"github.com/n0h4rt/chadango/utils"
	"github.com/stretchr/testify/assert"
)

func TestParseGroupMessage(t *testing.T) {
	group := &Group{
		LoginName: "Nekonyan",
		UserID:    48875733,
	}
	tests := []struct {
		name string
		data string
		want *models.Message
	}{
		{
			name: "NormalMessage",
			data: "1717866894:Nekonyan::48875733:moderationID:messageID:userIP:2392::<n33FFFF/><f x11000=\"century gothic\">Press F in the chat",
			want: &models.Message{
				Time:         time.Unix(1717866894, 0),
				UserID:       48875733,
				ModerationID: "moderationID",
				ID:           "messageID",
				UserIP:       "userIP",
				RawText:      "<n33FFFF/><f x11000=\"century gothic\">Press F in the chat",
				Text:         "Press F in the chat",
				User:         &models.User{Name: "Nekonyan", IsSelf: true},
				FromSelf:     true,
			},
		},
		{
			name: "NamedAnonMessage",
			data: "1721913578.62::anonName:23361675:moderationID:messageID:userIP:0::<n3512/>asdfghjkl",
			want: &models.Message{
				Time:         time.Unix(1721913578, int64(.62*1e6)),
				UserID:       23361675,
				ModerationID: "moderationID",
				ID:           "messageID",
				UserIP:       "userIP",
				RawText:      "<n3512/>asdfghjkl",
				Text:         "asdfghjkl",
				User:         &models.User{Name: "anonName", IsAnon: true},
				FromAnon:     true,
			},
		},
		{
			name: "UnnamedAnonMessage",
			data: "1721913578.62:::23361675:moderationID:messageID:userIP:0::<n3512/>asdfghjkl",
			want: &models.Message{
				Time:         time.Unix(1721913578, int64(.62*1e6)),
				UserID:       23361675,
				ModerationID: "moderationID",
				ID:           "messageID",
				UserIP:       "userIP",
				RawText:      "<n3512/>asdfghjkl",
				Text:         "asdfghjkl",
				User:         &models.User{Name: utils.GetAnonName(3512, 23361675), IsAnon: true},
				FromAnon:     true,
				AnonSeed:     3512,
			},
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
			assert.Equal(t, tt.want.User.Name, got.User.Name)
			assert.Equal(t, tt.want.User.IsAnon, got.User.IsAnon)
			assert.Equal(t, tt.want.User.IsSelf, got.User.IsSelf)
			assert.Equal(t, tt.want.FromAnon, got.FromAnon)
			assert.Equal(t, tt.want.FromSelf, got.FromSelf)
		})
	}
}

func TestParsePrivateMessage(t *testing.T) {
	private := &Private{
		LoginName: "Nekonyan",
	}
	tests := []struct {
		name string
		data string
		want *models.Message
	}{
		{
			name: "NormalMessage",
			data: "clonerxyz:clonerxyz:unknown:1723029464.85:0:<m v=\"1\">text</m>",
			want: &models.Message{
				Time:      time.Unix(1723029464, 850000),
				ID:        "1723029464",
				RawText:   "<m v=\"1\">text</m>",
				Text:      "text",
				User:      &models.User{Name: "clonerxyz"},
				IsPrivate: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParsePrivateMessage(tt.data, private)
			assert.Equal(t, tt.want.Time, got.Time)
			assert.Equal(t, tt.want.ID, got.ID)
			assert.Equal(t, tt.want.RawText, got.RawText)
			assert.Equal(t, tt.want.Text, got.Text)
			assert.Equal(t, tt.want.User.Name, got.User.Name)
			assert.Equal(t, tt.want.IsPrivate, got.IsPrivate)
		})
	}
}

func TestParseAnnouncement(t *testing.T) {
	group := &Group{
		Name: "testgroup",
	}
	tests := []struct {
		name string
		data string
		want *models.Message
	}{
		{
			name: "NormalAnnouncement",
			data: "testgroup:1688488704:This is an announcement",
			want: &models.Message{RawText: "This is an announcement", Text: "This is an announcement"},
		},
		{
			name: "AnnouncementWithHTML",
			data: "testgroup:1688488704:<font color=\"#FF0000\">This is an announcement</font>",
			want: &models.Message{RawText: "<font color=\"#FF0000\">This is an announcement</font>", Text: "This is an announcement"},
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
