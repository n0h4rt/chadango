package chadango

import (
	"html"
	"strconv"
	"strings"
	"time"
)

// Message represents a message that comes from a group or private chat.
type Message struct {
	Group        *Group    // Group represents the group chat where the message originated (nil if private chat).
	Private      *Private  // Private represents the private chat where the message originated (nil if group chat).
	IsPrivate    bool      // IsPrivate indicates whether the message is from a private chat.
	User         *User     // User represents the user who sent the message.
	Text         string    // Text contains the parsed text of the message.
	RawText      string    // RawText contains the raw text of the message.
	UserID       int       // UserID is the unique identifier of the user who sent the message.
	ID           string    // ID is the unique identifier of the message.
	ModerationID string    // ModerationID is the unique identifier used for moderation actions on the message.
	UserIP       string    // UserIP is the IP address of the user who sent the message.
	Time         time.Time // Time represents the time when the message was sent.
	ReceivedTime time.Time // ReceivedTime represents the time when the message was received.
	Flag         int64     // Flag represents the flag value associated with the message.
	FromSelf     bool      // FromSelf indicates whether the message was sent by the current user.
	FromAnon     bool      // FromAnon indicates whether the message was sent by an anonymous user.
	AnonSeed     int       // AnonSeed represents the seed value used for anonymous user identification.
}

// Channel returns the channel flag of the message.
func (m *Message) Channel() int64 {
	var c int64
	if m.Flag&FlagModIcon != 0 {
		c |= FlagModIcon
	}
	if m.Flag&FlagStaffIcon != 0 {
		c |= FlagStaffIcon
	}
	if m.Flag&FlagRedChannel != 0 {
		c |= FlagRedChannel
	}
	if m.Flag&FlagBlueChannel != 0 {
		c |= FlagBlueChannel
	}
	if m.Flag&FlagModChannel != 0 {
		c |= FlagModChannel
	}
	return c
}

// HasPremium checks if the message has premium flag.
func (m *Message) HasPremium() bool {
	return m.Flag&FlagPremium != 0
}

// HasBackground checks if the message has background flag.
func (m *Message) HasBackground() bool {
	return m.Flag&FlagBackground != 0
}

// HasMedia checks if the message has media flag.
func (m *Message) HasMedia() bool {
	return m.Flag&FlagMedia != 0
}

// NameColor returns the name color of the message.
func (m *Message) NameColor() string {
	reResult := NameColorRe.FindStringSubmatch(m.RawText)
	if len(reResult) > 1 {
		switch len(reResult[1]) {
		case 1, 3, 6:
			return reResult[1]
		}
	}
	return DEFAULT_COLOR
}

// TextStyle returns the text style of the message.
func (m *Message) TextStyle() (textColor, textFont string, textSize int) {
	textColor = DEFAULT_COLOR
	textFont = DEFAULT_TEXT_FONT
	textSize = DEFAULT_TEXT_SIZE

	if !m.IsPrivate {
		reResult := FontStyleRe.FindStringSubmatch(m.RawText)
		if len(reResult) > 2 {
			switch len(reResult[1]) {
			case 5, 8:
				textSize, _ = strconv.Atoi(reResult[1][:2])
				textColor = reResult[1][2:]
			case 3, 6:
				textColor = reResult[1]
			case 2:
				textSize, _ = strconv.Atoi(reResult[1][:2])
			}
			if reResult[2] != "" {
				textFont = reResult[2]
			}
		}
		return
	}
	reResult := PrivateFontStyleRe.FindStringSubmatch(m.RawText)
	if len(reResult) > 3 {
		if reResult[1] != "" {
			textSize, _ = strconv.Atoi(reResult[1])
		}
		if reResult[2] != "" {
			textColor = reResult[2]
		}
		if reResult[3] != "" {
			textFont = reResult[3]
		}
	}
	return
}

// Reply sends a reply message to the sender of the current message.
func (m *Message) Reply(text string) (*Message, error) {
	if !m.IsPrivate {
		return m.Group.SendMessage(text)
	}
	return nil, m.Private.SendMessage(m.User.Name, text)
}

func (m *Message) Delete() error {
	if !m.IsPrivate {
		return m.Group.Delete(m)
	}
	return nil
}

func (m *Message) DeleteAll() error {
	if !m.IsPrivate {
		return m.Group.DeleteAll(m)
	}
	return nil
}

func (m *Message) BanUser() error {
	if !m.IsPrivate {
		return m.Group.BanUser(m)
	}
	return nil
}

// ParseGroupMessage parses a group message data and returns a Message object.
func ParseGroupMessage(data string, group *Group) *Message {
	var (
		user     *User
		reResult []string
	)
	fields := strings.SplitN(data, ":", 10)
	msg := &Message{Group: group, ReceivedTime: time.Now()}
	msg.Time, _ = ParseTime(fields[0])
	msg.UserID, _ = strconv.Atoi(fields[3])
	if fields[1] != "" {
		user = &User{Name: fields[1]}
	} else if fields[2] != "" {
		user = &User{Name: fields[2], IsAnon: true}
		msg.FromAnon = true
	} else {
		if reResult = AnonSeedRe.FindAllString(fields[9], 1); len(reResult) > 0 {
			msg.AnonSeed, _ = strconv.Atoi(reResult[0][2:6])
		}
		user = &User{Name: GetAnonName(msg.AnonSeed, msg.UserID), IsAnon: true}
		msg.FromAnon = true
	}
	msg.User = user
	// Apparently, the `group.SessionID` can be a zero (uninitialized),
	// as this function is called before the "OK" event is received.
	// As a temporary workaround, the condition `fields[3] == group.SessionID[:8]`
	// will be treated as `true` to handle this situation.
	if group.UserID == 0 {
		user.IsSelf = strings.EqualFold(user.Name, group.LoginName)
	} else {
		user.IsSelf = msg.UserID == group.UserID && strings.EqualFold(user.Name, group.LoginName)
	}
	msg.FromSelf = user.IsSelf
	msg.ModerationID = fields[4]
	msg.ID = fields[5]
	msg.UserIP = fields[6]
	msg.Flag, _ = strconv.ParseInt(fields[7], 10, 64)
	// _ = fields[8]  // Omitted for now
	msg.RawText = fields[9]
	text := strings.Replace(fields[9], "<br/>", "\n", -1)
	text = HtmlTagRe.ReplaceAllString(text, "")
	text = html.UnescapeString(text)
	msg.Text = strings.Replace(text, "\n", "<br/>", -1)
	return msg
}

// ParsePrivateMessage parses a private message data and returns a Message object.
func ParsePrivateMessage(data string, private *Private) *Message {
	fields := strings.SplitN(data, ":", 6)
	msg := &Message{Private: private, IsPrivate: true}
	// It should not be possible to send a private message to oneself.
	// isSelf := strings.EqualFold(fields[0], private.LoginName)
	msg.User = &User{Name: fields[0]}
	msg.Time, _ = ParseTime(fields[3])
	msg.ID, _, _ = strings.Cut(fields[3], ".") // A fake ID used as the key in `p.Messages`.
	msg.Flag, _ = strconv.ParseInt(fields[4], 10, 64)
	msg.RawText = fields[5]
	text := strings.Replace(fields[5], "<br/>", "\n", -1)
	text = HtmlTagRe.ReplaceAllString(text, "")
	text = html.UnescapeString(text)
	msg.Text = strings.Replace(text, "\n", "<br/>", -1)
	return msg
}

// ParseAnnouncement parses a group announcement data and returns a Message object.
func ParseAnnouncement(data string, group *Group) *Message {
	fields := strings.SplitN(data, ":", 3)
	text := HtmlTagRe.ReplaceAllString(fields[2], "")
	text = html.UnescapeString(text)
	return &Message{
		Group:        group,
		ReceivedTime: time.Now(),
		RawText:      fields[2],
		Text:         text,
	}
}
