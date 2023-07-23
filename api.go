package chadango

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// API is a struct representing a compilation of various Chatango APIs.
type API struct {
	Username string
	Password string
	cookies  map[string]string // cookies stores cookies without the `url.URL` gimmick.
	client   *http.Client      // client is a client shared among requests.
	jar      *cookiejar.Jar    // jar is a `cookiejar.Jar` used for managing cookies.
}

// Transport is a custom RoundTripper implementation.
type Transport struct {
	Transport http.RoundTripper // Transport is the underlying RoundTripper.
	Headers   map[string]string // Headers contains custom headers to be added to the requests.
}

// RoundTrip executes a single HTTP request and returns its response.
// It adds custom headers to the request before performing the request using the underlying Transport.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Add custom headers to the request.
	for key, value := range t.Headers {
		req.Header.Set(key, value)
	}

	// Perform the request using the underlying Transport.
	return t.Transport.RoundTrip(req)
}

// Initialize initializes the Chatango API client and retrieves cookies from the specified username and password.
// It authenticates the client by obtaining and storing the necessary cookies.
// The retrieved cookies will be stored in the API's cookies field for external access through `api.GetCookie`.
func (api *API) Initialize() (err error) {
	if api.client == nil {
		transport := &Transport{
			Transport: http.DefaultTransport,
			Headers: map[string]string{
				"Host":       "script.st.chatango.com",
				"Origin":     "https://st.chatango.com",
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			},
		}
		if api.jar, err = cookiejar.New(nil); err != nil {
			return
		}
		api.client = &http.Client{
			Transport: transport,
			Timeout:   API_TIMEOUT,
			Jar:       api.jar,
		}
		api.cookies = make(map[string]string)
	}

	var res *http.Response

	data := url.Values{
		"user_id":     {api.Username},
		"password":    {api.Password},
		"storecookie": {"on"},
		"checkerrors": {"yes"},
	}

	if res, err = api.client.PostForm("https://chatango.com/login", data); err != nil {
		return
	}
	defer res.Body.Close()

	// This is a peculiar method of storing cookies.
	// TODO: Find a more elegant alternative.
	for _, cookie := range api.jar.Cookies(res.Request.URL) {
		api.cookies[cookie.Name] = cookie.Value
	}

	return
}

// GetCookie retrieves cookie value.
func (api *API) GetCookie(cookiename string) (cookie string, ok bool) {
	cookie, ok = api.cookies[cookiename]
	return
}

// IsGroup checks whether the specified name is a group.
func (api *API) IsGroup(groupname string) (answer bool, err error) {
	var res *http.Response

	data := url.Values{
		"name":      {groupname},
		"makegroup": {"1"},
	}

	if res, err = api.client.PostForm("https://chatango.com/checkname", data); err != nil {
		return
	}
	defer res.Body.Close()

	var body []byte
	if body, err = io.ReadAll(res.Body); err != nil {
		return
	}

	var values url.Values
	if values, err = url.ParseQuery(string(body)); err != nil {
		return
	}

	answer = values.Get("answer") == "1"

	return
}

// PeopleQuery represents a query for searching people.
type PeopleQuery struct {
	AgeFrom    int    // Minimum age
	AgeTo      int    // Maximum age
	Gender     string // Gender (B, M, F, N)
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

// SearchPeople searches for people based on the provided query.
// It returns a list of usernames and online status data.
func (api *API) SearchPeople(query PeopleQuery) (usernames [][2]string, err error) {
	var res *http.Response

	if res, err = api.client.PostForm("https://chatango.com/search", query.GetForm()); err != nil {
		return
	}
	defer res.Body.Close()

	var body []byte
	if body, err = io.ReadAll(res.Body); err != nil {
		return
	}

	if head, data, ok := strings.Cut(string(body), "="); ok && head == "h" {
		var username string
		var isonline string // Should we parse it into boolean? "0" || "1"
		for _, entry := range strings.Split(data, "") {
			username, isonline, _ = strings.Cut(entry, ";")
			usernames = append(usernames, [2]string{username, isonline})
		}
	}

	return
}

// GroupsList represents the created and recently visited chat groups of the user.
// It also provides the number of unread private messages.
type GroupsList struct {
	RecentGroups   [][2]string `json:"recent_groups"` // List of recently visited groups.
	UnreadMessages int         `json:"n_msg"`         // Number of unread private messages.
	Groups         [][2]string `json:"groups"`        // List of created groups.
}

// GetRecentGroups returns the map of recent chat groups.
func (mg GroupsList) GetRecentGroups() map[string]string {
	recent := make(map[string]string)
	for _, arr := range mg.RecentGroups {
		recent[arr[0]], _ = url.QueryUnescape(arr[1])
	}

	return recent
}

// GetGroups returns the map of all chat groups.
func (mg GroupsList) GetGroups() map[string]string {
	groups := make(map[string]string)
	for _, arr := range mg.Groups {
		groups[arr[0]], _ = url.QueryUnescape(arr[1])
	}

	return groups
}

// GetGroupList retrieves the user's chat group list.
func (api *API) GetGroupList() (groups GroupsList, err error) {
	var res *http.Response

	if res, err = api.client.PostForm("https://chatango.com/groupslistupdate", url.Values{}); err != nil {
		return
	}
	defer res.Body.Close()

	var body []byte
	if body, err = io.ReadAll(res.Body); err != nil {
		return
	}

	err = json.Unmarshal(body, &groups)

	return
}

// MiniProfile represents a mini profile of a user.
type MiniProfile struct {
	XMLName  xml.Name     `xml:"mod"`  // Tag name
	Body     QueryEscaped `xml:"body"` // Mini profile info
	Gender   string       `xml:"s"`    // Gender (M, F)
	Birth    BirthDate    `xml:"b"`    // Date of birth (yyyy-mm-dd)
	Location Location     `xml:"l"`    // Location
	Premium  PremiumDate  `xml:"d"`    // Premium expiration
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

// GetMiniProfile retrieves the mini profile of the specified username.
func (api *API) GetMiniProfile(username string) (profile MiniProfile, err error) {
	var res *http.Response

	path0 := username[0:1]
	path1 := username[0:1]
	if len(username) > 1 {
		path1 = username[1:2]
	}
	url := fmt.Sprintf("https://ust.chatango.com/profileimg/%s/%s/%s/mod1.xml", path0, path1, username)

	if res, err = api.client.Get(url); err != nil {
		return
	}
	defer res.Body.Close()

	var body []byte
	if body, err = io.ReadAll(res.Body); err != nil {
		return
	}

	err = xml.Unmarshal(body, &profile)

	return
}

// FullProfile represents a full profile of a user.
type FullProfile struct {
	XMLName xml.Name     `xml:"mod"`  // Tag name
	Body    QueryEscaped `xml:"body"` // Full profile info
	T       string       `xml:"t"`    // Reserved
}

// GetFullProfile retrieves the full profile of the specified username.
func (api *API) GetFullProfile(username string) (profile FullProfile, err error) {
	var res *http.Response

	path0 := username[0:1]
	path1 := username[0:1]
	if len(username) > 1 {
		path1 = username[1:2]
	}
	url := fmt.Sprintf("https://ust.chatango.com/profileimg/%s/%s/%s/mod2.xml", path0, path1, username)

	if res, err = api.client.Get(url); err != nil {
		return
	}
	defer res.Body.Close()

	var body []byte
	if body, err = io.ReadAll(res.Body); err != nil {
		return
	}

	err = xml.Unmarshal(body, &profile)

	return
}

// MessageBackground represents the background information of a user.
type MessageBackground struct {
	Align        string `xml:"align,attr"`  // Background image alignment
	Alpha        int    `xml:"bgalp,attr"`  // Background color transparency
	Color        string `xml:"bgc,attr"`    // Background color
	HasRecording int64  `xml:"hasrec,attr"` // Media recording timestamp (ms)
	ImageAlpha   int    `xml:"ialp,attr"`   // Background image transparency
	IsVid        bool   `xml:"isvid,attr"`  // Media is a video?
	Tile         bool   `xml:"tile,attr"`   // Tile image?
	UseImage     bool   `xml:"useimg,attr"` // Use image?
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

// GetBackground retrieves the message background of the specified username.
func (api *API) GetBackground(username string) (background MessageBackground, err error) {
	var res *http.Response

	path0 := username[0:1]
	path1 := username[0:1]
	if len(username) > 1 {
		path1 = username[1:2]
	}
	url := fmt.Sprintf("https://ust.chatango.com/profileimg/%s/%s/%s/msgbg.xml", path0, path1, username)

	if res, err = api.client.Get(url); err != nil {
		return
	}
	defer res.Body.Close()

	var body []byte
	if body, err = io.ReadAll(res.Body); err != nil {
		return
	}

	err = xml.Unmarshal(body, &background)

	return
}

// SetBackground updates the message background of the current user.
// Normally, the `background` parameter is obtained/modified from `api.GetBackground()`.
func (api *API) SetBackground(background MessageBackground) (err error) {
	// OPTIONS https://chatango.com/updatemsgbg
	// POST https://chatango.com/updatemsgbg
	var res *http.Response

	data := background.GetForm()
	data.Set("lo", api.Username)
	data.Set("p", api.Password)

	if res, err = api.client.PostForm("https://chatango.com/updatemsgbg", data); err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return ErrRequestFailed
	}

	return
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

	return
}

// GetStyle retrieves the message style of the specified username.
func (api *API) GetStyle(username string) (style MessageStyle, err error) {
	var res *http.Response

	path0 := username[0:1]
	path1 := username[0:1]
	if len(username) > 1 {
		path1 = username[1:2]
	}
	url := fmt.Sprintf("https://ust.chatango.com/profileimg/%s/%s/%s/msgstyles.json", path0, path1, username)

	if res, err = api.client.Get(url); err != nil {
		return
	}
	defer res.Body.Close()

	var body []byte
	if body, err = io.ReadAll(res.Body); err != nil {
		return
	}

	err = json.Unmarshal(body, &style)

	return
}

// SetStyle updates the message style of the current user.
// Normally, the `style` parameter is obtained/modified from `api.GetStyle()`.
func (api *API) SetStyle(style MessageStyle) (err error) {
	// OPTIONS https://chatango.com/updatemsgstyles
	// POST https://chatango.com/updatemsgstyles
	var res *http.Response

	data := style.GetForm()
	data.Set("lo", api.Username)
	data.Set("p", api.Password)

	/* var bg MessageBackground
	if bg, err = api.GetBackground(api.Username); err != nil {
		data.Set("hasrec", fmt.Sprintf("%d", bg.HasRecording))
	} */

	if res, err = api.client.PostForm("https://chatango.com/updatemsgstyles", data); err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return ErrRequestFailed
	}

	return
}

// GetToken retrieves the token for the specified GCM ID.
func (api *API) GetToken(gcmID string) (token string, err error) {
	var res *http.Response

	data := url.Values{
		"sid":       {api.Username},
		"pwd":       {api.Password},
		"encrypted": {"false"},
		"gcm":       {gcmID},
		"version":   {"50"},
		"os":        {"oreo"},
		"serial":    {"UNKNOWN"},
		"model":     {"Samsung S8"},
	}

	if res, err = api.client.PostForm("https://chatango.com/settokenapp", data); err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		var jsonResponse struct {
			Data struct {
				Token string `json:"token"`
			} `json:"data"`
			Type string `json:"type"`
		}

		if err = json.NewDecoder(res.Body).Decode(&jsonResponse); err != nil {
			return
		}

		if jsonResponse.Type == "success" {
			token = jsonResponse.Data.Token
			return
		}
	}

	err = ErrRequestFailed

	return
}

// RegisterGCM registers the specified GCM ID with the provided token.
func (api *API) RegisterGCM(gcmID, token string) (err error) {
	var res *http.Response

	data := url.Values{
		"token":   {token},
		"gcm":     {gcmID},
		"version": {"50"},
		"os":      {"oreo"},
		"serial":  {"UNKNOWN"},
		"model":   {"Samsung S8"},
	}

	if res, err = api.client.PostForm("https://chatango.com/updategcm", data); err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return ErrRequestFailed
	}

	return
}

// UnregisterGCM unregisters the specified GCM ID.
func (api *API) UnregisterGCM(gcmID string) (err error) {
	var res *http.Response

	data := url.Values{
		"sid": {api.Username},
		"gcm": {gcmID},
	}

	if res, err = http.PostForm("https://chatango.com/unregistergcm", data); err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return ErrRequestFailed
	}

	return
}

// CheckUsername checks whether the specified username is available for purchase.
// It returns `ok=true` if the username is purchasable.
// If not, the reasons for it not being available will be provided in the `notOkReasons` slice.
func (api *API) CheckUsername(username string) (ok bool, notOkReasons []string, err error) {
	var res *http.Response

	data := url.Values{
		"name": {username},
	}

	if res, err = api.client.PostForm("https://st.chatango.com/script/namecheckeraccsales", data); err != nil {
		return
	}
	defer res.Body.Close()

	var body []byte
	if body, err = io.ReadAll(res.Body); err != nil {
		return
	}

	if strings.Contains(string(body), "error") {
		notOkReasons = append(notOkReasons, "invalid username")
		return
	}

	var values url.Values
	if values, err = url.ParseQuery(string(body)); err != nil {
		return
	}

	var answer int
	if answer, err = strconv.Atoi(values.Get("answer")); err != nil {
		return
	} else if answer == 0 {
		ok = true
		return
	}

	for k, v := range map[int]string{
		1:  "is not a current Chatango username",
		2:  "is a group",
		4:  "is an inappropriate word",
		8:  "contains inappropriate parts",
		16: "is not expired",
		32: "is currently being purchased",
		64: "belongs to an active group owner",
	} {
		if answer&k != 0 {
			notOkReasons = append(notOkReasons, v)
		}
	}

	return
}
