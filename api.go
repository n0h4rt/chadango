package chadango

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"

	"github.com/n0h4rt/chadango/models"
	"github.com/n0h4rt/chadango/utils"
)

// httpClient is an [http.Client] to interact with the Chatango APIs.
var httpClient *http.Client

// initHttpClient initializes the [http.Client] with custom headers and a cookie jar.
func initHttpClient() {
	httpClient = &http.Client{
		Transport: &Transport{
			Transport: http.DefaultTransport,
			Headers: map[string]string{
				"Host":       "script.st.chatango.com",
				"Origin":     "https://st.chatango.com",
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			},
		},
		Timeout: API_TIMEOUT,
	}

	var err error
	if httpClient.Jar, err = cookiejar.New(nil); err != nil {
		log.Fatalf("Failed to create cookie jar: %v", err)
	}
}

// Transport is a custom RoundTripper implementation.
// It adds custom headers to the request before performing the request using the underlying Transport.
type Transport struct {
	Transport http.RoundTripper // Underlying RoundTripper.
	Headers   map[string]string // Custom headers to be added to the requests.
}

// RoundTrip executes a single HTTP request and returns its response.
//
// Args:
//   - req: The HTTP request to be executed.
//
// Returns:
//   - *http.Response: The HTTP response.
//   - error: An error if the request fails.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Add custom headers to the request.
	for key, value := range t.Headers {
		req.Header.Set(key, value)
	}

	// Perform the request using the underlying Transport.
	return t.Transport.RoundTrip(req)
}

// APIClient represents a client for the Chatango API.
type APIClient struct {
	context context.Context
}

// executeRequest executes the given HTTP request and checks for errors.
//
// Args:
//   - req: The HTTP request to be executed.
//
// Returns:
//   - *http.Response: The HTTP response.
//   - error: An error if the request fails.
func executeRequest(req *http.Request) (res *http.Response, err error) {
	if res, err = httpClient.Do(req); err != nil {
		return
	}
	if res.StatusCode != http.StatusOK {
		defer res.Body.Close()
		err = ErrRequestFailed
	}
	return
}

// Get sends a GET request to the specified URL with the provided parameters.
//
// Args:
//   - url: The URL to send the GET request to.
//   - param: The URL parameters to include in the request.
//
// Returns:
//   - *http.Response: The HTTP response.
//   - error: An error if the request fails.
func (p *APIClient) Get(url string, param url.Values) (*http.Response, error) {
	req, err := http.NewRequestWithContext(p.context, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = param.Encode()

	return executeRequest(req)
}

// PostForm sends a POST request with form data to the specified URL.
//
// Args:
//   - url: The URL to send the POST request to.
//   - data: The form data to include in the POST request.
//
// Returns:
//   - *http.Response: The HTTP response.
//   - error: An error if the request fails.
func (p *APIClient) PostForm(url string, data url.Values) (*http.Response, error) {
	req, err := http.NewRequestWithContext(p.context, "POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	return executeRequest(req)
}

// PostMultipart sends a POST request with body and its content-type to the specified URL.
//
// Args:
//   - url: The URL to send the POST request to.
//   - body: The body of the POST request.
//   - ctype: The content type of the POST request.
//
// Returns:
//   - *http.Response: The HTTP response.
//   - error: An error if the request fails.
func (p *APIClient) PostMultipart(url string, body io.Reader, ctype string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(p.context, "POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", ctype)

	return executeRequest(req)
}

// PrivateAPI represents a compilation of various Chatango APIs that needs to be authenticated.
type PrivateAPI struct {
	APIClient

	MessageBackground models.MessageBackground
	MessageStyle      models.MessageStyle

	username       string
	password       string
	cookies        map[string]string
	loggedIn       bool
	recentGroups   map[string]string // Map of recently visited groups.[key=name,val=desc]
	unreadMsgCount int               // Number of unread private messages.
	createdGroups  map[string]string // Map of created groups.[key=name,val=desc]
	gcmID          string
	gcmToken       string
}

// NewPrivateAPI creates a new [chadango.PrivateAPI] instance.
//
// Args:
//   - username: The username of the Chatango account.
//   - password: The password of the Chatango account.
//   - ctx: The context for the API client.
//
// Returns:
//   - *PrivateAPI: A new [PrivateAPI] instance.
func NewPrivateAPI(username, password string, ctx context.Context) *PrivateAPI {
	api := &PrivateAPI{
		username:      username,
		password:      password,
		cookies:       make(map[string]string),
		recentGroups:  make(map[string]string),
		createdGroups: make(map[string]string),
	}
	api.context = ctx

	return api
}

// GetCookie retrieves a cookie value.
//
// Args:
//   - name: The name of the cookie.
//
// Returns:
//   - string: The value of the cookie.
//   - bool: True if the cookie exists, false otherwise.
func (p *PrivateAPI) GetCookie(name string) (string, bool) {
	value, ok := p.cookies[name]
	return value, ok
}

// GetCreatedGroups returns the map of created chat groups.
//
// Returns:
//   - map[string]string: A map of created groups, where the key is the group name and the value is the group description.
func (p *PrivateAPI) GetCreatedGroups() map[string]string {
	return p.createdGroups
}

// GetGCMID returns the GCM ID of the user.
//
// Returns:
//   - string: The GCM ID of the user.
func (p *PrivateAPI) GetGCMID() string {
	return p.gcmID
}

// GetGCMToken returns the GCM token of the user.
//
// Returns:
//   - string: The GCM token of the user.
func (p *PrivateAPI) GetGCMToken() string {
	return p.gcmToken
}

// GetMsgBgImage returns the URL of the message background image.
//
// Returns:
//   - string: The URL of the message background image.
func (p *PrivateAPI) GetMsgBgImage() string {
	return p.MessageBackground.GetImageURL()
}

// GetRecentGroups returns the map of recent chat groups.
//
// Returns:
//   - map[string]string: A map of recent groups, where the key is the group name and the value is the group description.
func (p *PrivateAPI) GetRecentGroups() map[string]string {
	return p.recentGroups
}

// GetUnreadMsgCount returns the number of unread private messages.
//
// Returns:
//   - int: The number of unread private messages.
func (p *PrivateAPI) GetUnreadMsgCount() int {
	return p.unreadMsgCount
}

// IsLoggedIn returns true if the user is logged in, false otherwise.
//
// Returns:
//   - bool: True if the user is logged in, false otherwise.
func (p *PrivateAPI) IsLoggedIn() bool {
	return p.loggedIn
}

// Login authenticates the client by obtaining and storing the necessary cookies.
//
// Returns:
//   - error: An error if the login fails.
func (p *PrivateAPI) Login() error {
	data := url.Values{
		"user_id":     {p.username},
		"password":    {p.password},
		"storecookie": {"on"},
		"checkerrors": {"yes"},
	}

	res, err := p.PostForm(API_LOGIN, data)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	p.loggedIn = true

	for _, cookie := range httpClient.Jar.Cookies(res.Request.URL) {
		p.cookies[cookie.Name] = cookie.Value
	}

	return nil
}

// RegisterGCM registers the GCM ID with the provided token.
//
// Returns:
//   - error: An error if the registration fails.
func (p *PrivateAPI) RegisterGCM() (err error) {
	data := url.Values{
		"token":   {p.gcmToken},
		"gcm":     {p.gcmID},
		"version": {"50"},
		"os":      {"oreo"},
		"serial":  {"UNKNOWN"},
		"model":   {"Samsung S8"},
	}

	var res *http.Response
	res, err = p.PostForm(API_REG_GCM, data)
	if err != nil {
		return
	}
	defer res.Body.Close()

	return
}

// RetrieveMsgBg retrieves the message background of the current user.
//
// Returns:
//   - error: An error if the retrieval fails.
func (p *PrivateAPI) RetrieveMsgBg() error {
	msgBg, err := publicAPI.GetBackground(p.username)
	if err != nil {
		return err
	}

	p.MessageBackground = msgBg

	return nil
}

// RetrieveMsgStyle retrieves the message style of the current user.
//
// Returns:
//   - error: An error if the retrieval fails.
func (p *PrivateAPI) RetrieveMsgStyle() error {
	msgStyle, err := publicAPI.GetStyle(p.username)
	if err != nil {
		return err
	}

	p.MessageStyle = msgStyle

	return nil
}

// RetrieveTokenGCM retrieves the GCM token for the specified GCM ID.
//
// Returns:
//   - error: An error if the retrieval fails.
func (p *PrivateAPI) RetrieveTokenGCM() (err error) {
	data := url.Values{
		"sid":       {p.username},
		"pwd":       {p.password},
		"encrypted": {"false"},
		"gcm":       {p.gcmID},
		"version":   {"50"},
		"os":        {"oreo"},
		"serial":    {"UNKNOWN"},
		"model":     {"Samsung S8"},
	}

	var res *http.Response
	res, err = p.PostForm(API_SET_TOKEN_GCM, data)
	if err != nil {
		return
	}
	defer res.Body.Close()

	resp := struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
		Type string `json:"type"`
	}{}

	err = json.NewDecoder(res.Body).Decode(&resp)

	if resp.Type == "success" {
		// This token must be keep in a persistent file for the next usage.
		p.gcmToken = resp.Data.Token
		return
	}

	return
}

// RetrieveUpdate retrieves the latest updates for the user's groups and unread messages.
//
// Returns:
//   - error: An error if the retrieval fails.
func (p *PrivateAPI) RetrieveUpdate() (err error) {
	var res *http.Response
	if res, err = p.PostForm(API_GRP_LIST_UPD, nil); err != nil {
		return
	}
	defer res.Body.Close()

	updates := struct {
		RecentGroups   [][2]string `json:"recent_groups"`
		UnreadMsgCount int         `json:"n_msg"`
		CreatedGroups  [][2]string `json:"groups"`
	}{}

	if err = json.NewDecoder(res.Body).Decode(&updates); err != nil {
		return
	}

	for _, g := range updates.RecentGroups {
		p.recentGroups[g[0]] = g[1]
	}

	p.unreadMsgCount = updates.UnreadMsgCount

	for _, g := range updates.CreatedGroups {
		p.createdGroups[g[0]] = g[1]
	}

	return
}

// SearchPeople searches for people based on the provided query.
//
// Args:
//   - query: The search query.
//
// Returns:
//   - []PeopleResult: A list of search results.
//   - error: An error if the search fails.
func (p *PrivateAPI) SearchPeople(query models.PeopleQuery) (result []models.PeopleResult, err error) {
	var res *http.Response
	if res, err = p.PostForm(API_SEARCH_PEOPLE, query.GetForm()); err != nil {
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
		for _, entry := range strings.Split(data, ",") {
			username, isonline, _ = strings.Cut(entry, ";")
			result = append(result, models.PeopleResult{Username: username, IsOnline: isonline == "1"})
		}
	}

	return
}

// SetGCMID sets the GCM ID of the user.
//
// Args:
//   - gcmID: The GCM ID of the user.
func (p *PrivateAPI) SetGCMID(gcmID string) {
	p.gcmID = gcmID
}

// SetGCMToken sets the GCM token of the user.
//
// Args:
//   - token: The GCM token of the user.
func (p *PrivateAPI) SetGCMToken(token string) {
	p.gcmToken = token
}

// UnregisterGCM unregisters the specified GCM ID.
//
// Returns:
//   - error: An error if the unregistration fails.
func (p *PrivateAPI) UnregisterGCM() (err error) {
	data := url.Values{
		"sid": {p.username},
		"gcm": {p.gcmID},
	}

	var res *http.Response
	res, err = p.PostForm(API_UNREG_GCM, data)
	if err != nil {
		return
	}
	defer res.Body.Close()

	return
}

// UpdateMsgBg updates the message background of the current user.
//
// Returns:
//   - error: An error if the update fails.
func (p *PrivateAPI) UpdateMsgBg() (err error) {
	bg := p.MessageBackground.GetForm()
	bg.Set("lo", p.username)
	bg.Set("p", p.password)

	var res *http.Response
	res, err = p.PostForm(API_UPD_MSG_BG, bg)
	if err != nil {
		return
	}
	defer res.Body.Close()

	return
}

// UpdateMsgStyle updates the message style of the current user.
//
// Returns:
//   - error: An error if the update fails.
func (p *PrivateAPI) UpdateMsgStyle() (err error) {
	style := p.MessageStyle.GetForm()
	style.Set("lo", p.username)
	style.Set("p", p.password)

	var res *http.Response
	res, err = p.PostForm(API_UPD_MSG_STYLE, style)
	if err != nil {
		return
	}
	defer res.Body.Close()

	return
}

// UploadImage uploads an image to the Chatango server.
//
// Args:
//   - filename: The name of the image file.
//   - image: The image data.
//
// Returns:
//   - UploadedImage: The uploaded image information.
//   - error: An error if the upload fails.
func (p *PrivateAPI) UploadImage(filename string, image io.Reader) (img models.UploadedImage, err error) {
	var (
		reqBody *bytes.Buffer
		writer  = multipart.NewWriter(reqBody)
		part    io.Writer
	)

	if err = writer.WriteField("u", p.username); err != nil {
		return
	}

	if err = writer.WriteField("p", p.password); err != nil {
		return
	}

	if part, err = writer.CreateFormFile("filedata", filename); err != nil {
		return
	}
	if _, err = io.Copy(part, image); err != nil {
		return
	}

	var res *http.Response
	res, err = p.PostMultipart(API_UPLOAD_IMG, reqBody, writer.FormDataContentType())
	if err != nil {
		return
	}
	defer res.Body.Close()

	var body []byte
	if body, err = io.ReadAll(res.Body); err != nil {
		return
	}

	status, id, ok := strings.Cut(string(body), ":")
	if !ok || !strings.EqualFold(status, "success") {
		err = ErrRequestFailed
		return
	}

	img.ID, _ = strconv.Atoi(id)
	img.Username = strings.ToLower(p.username)

	return
}

// UploadMsgBgImage uploads an image to be used as the message background.
//
// Args:
//   - filename: The name of the image file.
//   - image: The image data.
//
// Returns:
//   - error: An error if the upload fails.
func (p *PrivateAPI) UploadMsgBgImage(filename string, image io.Reader) (err error) {
	var (
		reqBody *bytes.Buffer
		writer  = multipart.NewWriter(reqBody)
	)

	if err = writer.WriteField("lo", p.username); err != nil {
		return
	}

	if err = writer.WriteField("p", p.password); err != nil {
		return
	}

	var part io.Writer
	if part, err = writer.CreateFormFile("Filedata", filename); err != nil {
		return
	}
	if _, err = io.Copy(part, image); err != nil {
		return
	}

	var res *http.Response
	res, err = p.PostMultipart(API_UPD_MSG_BG, reqBody, writer.FormDataContentType())
	if err != nil {
		return
	}
	defer res.Body.Close()

	return
}

// PublicAPI represents a compilation of various Chatango APIs that doesn't needs to be authenticated.
type PublicAPI struct {
	APIClient
}

// NewPublicAPI creates a new PublicAPI instance.
//
// Args:
//   - ctx: The context for the API client.
//
// Returns:
//   - *PublicAPI: A new instance of PublicAPI.
func NewPublicAPI(ctx context.Context) *PublicAPI {
	api := &PublicAPI{}
	api.context = ctx

	return api
}

// CheckUsername checks whether the specified username is available for purchase.
//
// Args:
//   - username: The username to check for availability.
//
// Returns:
//   - bool: `true` if the username is purchasable, `false` otherwise.
//   - []string: A list of reasons why the username is not available.
//   - error: An error if the check fails.
func (p *PublicAPI) CheckUsername(username string) (ok bool, notOkReasons []string, err error) {
	data := url.Values{
		"name": {username},
	}

	var res *http.Response
	if res, err = p.PostForm(API_CHECK_USER, data); err != nil {
		return
	}
	defer res.Body.Close()

	var body []byte
	if body, err = io.ReadAll(res.Body); err != nil {
		return
	}
	strBody := string(body)

	if strings.Contains(strBody, "error") {
		notOkReasons = append(notOkReasons, "invalid username")
		return
	}

	var values url.Values
	if values, err = url.ParseQuery(strBody); err != nil {
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

// GetBackground retrieves the message background of the specified username.
//
// Args:
//   - username: The username to retrieve the background for.
//
// Returns:
//   - MessageBackground: The message background of the specified username.
//   - error: An error if the retrieval fails.
func (p *PublicAPI) GetBackground(username string) (background models.MessageBackground, err error) {
	username = strings.ToLower(username)

	var res *http.Response
	res, err = p.Get(utils.UsernameToURL(API_MSG_BG_XML, username), nil)
	if err != nil {
		return
	}
	defer res.Body.Close()

	if err = xml.NewDecoder(res.Body).Decode(&background); err != nil {
		return
	}

	background.Username = username

	return
}

// GetFullProfile retrieves the full profile of the specified username.
//
// Args:
//   - username: The username to retrieve the full profile for.
//
// Returns:
//   - FullProfile: The full profile of the specified username.
//   - error: An error if the retrieval fails.
func (p *PublicAPI) GetFullProfile(username string) (profile models.FullProfile, err error) {
	username = strings.ToLower(username)

	var res *http.Response
	if res, err = p.Get(utils.UsernameToURL(API_MINI_XML, username), nil); err != nil {
		return
	}
	defer res.Body.Close()

	err = xml.NewDecoder(res.Body).Decode(&profile)

	return
}

// GetMiniProfile retrieves the mini profile of the specified username.
//
// Args:
//   - username: The username to retrieve the mini profile for.
//
// Returns:
//   - MiniProfile: The mini profile of the specified username.
//   - error: An error if the retrieval fails.
func (p *PublicAPI) GetMiniProfile(username string) (profile models.MiniProfile, err error) {
	username = strings.ToLower(username)

	var res *http.Response
	if res, err = p.Get(utils.UsernameToURL(API_MINI_XML, username), nil); err != nil {
		return
	}
	defer res.Body.Close()

	if err = xml.NewDecoder(res.Body).Decode(&profile); err != nil {
		return
	}

	profile.Username = username

	return
}

// GetStyle retrieves the message style of the specified username.
//
// Args:
//   - username: The username to retrieve the message style for.
//
// Returns:
//   - MessageStyle: The message style of the specified username.
//   - error: An error if the retrieval fails.
func (p *PublicAPI) GetStyle(username string) (style models.MessageStyle, err error) {
	username = strings.ToLower(username)

	var res *http.Response
	res, err = p.Get(utils.UsernameToURL(API_MSG_STYLE_JSON, username), nil)
	if err != nil {
		return
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&style)

	return
}

// IsGroup checks whether the specified name is a group.
//
// Args:
//   - groupname: The name to check for being a group.
//
// Returns:
//   - bool: `true` if the name is a group, `false` otherwise.
//   - error: An error if the check fails.
func (p *PublicAPI) IsGroup(groupname string) (answer bool, err error) {
	data := url.Values{
		"name":      {groupname},
		"makegroup": {"1"},
	}

	var res *http.Response
	if res, err = p.PostForm(API_CHECK_GROUP, data); err != nil {
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

var (
	privateAPI *PrivateAPI
	publicAPI  *PublicAPI
)

// initAPI initializes the API clients with the provided username, password, and context.
//
// Args:
//   - username: The username for the private API.
//   - password: The password for the private API.
//   - ctx: The context for the API clients.
func initAPI(username, password string, ctx context.Context) {
	privateAPI = NewPrivateAPI(username, password, ctx)
	publicAPI = NewPublicAPI(ctx)
}
