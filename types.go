package chadango

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
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

// KeyValue represents a key-value pair.
// This is a struct helper for "emod" ModAction parsing.
type KeyValue struct {
	Key   string
	Value int64
}

// MessageBackground represents the message background information of a user.
type MessageBackground struct {
	Align        string `xml:"align,attr"`  // Background image alignment
	Alpha        int    `xml:"bgalp,attr"`  // Background color transparency
	Color        string `xml:"bgc,attr"`    // Background color
	HasRecording int64  `xml:"hasrec,attr"` // Media recording timestamp (ms)
	ImageAlpha   int    `xml:"ialp,attr"`   // Background image transparency
	IsVid        bool   `xml:"isvid,attr"`  // Is the Media a video?
	Tile         bool   `xml:"tile,attr"`   // Tile image?
	UseImage     bool   `xml:"useimg,attr"` // Use image?

	username string
}

// GetForm returns the URL-encoded form values for the `MessageBackground`.
func (mb *MessageBackground) GetForm() url.Values {
	switch mb.Align {
	case "tr", "br", "tl", "bl":
	default:
		mb.Align = "tl"
	}
	mb.Alpha = Min(100, Max(0, mb.Alpha))
	if mb.Color == "" {
		mb.Color = "ffffff"
	}
	mb.ImageAlpha = Min(100, Max(0, mb.ImageAlpha))

	form := url.Values{
		"align":  {mb.Align},
		"bgalp":  {strconv.Itoa(mb.Alpha)},
		"bgc":    {mb.Color},
		"hasrec": {strconv.FormatInt(mb.HasRecording, 10)},
		"ialp":   {strconv.Itoa(mb.ImageAlpha)},
		"isvid":  {BoolZeroOrOne(mb.IsVid)},
		"tile":   {BoolZeroOrOne(mb.Tile)},
		"useimg": {BoolZeroOrOne(mb.UseImage)},
	}

	return form
}

func (mb *MessageBackground) GetImageURL() string {
	return UsernameToURL(API_MSG_BG_IMG, mb.username)
}

// MessageStyle represents the style settings for a message.
type MessageStyle struct {
	FontFamily    string `json:"fontFamily"`    // The font family used for the message text.
	FontSize      string `json:"fontSize"`      // The font size used for the message text.
	Bold          bool   `json:"bold"`          // A boolean value indicating whether the message text should be displayed in bold.
	StylesOn      bool   `json:"stylesOn"`      // A boolean value indicating whether the message styles are enabled.
	UseBackground string `json:"usebackground"` // The background color used for the message text.
	Italics       bool   `json:"italics"`       // A boolean value indicating whether the message text should be displayed in italics.
	TextColor     string `json:"textColor"`     // The color used for the message text.
	Underline     bool   `json:"underline"`     // A boolean value indicating whether the message text should be underlined.
	NameColor     string `json:"nameColor"`     // The color used for the username or sender's name in the message.
}

// GetForm returns the URL-encoded form values for the `MessageStyle`.
func (mb MessageStyle) GetForm() url.Values {
	form := url.Values{}
	configType := reflect.TypeOf(mb)
	configValue := reflect.ValueOf(mb)
	var (
		field reflect.StructField
		value reflect.Value
		tag   string
	)

	for i := 0; i < configType.NumField(); i++ {
		field = configType.Field(i)
		value = configValue.Field(i)
		tag = field.Tag.Get("json")

		form.Set(tag, fmt.Sprintf("%v", value.Interface()))
	}

	return form
}

type UploadedImage struct {
	ID       int
	Username string
}

func (i UploadedImage) ThumbURL() string {
	return fmt.Sprintf(UsernameToURL(API_UM_THUMB, i.Username), i.ID)
}

func (i UploadedImage) LargeURL() string {
	return fmt.Sprintf(UsernameToURL(API_UM_LARGE, i.Username), i.ID)
}

func (i UploadedImage) MessageEmbed() string {
	return fmt.Sprintf("img%d", i.ID)
}

// MiniProfile represents a mini profile of a user.
type MiniProfile struct {
	username string
	XMLName  xml.Name     `xml:"mod"`  // Tag name
	Body     QueryEscaped `xml:"body"` // Mini profile info
	Gender   string       `xml:"s"`    // Gender (M, F)
	Birth    BirthDate    `xml:"b"`    // Date of birth (yyyy-mm-dd)
	Location Location     `xml:"l"`    // Location
	Premium  PremiumDate  `xml:"d"`    // Premium expiration
}

func (m MiniProfile) PhotoLargeURL() string {
	return UsernameToURL(API_PHOTO_FULL_IMG, m.username)
}

// QueryEscaped represents a query-escaped string.
type QueryEscaped string

// UnmarshalXML unmarshals the XML data into the QueryEscaped value.
func (c *QueryEscaped) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var rawText string
	if err := d.DecodeElement(&rawText, &start); err != nil {
		return err
	}

	parsedText, _ := url.QueryUnescape(rawText)

	*c = QueryEscaped(parsedText)
	return nil
}

// BirthDate represents a birth date of a user.
type BirthDate time.Time

// UnmarshalXML unmarshals the XML data into the BirthDate value.
func (c *BirthDate) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var rawDate string
	if err := d.DecodeElement(&rawDate, &start); err != nil {
		return err
	}

	parsedDate, _ := time.Parse("2006-01-02", rawDate)

	*c = BirthDate(parsedDate)
	return nil
}

// Location represents the location information of a user.
type Location struct {
	Country   string  `xml:"c,attr"`    // Country name or US postal code
	G         string  `xml:"g,attr"`    // Reserved
	Latitude  float64 `xml:"lat,attr"`  // Latitude
	Longitude float64 `xml:"lon,attr"`  // Longitude
	Text      string  `xml:",chardata"` // String text of the location
}

// PremiumDate represents a premium date of a user.
type PremiumDate time.Time

// UnmarshalXML unmarshals the XML data into the PremiumDate value.
func (c *PremiumDate) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var rawDate string
	if err := d.DecodeElement(&rawDate, &start); err != nil {
		return err
	}

	parsedTimestamp, _ := ParseTime(rawDate)

	*c = PremiumDate(parsedTimestamp)
	return nil
}

// FullProfile represents a full profile of a user.
type FullProfile struct {
	XMLName xml.Name     `xml:"mod"`  // Tag name
	Body    QueryEscaped `xml:"body"` // Full profile info
	T       string       `xml:"t"`    // Reserved
}

// PeopleQuery represents a query for searching people.
type PeopleQuery struct {
	AgeFrom    int    // Minimum age
	AgeTo      int    // Maximum age
	Gender     string // Gender (B: both, M: male, F: female, N: unset)
	Username   string // Username
	Radius     int    // Radius
	Latitude   string // Latitude
	Longtitude string // Longitude
	Online     bool   // Online status
	Offset     int    // Offset
	Amount     int    // Amount
}

// GetForm returns the URL-encoded form values for the PeopleQuery.
func (pq *PeopleQuery) GetForm() url.Values {
	pq.AgeFrom = Min(99, Max(0, pq.AgeFrom))
	pq.AgeTo = Min(99, Max(0, pq.AgeTo))

	switch pq.Gender {
	case "B", "M", "F", "N":
	default:
		pq.Gender = "B"
	}

	pq.Radius = Min(9999, Max(0, pq.Radius))

	form := url.Values{
		"ami": {strconv.Itoa(pq.AgeFrom)},
		"ama": {strconv.Itoa(pq.AgeTo)},
		"s":   {pq.Gender},
	}

	if pq.Username != "" {
		form.Set("ss", pq.Username)
	}
	if pq.Radius > 0 {
		form.Set("r", strconv.Itoa(pq.Radius))
	}
	if pq.Latitude != "" && pq.Longtitude != "" {
		form.Set("la", pq.Latitude)
		form.Set("lo", pq.Longtitude)
	}
	if pq.Online {
		form.Set("o", "1")
	}

	form.Set("h5", "1")
	form.Set("f", strconv.Itoa(pq.Offset))
	form.Set("t", strconv.Itoa(pq.Offset+pq.Amount))

	return form
}

// NextOffset updates the offset to retrieve the next set of results.
func (pq *PeopleQuery) NextOffset() {
	pq.Offset += pq.Amount
}

type PeopleResult struct {
	Username string
	IsOnline bool
}
