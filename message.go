package chadango

import (
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/n0h4rt/chadango/models"
	"github.com/n0h4rt/chadango/utils"
)

var (
	AnonSeedRe = regexp.MustCompile(`<n\d{4}/>`)

	// Go does not support negative lookahead `<(?!br\s*\/?>).*?>`.
	// This alternative will match either the `<br>` and `<br/>` tags (captured in group 1)
	// or any other HTML tags (captured in group 2).
	// Then the `ReplaceAllString(text, "$1")` method will then keep the content matched by group 1
	// and remove the content matched by group 2.
	HtmlTagRe = regexp.MustCompile(`(<br\s*\/?>)|(<[^>]+>)`)
)

type Message struct {
	Group   *Group   // Group represents the group chat where the message originated (nil if private chat).
	Private *Private // Private represents the private chat where the message originated (nil if group chat).
	models.Message
}

// Reply sends a reply to the message with the specified text.
//
// If the message is from a group, it calls the [SendMessage] method on the associated [Group] to send the reply.
// If the message is private, it calls the [SendMessage] method on the associated [Private] to send the reply to the user.
// The text can include formatting placeholders (%s, %d, etc.), and optional arguments can be provided to fill in these placeholders.
//
// Args:
//   - text: The text of the reply message.
//   - a: Optional arguments to fill in placeholders in the message text.
//
// Returns:
//   - *Message: A pointer to the sent message (nil if the message is private).
//   - error: An error if any occurs during the message sending process.
func (m *Message) Reply(text string, a ...any) (*Message, error) {
	if !m.IsPrivate {
		return m.Group.SendMessage(text, a...)
	}

	return nil, m.Private.SendMessage(m.User.Name, text, a...)
}

// Delete deletes the message.
//
// If the message is from a group, it calls the [Delete] method on the associated [Group] to remove the message from the group chat.
// If the message is private, this method does nothing and returns nil.
//
// Returns:
//   - error: An error if any occurs during the deletion process.
func (m *Message) Delete() error {
	if !m.IsPrivate {
		return m.Group.Delete(m)
	}

	return nil
}

// DeleteAll deletes all messages from the sender of the current message.
//
// If the message is from a group, it calls the [DeleteAll] method on the associated [Group] to remove all messages from the sender in the group chat.
// If the message is private, this method does nothing and returns nil.
//
// Returns:
//   - error: An error if any occurs during the deletion process.
func (m *Message) DeleteAll() error {
	if !m.IsPrivate {
		return m.Group.DeleteAll(m)
	}

	return nil
}

// BanUser bans the user who sent the current message from the group.
//
// If the message is from a group, it calls the [BanUser] method on the associated [Group] to perform the banning action.
// If the message is private, this method does nothing and returns nil.
//
// Returns:
//   - error: An error if any occurs during the banning process.
func (m *Message) BanUser() error {
	if !m.IsPrivate {
		return m.Group.BanUser(m)
	}

	return nil
}

// ParseGroupMessage parses a group message data.
//
// It extracts information about the sender, the content, the time of sending, and the channel flags from the provided data.
//
// Args:
//   - data: The group message data.
//   - group: The group chat where the message originated.
//
// Returns:
//   - *Message: A pointer to the parsed [Message] object.
func ParseGroupMessage(data string, group *Group) *Message {
	var (
		user     *models.User
		reResult []string
	)

	fields := strings.SplitN(data, ":", 10)
	msg := &Message{Group: group}
	msg.ReceivedTime = time.Now()
	msg.Time, _ = utils.ParseTime(fields[0])
	msg.UserID, _ = strconv.Atoi(fields[3])

	if fields[1] != "" {
		user = &models.User{Name: fields[1]}
	} else if fields[2] != "" {
		user = &models.User{Name: fields[2], IsAnon: true}
		msg.FromAnon = true
	} else {
		if reResult = AnonSeedRe.FindAllString(fields[9], 1); len(reResult) > 0 {
			msg.AnonSeed, _ = strconv.Atoi(reResult[0][2:6])
		}
		user = &models.User{Name: utils.GetAnonName(msg.AnonSeed, msg.UserID), IsAnon: true}
		msg.FromAnon = true
	}
	msg.User = user

	// Apparently, the [Group.SessionID] can be a zero (uninitialized),
	// as this function is called before the "OK" event is received.
	// As a temporary workaround, the condition fields[3] == Group.SessionID[:8]
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
	flag, _ := strconv.ParseInt(fields[7], 10, 64)
	msg.Flag = models.MessageChannel(flag)
	// _ = fields[8]  // Omitted for now
	msg.RawText = fields[9]
	text := HtmlTagRe.ReplaceAllString(fields[9], "$1")
	text = strings.ReplaceAll(text, "<br/>", "\n")
	msg.Text = html.UnescapeString(text)

	return msg
}

// ParsePrivateMessage parses a private message data.
//
// It extracts information about the sender, the content, the time of sending, and the channel flags from the provided data.
//
// Args:
//   - data: The private message data.
//   - private: The private chat where the message originated.
//
// Returns:
//   - *Message: A pointer to the parsed [Message] object.
func ParsePrivateMessage(data string, private *Private) *Message {
	fields := strings.SplitN(data, ":", 6)
	msg := &Message{Private: private}
	msg.IsPrivate = true

	// It should not be possible to send a private message to oneself.
	// isSelf := strings.EqualFold(fields[0], private.LoginName)
	msg.User = &models.User{Name: fields[0]}
	msg.Time, _ = utils.ParseTime(fields[3])
	msg.ID, _, _ = strings.Cut(fields[3], ".") // A fake ID used as the key in [Private.Messages].
	flag, _ := strconv.ParseInt(fields[4], 10, 64)
	msg.Flag = models.MessageChannel(flag)
	msg.RawText = fields[5]

	text := HtmlTagRe.ReplaceAllString(fields[5], "$1")
	msg.Text = html.UnescapeString(text)

	return msg
}

// ParseAnnouncement parses a group announcement data.
//
// It extracts the announcement text from the provided data.
//
// Args:
//   - data: The group announcement data.
//   - group: The group chat where the announcement originated.
//
// Returns:
//   - *Message: A pointer to the parsed [Message] object.
func ParseAnnouncement(data string, group *Group) *Message {
	fields := strings.SplitN(data, ":", 3)
	text := HtmlTagRe.ReplaceAllString(fields[2], "$1")
	text = html.UnescapeString(text)

	msg := &Message{Group: group}
	msg.ReceivedTime = time.Now()
	msg.RawText = fields[2]
	msg.Text = text

	return msg
}
