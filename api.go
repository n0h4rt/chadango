package chadango

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
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
	data := url.Values{
		"name":      {groupname},
		"makegroup": {"1"},
	}
	var res *http.Response
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
func (pq *PeopleQuery) GetForm() (form url.Values) {
	pq.AgeFrom = Min(99, Max(0, pq.AgeFrom))
	pq.AgeTo = Min(99, Max(0, pq.AgeTo))
	switch pq.Gender {
	case "B", "M", "F", "N":
	default:
		pq.Gender = "B"
	}
	pq.Radius = Min(9999, Max(0, pq.Radius))

	form.Set("ami", strconv.Itoa(pq.AgeFrom))
	form.Set("ama", strconv.Itoa(pq.AgeTo))
	form.Set("s", pq.Gender)
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
	return
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

// MyChatGroups represents the created & recent chat groups of the user.
// It also provides the number of unread private message.
type MyChatGroups struct {
	RecentGroups   [][2]string `json:"recent_groups"`
	UnreadMessages int         `json:"n_msg"`
	Groups         [][2]string `json:"groups"`
}

// GetRecentGroups returns the map of recent chat groups.
func (mg MyChatGroups) GetRecentGroups() map[string]string {
	recent := make(map[string]string)
	for _, arr := range mg.RecentGroups {
		recent[arr[0]], _ = url.QueryUnescape(arr[1])
	}
	return recent
}

// GetGroups returns the map of all chat groups.
func (mg MyChatGroups) GetGroups() map[string]string {
	groups := make(map[string]string)
	for _, arr := range mg.Groups {
		groups[arr[0]], _ = url.QueryUnescape(arr[1])
	}
	return groups
}

// GetGroupList retrieves the user's chat group list.
func (api *API) GetGroupList() (groups MyChatGroups, err error) {
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
	XMLName  xml.Name     `xml:"mod"`
	Body     QueryEscaped `xml:"body"`
	Gender   string       `xml:"s"`
	Birth    BirthDate    `xml:"b"`
	Location Location     `xml:"l"`
	Premium  PremiumDate  `xml:"d"`
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
	Country   string  `xml:"c,attr"`
	G         string  `xml:"g,attr"`
	Latitude  float64 `xml:"lat,attr"`
	Longitude float64 `xml:"lon,attr"`
	Text      string  `xml:",chardata"`
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
	XMLName xml.Name     `xml:"mod"`
	Body    QueryEscaped `xml:"body"`
	T       string       `xml:"t"`
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

// BackgroundInfo represents the background information of a user.
type BackgroundInfo struct {
	Align        string `xml:"align,attr"`
	Alpha        int    `xml:"bgalp,attr"`
	Color        string `xml:"bgc,attr"`
	HasRecording int64  `xml:"hasrec,attr"`
	ImageAlpha   int    `xml:"ialp,attr"`
	IsVid        bool   `xml:"isvid,attr"`
	Tile         bool   `xml:"tile,attr"`
	UseImage     bool   `xml:"useimg,attr"`
}

// GetBackgroundInfo retrieves the background information of the specified username.
func (api *API) GetBackgroundInfo(username string) (info BackgroundInfo, err error) {
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
	err = xml.Unmarshal(body, &info)
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

	err = ErrSetTokenFailed
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
		return ErrGCMRegFailed
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
		return ErrGCMUnregFailed
	}

	return
}
