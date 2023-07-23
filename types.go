package chadango

import (
	"net/url"
	"strings"
	"time"
)

// Number represents a numeric type.
type Number interface {
	int | int16 | int32 | int64 | float32 | float64
}

// Config represents a configuration object.
type Config struct {
	Username  string   `json:"username"`  // Username of the configuration.
	Password  string   `json:"password"`  // Password of the configuration.
	AnonName  string   `json:"anonname"`  // Anonymous name of the configuration.
	Groups    []string `json:"groups"`    // List of groups in the configuration.
	NameColor string   `json:"namecolor"` // Name color in the configuration.
	TextColor string   `json:"textcolor"` // Text color in the configuration.
	TextFont  string   `json:"textfont"`  // Text font in the configuration.
	TextSize  int      `json:"textsize"`  // Text size in the configuration.
	SessionID string   `json:"sessionid"` // Session ID in the configuration.
	EnableBG  bool     `json:"enablebg"`  // Enable background in the configuration.
	EnablePM  bool     `json:"enablepm"`  // Enable private messages in the configuration.
	Debug     bool     `json:"debug"`     // Debug mode in the configuration.
	Prefix    string   `json:"prefix"`    // Prefix for commands in the configuration.
}

// User represents a user object.
type User struct {
	Name   string // Name of the user.
	IsSelf bool   // Indicates if the user is self.
	IsAnon bool   // Indicates if the user is anonymous.
}

// Participant represents a group participant object.
type Participant struct {
	ParticipantID string    // Participant ID.
	UserID        int       // User ID.
	User          *User     // User object.
	Time          time.Time // Time of participation.
}

// Blocked represents a banned user in a group.
type Blocked struct {
	IP           string    // IP address of the blocked user.
	ModerationID string    // Moderation ID of the block.
	Target       string    // Target username.
	Blocker      string    // Username of the blocker.
	Time         time.Time // Time of blocking.
}

// Unblocked represents an unblocked user in a group.
type Unblocked struct {
	IP           string    // IP address of the unblocked user.
	ModerationID string    // Moderation ID of the unblock.
	Target       string    // Target username.
	Unblocker    string    // Username of the unblocker.
	Time         time.Time // Time of unblocking.
}

// UserStatus represents the online status of a user.
type UserStatus struct {
	User *User         // User object.
	Time time.Time     // Time of the status.
	Info string        // Additional information about the status.
	Idle time.Duration // Idle duration of the user.
}

// PrivateSetting represents the private message settings.
type PrivateSetting struct {
	DisableIdleTime bool // Disable idle time in private messages.
	AllowAnon       bool // Allow anonymous messages.
	EmailOfflineMsg bool // Email offline messages.
}

// BanWord represents a banned word.
type BanWord struct {
	WholeWords string `json:"wholeWords"` // Whole words to be banned.
	Words      string `json:"words"`      // Words to be banned.
}

// SetWhole sets the whole banned words.
func (bw *BanWord) SetWhole(exact []string) {
	bw.WholeWords = url.QueryEscape(strings.Join(exact, ","))
}

// GetWhole returns the whole banned words as a slice of strings.
func (bw *BanWord) GetWhole() []string {
	unquoted, _ := url.QueryUnescape(bw.WholeWords)
	return strings.Split(unquoted, ",")
}

// SetPartial sets the partial banned words.
func (bw *BanWord) SetPartial(partial []string) {
	bw.Words = url.QueryEscape(strings.Join(partial, ","))
}

// GetPartial returns the partial banned words as a slice of strings.
func (bw *BanWord) GetPartial() []string {
	unquoted, _ := url.QueryUnescape(bw.Words)
	return strings.Split(unquoted, ",")
}

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

// KeyValue represents a key-value pair
// This is a struct helper for "emod" ModAction parsing.
type KeyValue struct {
	Key   string
	Value int64
}
