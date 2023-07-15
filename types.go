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
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	AnonName  string   `json:"anon_name"`
	Groups    []string `json:"groups"`
	NameColor string   `json:"name_color"`
	TextColor string   `json:"text_color"`
	TextFont  string   `json:"text_font"`
	TextSize  int      `json:"text_size"`
	SessionID string   `json:"session_id"`
	EnableBG  bool     `json:"enable_bg"`
	EnablePM  bool     `json:"enable_pm"`
	Debug     bool     `json:"debug"`
	Prefix    string   `json:"prefix"`
}

// User represent a user object.
type User struct {
	Name   string
	IsSelf bool
	IsAnon bool
}

// Participant represents a group participant object.
type Participant struct {
	ParticipantID string
	ID            string
	User          *User
	Time          time.Time
}

// Blocked represents a group banned user.
type Blocked struct {
	IP           string
	ModerationID string
	Target       string
	Blocker      string
	Time         time.Time
}

// Unblocked represents a group unbanned user.
type Unblocked struct {
	IP           string
	ModerationID string
	Target       string
	Unblocker    string
	Time         time.Time
}

// UserStatus represents online status of a user.
type UserStatus struct {
	User *User
	Time time.Time
	Info string
	Idle time.Duration
}

// UserStatus represents private message settings.
type PrivateSetting struct {
	DisableIdleTime bool
	AllowAnon       bool
	EmailOfflineMsg bool
}

// BanWord represents a banned word.
type BanWord struct {
	WholeWords string `json:"wholeWords"`
	Words      string `json:"words"`
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
// This is a function helper for "emod" ModAction parsing.
type KeyValue struct {
	Key   string
	Value int64
}
