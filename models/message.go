package models

import (
	"regexp"
	"strconv"
	"time"
)

type MessageChannel int64

const (
	_ MessageChannel = 1 << iota
	_
	FlagPremium
	FlagBackground
	FlagMedia
	FlagCensored
	FlagModIcon
	FlagStaffIcon
	FlagRedChannel    // js: #ed1c24
	FlagOrangeChannel // js: #ee7f22
	FlagGreenChannel  // js: #39b54a
	FlagBlueChannel   // js: #25aae1
	FlagAzureChannel  // js: #0e76bc
	FlagPurpleChannel // js: #662d91
	FlagPinkChannel   // js: #ed217c
	FlagModChannel
)

var (
	NameColorRe        = regexp.MustCompile(`<n([\da-fA-F]{1,6})\/>`)
	FontStyleRe        = regexp.MustCompile(`<f x([\da-fA-F]+)?="([\d\w]+)?">`)
	PrivateFontStyleRe = regexp.MustCompile(`<g x(\d+)?s([\da-fA-F]+)?="([\d\w]+)?">`)
)

const (
	DEFAULT_COLOR     = "000"
	DEFAULT_TEXT_FONT = "1"
	DEFAULT_TEXT_SIZE = 11
)

// Message represents a message that comes from a group or private chat.
//
// It provides methods for replying to the message, deleting the message, deleting all messages from the sender,
// banning the user who sent the message, and accessing various properties of the message.
// The message object also contains information about the sender, the content, the time of sending, and the channel flags.
type Message struct {
	IsPrivate    bool           // IsPrivate indicates whether the message is from a private chat.
	User         *User          // User represents the user who sent the message.
	Text         string         // Text contains the parsed text of the message.
	RawText      string         // RawText contains the raw text of the message.
	UserID       int            // UserID is the unique identifier of the user who sent the message.
	ID           string         // ID is the unique identifier of the message.
	ModerationID string         // ModerationID is the unique identifier used for moderation actions on the message.
	UserIP       string         // UserIP is the IP address of the user who sent the message.
	Time         time.Time      // Time represents the time when the message was sent.
	ReceivedTime time.Time      // ReceivedTime represents the time when the message was received.
	Flag         MessageChannel // Flag represents the flag value associated with the message.
	FromSelf     bool           // FromSelf indicates whether the message was sent by the current user.
	FromAnon     bool           // FromAnon indicates whether the message was sent by an anonymous user.
	AnonSeed     int            // AnonSeed represents the seed value used for anonymous user identification.
}

// Channel returns the channel flag of the message.
//
// It combines the individual channel flags (e.g., [FlagRedChannel], [FlagOrangeChannel], etc.) into a single [MessageChannel] value.
//
// Returns:
//   - MessageChannel: The combined channel flag value.
func (m *Message) Channel() (c MessageChannel) {
	if m.Flag&FlagRedChannel != 0 {
		c |= FlagRedChannel
	}
	if m.Flag&FlagOrangeChannel != 0 {
		c |= FlagOrangeChannel
	}
	if m.Flag&FlagGreenChannel != 0 {
		c |= FlagGreenChannel
	}
	if m.Flag&FlagBlueChannel != 0 {
		c |= FlagBlueChannel
	}
	if m.Flag&FlagAzureChannel != 0 {
		c |= FlagAzureChannel
	}
	if m.Flag&FlagPurpleChannel != 0 {
		c |= FlagPurpleChannel
	}
	if m.Flag&FlagPinkChannel != 0 {
		c |= FlagPinkChannel
	}

	return
}

// HasPremium checks if the message has premium flag.
//
// Returns:
//   - bool: True if the message has the premium flag, otherwise false.
func (m *Message) HasPremium() bool {
	return m.Flag&FlagPremium != 0
}

// HasBackground checks if the message has background flag.
//
// Returns:
//   - bool: True if the message has the background flag, otherwise false.
func (m *Message) HasBackground() bool {
	return m.Flag&FlagBackground != 0
}

// HasMedia checks if the message has media flag.
//
// Returns:
//   - bool: True if the message has the media flag, otherwise false.
func (m *Message) HasMedia() bool {
	return m.Flag&FlagMedia != 0
}

// IsCensored checks if the message contains censor.
//
// Returns:
//   - bool: True if the message is censored, otherwise false.
func (m *Message) IsCensored() bool {
	return m.Flag&FlagCensored != 0
}

// HasModIcon checks if the message has mod icon flag.
//
// Returns:
//   - bool: True if the message has the mod icon flag, otherwise false.
func (m *Message) HasModIcon() bool {
	return m.Flag&FlagModIcon != 0
}

// HasStaffIcon checks if the message has staff icon flag.
//
// Returns:
//   - bool: True if the message has the staff icon flag, otherwise false.
func (m *Message) HasStaffIcon() bool {
	return m.Flag&FlagStaffIcon != 0
}

// IsInModChannel checks if the message is in a mod channel.
//
// Returns:
//   - bool: True if the message is in a mod channel, otherwise false.
func (m *Message) IsInModChannel() bool {
	return m.Flag&FlagModChannel != 0
}

// NameColor returns the name color of the message.
//
// It extracts the name color from the raw text of the message.
//
// Returns:
//   - string: The name color of the message.
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
//
// It extracts the text color, font, and size from the raw text of the message.
//
// Returns:
//   - string: The text color of the message.
//   - string: The text font of the message.
//   - int: The text size of the message.
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
