package chadango

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/n0h4rt/chadango/models"
	"github.com/n0h4rt/chadango/utils"
	"github.com/rs/zerolog/log"
)

// Group represents a chat group with various properties and state.
//
// It provides methods for connecting, disconnecting, sending messages, retrieving user status, and managing settings.
// The [Group] struct also handles events related to private messages, friend status, and user profile updates.
type Group struct {
	App       *Application // Reference to the application.
	Name      string       // The name of the group.
	AnonName  string       // The anonymous user name format.
	NameColor string       // The color for displaying name in the message.
	TextColor string       // The color for displaying text in the message.
	TextFont  string       // The font style for displaying text in the message.
	TextSize  int          // The font size for displaying text in the message.

	WsUrl     string               // The WebSocket URL for connecting to the group.
	ws        *WebSocket           // The WebSocket connection to the group.
	Connected bool                 // Indicates if the group is currently connected.
	events    chan string          // Channel for propagating events back to the listener.
	takeOver  chan context.Context // Channel for taking over the WebSocket connection.
	backoff   *Backoff             // Cancelable backoff for reconnection.
	context   context.Context      // Context for running the group operations.
	cancelCtx context.CancelFunc   // Function for stopping group operations.

	Version    [2]int                 // The version of the group.
	Owner      string                 // The owner of the group.
	SessionID  string                 // The session ID for the group.
	UserID     int                    // The user ID for the group.
	LoggedIn   bool                   // Indicates if the user is logged in to the group.
	LoginName  string                 // The login name of the user.
	LoginTime  time.Time              // The time when the user logged in.
	TimeDiff   time.Duration          // The time difference between the server and client.
	LoginIp    string                 // The IP address of the user.
	Moderators SyncMap[string, int64] // Map of moderators and their access levels.
	Flag       int64                  // The flag value for the group.

	Channel          int64         // The channel flag of the group.
	Restrict         time.Time     // The time when the group is restricted from the flood ban and auto moderation.
	RateLimit        time.Duration // The rate limit duration for sending messages.
	RateLimited      time.Time     // The time when the group is rate-limited.
	MaxMessageLength int           // The maximum allowed length of a message.
	PremiumExpireAt  time.Time     // The time when the premium membership expires.

	Messages       OrderedSyncMap[string, *Message] // Ordered map of messages history in the group.
	TempMessages   SyncMap[string, *Message]        // Map of temporary messages in the group.
	TempMessageIds SyncMap[string, string]          // Map of temporary message IDs in the group.

	Participants     SyncMap[string, *models.Participant] // Map of participants in the group. Invoke [Group.GetParticipantsStart] to initiate the participant feeds.
	ParticipantCount int64                                // The total count of participants in the group.
	UserCount        int                                  // The count of registered users in the group.
	AnonCount        int                                  // The count of anonymous users in the group.
}

func (g *Group) initFields() {
	g.Moderators = NewSyncMap[string, int64]()
	g.Messages = NewOrderedSyncMap[string, *Message]()
	g.TempMessages = NewSyncMap[string, *Message]()
	g.TempMessageIds = NewSyncMap[string, string]()
	g.Participants = NewSyncMap[string, *models.Participant]()
}

// Connect establishes a connection to the server.
//
// Args:
//   - ctx: The context for the connection.
//
// Returns:
//   - error: An error if the connection cannot be established.
func (g *Group) Connect(ctx context.Context) (err error) {
	if g.Connected {
		return ErrAlreadyConnected
	}

	g.context, g.cancelCtx = context.WithCancel(ctx)

	log.Debug().Str("Name", g.Name).Msg("Connecting")

	defer func() {
		if err != nil {
			if g.ws != nil {
				g.ws.Close()
			}

			log.Debug().Str("Name", g.Name).Msg("Connect failed")
		}
	}()

	err = g.connect()
	if err != nil {
		g.cancelCtx()
		return
	}

	g.Connected = true

	log.Debug().Str("Name", g.Name).Msg("Connected")

	return
}

// connect establishes a WebSocket connection to the group chat server.
//
// It initializes necessary channels and attempts to log in to the group chat.
// If successful, it sets up the WebSocket to sustain the connection and starts
// listening for incoming events.
//
// Returns:
//   - error: An error if the connection cannot be established.
func (g *Group) connect() (err error) {
	g.ws = &WebSocket{
		OnError: g.wsOnError,
	}
	if err = g.ws.Connect(g.WsUrl); err != nil {
		return
	}

	// Initializing channels.
	g.events = make(chan string, EVENT_BUFFER_SIZE)
	g.takeOver = make(chan context.Context)

	var frame string

	// This may not be necessary, but oh well, let's just do it anyway.
	if err = g.Send("v", "\x00"); err != nil {
		return
	}
	if frame, err = g.ws.Recv(); err != nil {
		return
	}
	if !strings.HasPrefix(frame, "v") {
		return ErrLoginFailed
	}
	g.events <- frame

	// Attempting to login to the group chat.
	if g.LoggedIn {
		err = g.Send("bauth", g.Name, g.App.Config.SessionID, g.App.Config.Username, g.App.Config.Password, "\x00")
	} else {
		err = g.Send("bauth", g.Name, g.App.Config.SessionID, g.App.Config.Username, "", "\x00")
	}
	if err != nil {
		return
	}
	if frame, err = g.ws.Recv(); err != nil {
		return
	}
	if head, _, ok := strings.Cut(frame, ":"); ok && head != "ok" {
		return ErrLoginFailed
	}
	g.events <- frame

	g.initFields()
	g.ws.Sustain(g.context)
	go g.listen()

	return
}

// listen listens for incoming messages and events on the WebSocket connection.
func (g *Group) listen() {
	var frame string
	var ok bool
	var release context.Context
	for {
		select {
		case <-g.context.Done():
			return
		case frame, ok = <-g.events:
			if !ok {
				return
			}
			go g.wsOnFrame(frame)
		case frame, ok = <-g.ws.Events:
			if !ok {
				return
			}
			go g.wsOnFrame(frame)
		case release, ok = <-g.takeOver:
			if !ok {
				return
			}
		inner:
			for {
				select {
				case <-g.context.Done():
					return
				case <-release.Done():
					break inner
				case frame, ok = <-g.events:
					if !ok {
						return
					}
					go g.wsOnFrame(frame)
				}
			}
		}
	}
}

// Disconnect gracefully closes the connection to the server.
func (g *Group) Disconnect() {
	if g.backoff != nil {
		g.backoff.Cancel()
	}

	if !g.Connected {
		return
	}

	g.Connected = false
	g.cancelCtx()
	g.ws.Close()
}

// Reconnect reconnects the group to the server.
//
// Returns:
//   - error: An error if the reconnection fails.
func (g *Group) Reconnect() (err error) {
	g.ws.Close()

	g.backoff = &Backoff{
		Duration:    BASE_BACKOFF_DUR,
		MaxDuration: MAX_BACKOFF_DUR,
	}
	defer func() {
		g.backoff = nil
	}()
	for retries := 0; retries < MAX_RETRIES && !g.backoff.Sleep(g.context); retries++ {
		if err = g.connect(); err == nil {
			return
		}
	}

	// Either canceled or reached the maximum retries.
	return ErrRetryEnds
}

// Send joins the [args] with a ":" separator and then sends it to the server asynchronously.
//
// Note:
//   - A terminator should be included in the last [args].
//   - The terminator can be "\r\n" or "\x00" depending on the command.
//
// Args:
//   - args: The arguments to send to the server.
//
// Returns:
//   - error: An error if sending the message fails.
func (g *Group) Send(args ...string) error {
	if !g.ws.Connected {
		return ErrNotConnected
	}

	length := len(args)
	if length == 0 {
		return ErrNoArgument
	}

	// The terminator should be appended without a separator.
	// Valid terminator: \r\n, \x00
	terminator := args[length-1]
	args = args[:length-1]
	command := strings.Join(args, ":")

	return g.ws.Send(command + terminator)
}

// SyncSendWithTimeout sends the specified arguments and waits for a response or timeout.
//
// First, a [Group.takeOver] request will be made and it will wait until the [listener] goroutine catches it.
// Then, the [args] will be sent to the server.
// Each time a frame is received, the [callback] function is invoked and passed the frame.
// The callback should return [false] if a correct frame is acquired, and [true] otherwise.
//
// Args:
//   - callback: The function to handle received frames.
//   - timeout: The duration to wait for a response.
//   - args: The arguments to send.
//
// Returns:
//   - error: An error if sending the arguments or receiving a response fails.
func (g *Group) SyncSendWithTimeout(callback func(string) bool, timeout time.Duration, args ...string) (err error) {
	ctx, cancel := context.WithTimeout(g.context, timeout)
	defer cancel()

	// Make a takeover request to allow the listener to relinquish the connection.
	select {
	case <-ctx.Done():
		return ErrTimeout
	case g.takeOver <- ctx:
	}

	if err = g.Send(args...); err != nil {
		return
	}

	var frame string
	var ok bool
	for {
		select {
		case <-ctx.Done():
			return ErrTimeout
		case frame, ok = <-g.ws.Events:
			if !ok {
				close(g.events)
				return ErrConnectionClosed
			}
			if strings.HasPrefix(frame, "climited") {
				// This response is received when messages are being sent too quickly.
				// I believe this applies globally to any type of command sent to the server,
				// not limited to sending messages alone.
				// If it turns out to be correct, move this `climited` check to the `SyncSend` method.
				// climited:1485666967794:bm:e8n2:0:<n000/><f x9000="1">n
				return ErrCLimited
			}
			if !callback(frame) {
				return
			}
		}
	}
}

// SyncSend will send the [args] and wait until receiving the correct reply or until timeout (default to 5 seconds).
//
// For more information, refer to the documentation of [Group.SyncSendWithTimeout].
//
// Args:
//   - callback: The function to handle received frames.
//   - args: The arguments to send.
//
// Returns:
//   - error: An error if sending the arguments or receiving a response fails.
func (g *Group) SyncSend(callback func(string) bool, args ...string) error {
	return g.SyncSendWithTimeout(callback, SYNC_SEND_TIMEOUT, args...)
}

// SendMessage sends a message to the group with the specified text and optional arguments.
//
// It returns a pointer to the sent [*Message] and an error (if any occurred).
// The text can include formatting placeholders (%s, %d, etc.), and optional arguments can be provided to fill in these placeholders.
// The function also handles flood warnings, restricted messages, spam warnings, rate limiting, and other Chatango-specific events.
// If the group is logged in, the message text is styled with the group's name color, text size, text color, and text font.
// If the group is anonymous, the message text is modified with the anonymous seed based on the group's [Group.AnonName] and [Group.UserID].
// The function replaces newlines with the "<br/>" HTML tag to format the message properly.
//
// Args:
//   - text: The message text.
//   - a: Optional arguments to format the message text.
//
// Returns:
//   - *Message: The sent message.
//   - error: An error if sending the message fails.
func (g *Group) SendMessage(text string, a ...any) (msg *Message, err error) {
	var idBuffer = map[string]string{}
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "b":
			g.events <- frame
			if msg == nil {
				message := ParseGroupMessage(data, g)
				if !message.User.IsSelf {
					return true
				}
				msg = message
				if newID, ok := idBuffer[message.ID]; ok {
					message.ID = newID
					return false
				}
			}
		case "u":
			g.events <- frame
			oldID, newID, _ := strings.Cut(data, ":")
			if msg != nil && msg.ID == oldID {
				msg.ID = newID
				return false
			}
			idBuffer[oldID] = newID
		case "show_fw":
			g.eventRestrictUpdate(data)
			err = ErrFloodWarning
			return false
		case "show_tb", "tb":
			g.eventRestrictUpdate(data)
			err = ErrRestricted
			return false
		case "show_nlp":
			if mask, _ := strconv.Atoi(data); mask&2 == 2 {
				err = ErrSpamWarning
			} else if mask&8 == 8 {
				err = ErrShortWarning
			} else {
				err = ErrNonSenseWarning
			}
			return false
		case "show_nlp_tb":
			// The first data in the fields is unknown, so let's leave it as it is for now.
			// show_nlp_tb:3:900
			_, min, _ := strings.Cut(data, ":")
			g.eventRestrictUpdate(min)
			err = ErrRestricted
			return false
		case "nlptb":
			g.eventRestrictUpdate(data)
			err = ErrRestricted
			return false
		case "msglexceeded":
			g.MaxMessageLength, _ = strconv.Atoi(data)
			err = ErrMessageLength
			return false
		case "ratelimited":
			dur, _ := time.ParseDuration(data + "s")
			g.RateLimited = time.Now().Add(dur)
			err = ErrRateLimited
			return false
		case "mustlogin":
			err = ErrMustLogin
			return false
		case "proxybanned":
			err = ErrProxyBanned
			return false
		case "verificationrequired":
			err = ErrVerificationRequired
			return false
		default:
			// Send the frame back to the listener.
			g.events <- frame
		}
		return true
	}

	// I'm not sure what it is for, but it gets sent back to the client when "climited" occurs.
	randomString := strconv.FormatInt(int64(15e5*rand.Float64()), 36)

	text = fmt.Sprintf(text, a...)

	// Style thing
	if g.LoggedIn {
		text = fmt.Sprintf(`<n%s/><f x%02d%s="%s">%s`, g.NameColor, g.TextSize, g.TextColor, g.TextFont, text)
	} else {
		// It would look nicer if it were wrapped in a separate method.
		if g.AnonName == "" {
			g.AnonName = "anon0001"
		}
		// Same as above, the anonymous seed should not be recalculated for each message sending.
		text = fmt.Sprintf(`<n%d/>%s`, utils.CreateAnonSeed(g.AnonName, g.UserID), text)
	}

	// Replacing newlines with the `<br/>` tag.
	text = strings.ReplaceAll(text, "\r\n", "<br/>")
	text = strings.ReplaceAll(text, "\n", "<br/>")

	if err2 := g.SyncSend(cb, "bm", randomString, fmt.Sprintf("%d", g.Channel), text, "\r\n"); err == nil && err2 != nil {
		err = err2
	}

	return
}

// SendMessageChunked sends the chunked [text] with a size of [chunkSize] and returns the sent [[]*Message].
//
// In the event of an error, the already sent messages will be returned along with the error for the unsent message.
//
// Args:
//   - text: The large message text.
//   - chunkSize: The size of each chunk.
//
// Returns:
//   - []*Message: The sent messages.
//   - error: An error if sending the message in chunks fails.
func (g *Group) SendMessageChunked(text string, chunkSize int) (msgs []*Message, err error) {
	var msg *Message
	for _, chunk := range utils.SplitTextIntoChunks(text, chunkSize) {
		if msg, err = g.SendMessage(chunk); err != nil {
			return
		}
		msgs = append(msgs, msg)
	}

	return
}

// IsRestricted checks if the group is restricted.
//
// The restriction can originate from either a flood ban or a rate limit.
//
// Returns:
//   - bool: True if the group is restricted.
func (g *Group) IsRestricted() bool {
	return g.Restrict.After(time.Now()) || g.RateLimited.After(time.Now())
}

// GetParticipantsStart initiates the "participant" event feeds and returns the current participants.
//
// The "participant" event will be triggered when there is user activity,
// such as joining, logging out, logging in, or leaving.
//
// Returns:
//   - *SyncMap[string, *Participant]: The current participants.
//   - error: An error if fetching the participants fails.
func (g *Group) GetParticipantsStart() (p *SyncMap[string, *models.Participant], err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "gparticipants":
			g.Participants.Clear()
			anoncount, entries, _ := strings.Cut(data, ":")
			g.AnonCount, _ = strconv.Atoi(anoncount)

			var fields []string
			var user *models.User
			var t time.Time
			var participant *models.Participant
			for _, entry := range strings.Split(entries, ";") {
				fields = strings.SplitN(entry, ":", 6)
				t, _ = utils.ParseTime(fields[1])
				userID, _ := strconv.Atoi(fields[2])
				if fields[3] != "None" {
					user = &models.User{Name: fields[3]}
				} else if fields[4] != "None" {
					user = &models.User{Name: fields[4], IsAnon: true}
				} else {
					user = &models.User{Name: utils.GetAnonName(int(t.Unix()), userID), IsAnon: true}
				}
				user.IsSelf = userID == g.UserID && user.Name == g.LoginName
				participant = &models.Participant{
					ParticipantID: fields[0],
					UserID:        userID,
					User:          user,
					Time:          t,
				}
				g.Participants.Set(fields[0], participant)
			}
			p = &g.Participants
			g.UserCount = g.Participants.Len()
			return false
		default:
			g.events <- frame
		}
		return true
	}

	err = g.SyncSend(cb, "gparticipants", "\r\n")

	return
}

// GetParticipantsStop stops the participant event feeds.
//
// This will leave [Group.Participants], [Group.UserCount], [Group.AnonCount] out of date.
//
// Returns:
//   - error: An error if stopping the fetch fails.
func (g *Group) GetParticipantsStop() error {
	return g.Send("gparticipants", "stop", "\r\n")
}

// GetRateLimit retrieves the rate limit settings for the group.
//
// Returns:
//   - time.Duration: The current rate limit.
//   - time.Duration: The current rate limit duration.
//   - error: An error if retrieving the rate limit fails.
func (g *Group) GetRateLimit() (rate, current time.Duration, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "getratelimit":
			r, c, _ := strings.Cut(data, ":")
			rate, _ = time.ParseDuration(r + "s")
			g.RateLimit = rate
			current, _ := time.ParseDuration(c + "s")
			g.RateLimited = time.Now().Add(current)
			return false
		default:
			g.events <- frame
		}
		return true
	}

	err = g.SyncSend(cb, "getratelimit", "\r\n")

	return
}

// SetRateLimit sets the rate limit interval for the group.
//
// Args:
//   - interval: The new rate limit interval.
//
// Returns:
//   - time.Duration: The set rate limit.
//   - error: An error if setting the rate limit fails.
func (g *Group) SetRateLimit(interval time.Duration) (rate time.Duration, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "ratelimitset":
			rate, _ = time.ParseDuration(data + "s")
			return false
		default:
			g.events <- frame
		}
		return true
	}

	err = g.SyncSend(cb, "setratelimit", fmt.Sprintf("%.0f", interval.Seconds()), "\r\n")

	return
}

// GetAnnouncement retrieves the announcement settings for the group.
//
// Returns:
//   - string: The announcement text.
//   - bool: True if announcements are enabled, otherwise false.
//   - time.Duration: The interval between announcements.
//   - error: An error if retrieving the announcement settings fails.
func (g *Group) GetAnnouncement() (annc string, enabled bool, interval time.Duration, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "getannc":
			fields := strings.SplitN(data, ":", 6)
			annc = fields[4]
			enabled = fields[0] != "0"
			interval, _ = time.ParseDuration(fields[3] + "s")
			return false
		default:
			g.events <- frame
		}
		return true
	}

	err = g.SyncSend(cb, "getannouncement", "\r\n")

	return
}

// SetAnnouncement sets the announcement settings for the group.
//
// Args:
//   - annc: The announcement text to set.
//   - enable: True to enable announcements, false to disable.
//   - interval: The interval between announcements.
//
// Returns:
//   - error: An error if setting the announcement fails.
func (g *Group) SetAnnouncement(annc string, enable bool, interval time.Duration) error {
	return g.Send("updateannouncement", utils.BoolZeroOrOne(enable), fmt.Sprintf("%.0f", interval.Seconds()), "\r\n")
}

// UpdateGroupFlag updates the group's flag by adding and removing specific flags.
//
// Args:
//   - addition: The flags to add.
//   - removal: The flags to remove.
//
// Returns:
//   - error: An error if updating the group's flag fails.
func (g *Group) UpdateGroupFlag(addition, removal int64) (err error) {
	cb := func(frame string) bool {
		head, _, _ := strings.Cut(frame, ":")
		switch head {
		case "groupflagstoggled":
			/* fields := strings.SplitN(data, ":", 3)
			addition, _ := strconv.ParseInt(fields[0], 10, 64)
			removal, _ := strconv.ParseInt(fields[1], 10, 64)
			status, _ := strconv.Atoi(fields[2])
			if fields[2] != "1" {
				err = ErrRequestFailed
			} */
			return false
		default:
			g.events <- frame
		}
		return true
	}

	err = g.SyncSend(cb, "updategroupflags", fmt.Sprintf("%d:%d", addition, removal), "\r\n")

	return
}

// GetPremiumInfo retrieves the premium status and expiration time for the group.
//
// This function would activate server validation for the premium status.
// For example, the message background and media won't activate before this command is sent.
//
// Returns:
//   - int: The premium flag.
//   - time.Time: The expiration time of the premium status.
//   - error: An error if retrieving the premium info fails.
func (g *Group) GetPremiumInfo() (flag int, expire time.Time, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "premium":
			fl, ti, _ := strings.Cut(data, ":")
			flag, _ = strconv.Atoi(fl)
			expire, _ = utils.ParseTime(ti)
			g.PremiumExpireAt = expire
			return false
		default:
			g.events <- frame
		}
		return true
	}

	err = g.SyncSend(cb, "getpremium", "\r\n")

	return
}

// SetBackground sets the background status of the group.
//
// Args:
//   - enable: True to enable the background feature, false to disable.
//
// Returns:
//   - error: An error if setting the background fails.
func (g *Group) SetBackground(enable bool) (err error) {
	if enable {
		if g.PremiumExpireAt.IsZero() {
			if _, g.PremiumExpireAt, err = g.GetPremiumInfo(); err != nil {
				return
			}
		}

		if g.PremiumExpireAt.Before(time.Now()) {
			return ErrRequestFailed
		}
	}

	return g.Send("msgbg", utils.BoolZeroOrOne(enable), "\r\n")
}

// SetMedia sets the media status of the group.
//
// Args:
//   - enable: True to enable the media feature, false to disable.
//
// Returns:
//   - error: An error if setting the media fails.
func (g *Group) SetMedia(enable bool) (err error) {
	if enable {
		if g.PremiumExpireAt.IsZero() {
			if _, g.PremiumExpireAt, err = g.GetPremiumInfo(); err != nil {
				return
			}
		}

		if g.PremiumExpireAt.Before(time.Now()) {
			return ErrRequestFailed
		}
	}

	return g.Send("msgmedia", utils.BoolZeroOrOne(enable), "\r\n")
}

// GetBanList retrieves a list of blocked users (ban list) for the group.
//
// The offset can be set to zero time to retrieve the newest result.
// The next offset can be defined from the last [Blocked.Time].
// The returned order is from newer to older.
//
// Args:
//   - offset: The offset time to start retrieving the ban list.
//   - amount: The number of banned users to retrieve.
//
// Returns:
//   - []Blocked: A list of blocked users.
//   - error: An error if retrieving the ban list fails.
func (g *Group) GetBanList(offset time.Time, ammount int) (banList []models.Blocked, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "blocklist":
			var fields []string
			var target string
			var t time.Time
			var banned models.Blocked
			for _, entry := range strings.Split(data, ";") {
				fields = strings.SplitN(entry, ":", 5)
				target = fields[2]
				if target == "" {
					target = "anon"
				}
				t, _ = utils.ParseTime(fields[3])
				banned = models.Blocked{
					IP:           fields[1],
					ModerationID: fields[0],
					Target:       target,
					Blocker:      fields[4],
					Time:         t,
				}
				banList = append(banList, banned)
			}
			return false
		default:
			g.events <- frame
		}
		return true
	}

	offsetS := fmt.Sprintf("%d", offset.Unix())
	ammountS := fmt.Sprintf("%d", ammount)

	err = g.SyncSend(cb, "blocklist", "block", offsetS, "next", ammountS, "anons", "1", "\r\n")

	return
}

// SearchBannedUser searches for a banned user in the group's ban list.
//
// The query can be either a user name or an IP address.
//
// Args:
//   - query: The user name or IP address to search for.
//
// Returns:
//   - Blocked: The banned user details.
//   - bool: True if the banned user is found, otherwise false.
//   - error: An error if searching for the banned user fails.
func (g *Group) SearchBannedUser(query string) (banned models.Blocked, ok bool, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "bansearchresult":
			fields := strings.SplitN(data, ":", 6)
			target := fields[1]
			if target == "" {
				target = "anon"
			}
			// It would be nicer to use a constant rather than a naked string.
			t, _ := time.Parse("2006-01-02 15:04:05", fields[5])
			banned = models.Blocked{
				IP:           fields[2],
				ModerationID: fields[3],
				Target:       target,
				Blocker:      fields[4],
				Time:         t,
			}
			ok = true
			return false
		case "badbansearchstring":
			// Simply return an empty result.
			return false
		default:
			g.events <- frame
		}
		return true
	}

	query = strings.TrimSpace(query)
	err = g.SyncSend(cb, "searchban", query, "\r\n")

	return
}

// BanUser bans the user associated with the specified message.
//
// Args:
//   - message: The message containing the user details to ban.
//
// Returns:
//   - error: An error if banning the user fails.
func (g *Group) BanUser(message *Message) (err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "blocked":
			g.events <- frame
			moderationID, _, _ := strings.Cut(data, ":")
			if moderationID == message.ModerationID {
				return false
			}
		default:
			g.events <- frame
		}
		return true
	}

	err = g.SyncSend(cb, "block", message.ModerationID, message.UserIP, message.User.Name, "\r\n")

	return
}

// GetUnbanList retrieves a list of unblocked users (unban list) for the group.
//
// The offset can be set to zero time to retrieve the newest result.
// The next offset can be defined from the last [Unblocked.Time].
// The returned order is from newer to older.
//
// Args:
//   - offset: The offset time to start retrieving the unban list.
//   - amount: The number of unbanned users to retrieve.
//
// Returns:
//   - []Unblocked: A list of unblocked users.
//   - error: An error if retrieving the unban list fails.
func (g *Group) GetUnbanList(offset time.Time, ammount int) (unbanList []models.Unblocked, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "unblocklist":
			var fields []string
			var target string
			var t time.Time
			var unbanned models.Unblocked
			for _, entry := range strings.Split(data, ";") {
				fields = strings.SplitN(entry, ":", 5)
				target = fields[2]
				if target == "" {
					target = "anon"
				}
				t, _ = utils.ParseTime(fields[3])
				unbanned = models.Unblocked{
					IP:           fields[1],
					ModerationID: fields[0],
					Target:       target,
					Unblocker:    fields[4],
					Time:         t,
				}
				unbanList = append(unbanList, unbanned)
			}
			return false
		default:
			g.events <- frame
		}
		return true
	}

	offsetS := fmt.Sprintf("%d", offset.Unix())
	ammountS := fmt.Sprintf("%d", ammount)

	err = g.SyncSend(cb, "blocklist", "unblock", offsetS, "next", ammountS, "anons", "1", "\r\n")

	return
}

// UnbanUser unblocks the specified blocked user.
//
// Args:
//   - blocked: The blocked user details to unblock.
//
// Returns:
//   - error: An error if unblocking the user fails.
func (g *Group) UnbanUser(blocked *models.Blocked) (err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "unblocked":
			g.events <- frame
			moderationID, _, _ := strings.Cut(data, ":")
			if moderationID == blocked.ModerationID {
				return false
			}
		default:
			g.events <- frame
		}
		return true
	}

	err = g.SyncSend(cb, "removeblock", blocked.ModerationID, blocked.IP, "\r\n")

	return
}

// UnbanAll unblocks all blocked users.
//
// Returns:
//   - int: The amount of unblocked users.
//   - error: An error if unblocking all users fails.
func (g *Group) UnbanAll() (amount int, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "allunblocked":
			amount, _ = strconv.Atoi(data)
			g.events <- frame
			return false
		default:
			g.events <- frame
		}
		return true
	}

	err = g.SyncSend(cb, "unbanall", "\r\n")

	return
}

// Login logs in to the group with the provided username and password.
//
// If the password is an empty string, it will log in as the named anon instead.
//
// Args:
//   - username: The username to log in with.
//   - password: The password to log in with. If empty, logs in as an anonymous user.
//
// Returns:
//   - error: An error if the login fails.
func (g *Group) Login(username, password string) (err error) {
	cb := func(frame string) bool {
		head, _, _ := strings.Cut(frame, ":")
		switch head {
		case "badalias":
			// badalias:8 (reserved for anons)
			err = ErrBadAlias
			return false
		case "aliasok":
			g.LoginName = username
			return false
		case "badlogin":
			// badlogin:2 (wrong password)
			err = ErrBadLogin
			return false
		case "pwdok":
			g.LoginName = username
			g.LoggedIn = true
			return false
		default:
			g.events <- frame
		}
		return true
	}

	var err2 error
	if password != "" {
		err2 = g.SyncSend(cb, "blogin", username, password, "\r\n")
	} else {
		err2 = g.SyncSend(cb, "blogin", username, "\r\n")
	}
	if err == nil && err2 != nil {
		err = err2
	}

	return
}

// Logout logs out from the group.
//
// Returns:
//   - error: An error if the logout fails.
func (g *Group) Logout() (err error) {
	cb := func(frame string) bool {
		head, _, _ := strings.Cut(frame, ":")
		switch head {
		case "logoutok":
			if access, ok := g.Moderators.Get(strings.ToLower(g.LoginName)); ok && access > 0 {
				// If the used account was a moderator, reload it.
				defer func() {
					g.Messages.Clear()
					g.TempMessages.Clear()
					g.TempMessageIds.Clear()
					g.Moderators.Clear()
					g.Send("reload_init_batch", "\r\n")
				}()
			}

			g.LoginName = utils.GetAnonName(int(g.LoginTime.Unix()), g.UserID)
			g.LoggedIn = false
			return false
		default:
			g.events <- frame
		}
		return true
	}

	err = g.SyncSend(cb, "blogout", "\r\n")

	return
}

// GetBanWords retrieves the banned word settings for the group.
//
// Returns:
//   - BanWord: The banned word settings for the group.
//   - error: An error if retrieving the banned word settings fails.
func (g *Group) GetBanWords() (banWord models.BanWord, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "bw":
			partial, exact, _ := strings.Cut(data, ":")
			banWord = models.BanWord{WholeWords: exact, Words: partial}
			return false
		default:
			g.events <- frame
		}
		return true
	}

	err = g.SyncSend(cb, "getbannedwords", "\r\n")

	return
}

// SetBanWords sets the banned word settings for the group.
//
// Args:
//   - banWord: The banned word settings to set.
//
// Returns:
//   - error: An error if setting the banned word settings fails.
func (g *Group) SetBanWords(banWord models.BanWord) (err error) {
	cb := func(frame string) bool {
		head, _, _ := strings.Cut(frame, ":")
		switch head {
		case "ubw":
			return false
		default:
			g.events <- frame
		}
		return true
	}

	err = g.SyncSend(cb, "\x00setbannedwords", banWord.Words, banWord.WholeWords, "\r\n")

	return
}

// Delete deletes the specified message from the group.
//
// Args:
//   - message: The message to delete.
//
// Returns:
//   - error: An error if deleting the message fails.
func (g *Group) Delete(message *Message) (err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "delete":
			g.events <- frame
			if data == message.ID {
				return false
			}
		default:
			g.events <- frame
		}
		return true
	}

	err = g.SyncSend(cb, "delmsg", message.ID, "\r\n")

	return
}

// DeleteAll deletes all messages in the group.
//
// Args:
//   - message: The message details to match for deletion.
//
// Returns:
//   - error: An error if deleting all messages fails.
func (g *Group) DeleteAll(message *Message) (err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "deleteall":
			g.events <- frame
			for _, messageID := range strings.Split(data, ":") {
				if messageID == message.ID {
					return false
				}
			}
		default:
			g.events <- frame
		}
		return true
	}

	err = g.SyncSend(cb, "delallmsg", message.ModerationID, message.UserIP, message.User.Name, "\r\n")

	return
}

// ClearAll clears all messages in the group.
//
// Returns:
//   - error: An error if clearing all messages fails.
func (g *Group) ClearAll() (err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "clearall":
			g.events <- frame
			if data == "error" {
				err = ErrRequestFailed
			}
			return false
		default:
			g.events <- frame
		}
		return true
	}

	if err2 := g.SyncSend(cb, "clearall", "\r\n"); err == nil && err2 != nil {
		err = err2
	}

	return
}

// AddModerator adds a moderator to the group with the specified username and access level.
//
// Args:
//   - username: The username of the new moderator.
//   - access: The access level to assign to the moderator.
//
// Returns:
//   - error: An error if adding the moderator fails.
func (g *Group) AddModerator(username string, access int64) (err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "mods":
			g.events <- frame
			var uname string
			for _, entry := range strings.Split(data, ":") {
				uname, _, _ = strings.Cut(entry, ",")
				if strings.EqualFold(uname, username) {
					return false
				}
			}
			err = ErrRequestFailed
			return false
		case "addmoderr":
			err = ErrRequestFailed
			return false
		default:
			g.events <- frame
		}
		return true
	}

	if err2 := g.SyncSend(cb, "addmod", username, fmt.Sprintf("%d", access), "\r\n"); err == nil && err2 != nil {
		err = err2
	}

	return
}

// UpdateModerator updates the access level of the specified moderator.
//
// Args:
//   - username: The name of the moderator to update.
//   - access: The new access level to set for the moderator.
//
// Returns:
//   - error: An error if the update fails.
func (g *Group) UpdateModerator(username string, access int64) (err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "mods":
			g.events <- frame
			var uname, flag string
			var flagInt int64
			for _, entry := range strings.Split(data, ":") {
				uname, flag, _ = strings.Cut(entry, ",")
				flagInt, _ = strconv.ParseInt(flag, 10, 64)
				if strings.EqualFold(uname, username) && flagInt == access {
					return false
				}
			}
		case "updatemoderr":
			err = ErrRequestFailed
			return false
		default:
			g.events <- frame
		}
		return true
	}

	if err2 := g.SyncSend(cb, "addmod", username, fmt.Sprintf("%d", access), "\r\n"); err == nil && err2 != nil {
		err = err2
	}

	return
}

// RemoveModerator removes the specified moderator from the group.
//
// Args:
//   - username: The name of the moderator to remove.
//
// Returns:
//   - error: An error if the removal fails.
func (g *Group) RemoveModerator(username string) (err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "mods":
			g.events <- frame
			var uname string
			for _, entry := range strings.Split(data, ":") {
				uname, _, _ = strings.Cut(entry, ",")
				if strings.EqualFold(uname, username) {
					err = ErrRequestFailed
					return false
				}
			}
			return false
		case "removemoderr":
			err = ErrRequestFailed
			return false
		default:
			g.events <- frame
		}
		return true
	}

	if err2 := g.SyncSend(cb, "removemod", username, "\r\n"); err == nil && err2 != nil {
		err = err2
	}

	return
}

// GetModActions retrieves a list of moderator actions (mod actions) for the group.
//
// Args:
//   - dir: The direction to retrieve the mod actions, "prev" for later or "next" for earlier actions.
//   - offset: The ID offset to retrieve mod actions from [ModAction.ID]. Later ID - 1 for "prev" or earlier ID + 1 for "next".
//
// Returns:
//   - []*ModAction: A list of ModAction objects representing the moderator actions.
//   - error: An error if retrieving the mod actions fails.
func (g *Group) GetModActions(dir string, offset int) (modactions []*models.ModAction, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "modactions":
			modactions = models.ParseModActions(data)
			return false
		default:
			g.events <- frame
		}
		return true
	}

	err = g.SyncSend(cb, "getmodactions", dir, fmt.Sprintf("%d", offset), "50", "\r\n")

	return
}

// GetLastUserMessage retrieves the last message sent by the specified username in the group.
//
// Note:
//   - This is useful for banning a user by their username.
//
// Args:
//   - username: The username to retrieve the last message for.
//
// Returns:
//   - *Message: The last Message object sent by the specified user.
//   - bool: A boolean indicating if a message was found.
func (g *Group) GetLastUserMessage(username string) (msg *Message, ok bool) {
	cb := func(_ string, v *Message) bool {
		if strings.EqualFold(v.User.Name, username) {
			msg = v
			ok = true
			return false
		}
		return true
	}

	g.Messages.RangeReversed(cb)

	return
}

// getMoreHistory retrieves additional history messages from the group.
//
// The offset starts with 0 from the latest messages, then [nextOffset = prevOffset + amount].
//
// Args:
//   - offset: The offset from which to start retrieving messages.
//   - amount: The number of history messages to fetch.
//
// Returns:
//   - int: The count of received messages.
//   - bool: A boolean indicating if there are no more messages to fetch.
//   - error: An error if the operation encounters any issues.
func (g *Group) getMoreHistory(offset, amount int) (count int, nomore bool, err error) {
	cb := func(frame string) bool {
		head, _, _ := strings.Cut(frame, ":")
		switch head {
		case "i":
			g.events <- frame
			count++
		case "gotmore":
			return false
		case "nomore":
			nomore = true
		default:
			g.events <- frame
		}
		return true
	}

	err = g.SyncSend(cb, "get_more", fmt.Sprintf("%d", amount), fmt.Sprintf("%d", offset), "\r\n")

	return
}

// ProfileRefresh notifies the server to refresh the profile.
//
// Returns:
//   - error: An error if the operation fails.
func (g *Group) ProfileRefresh() error {
	cb := func(frame string) bool {
		head, _, _ := strings.Cut(frame, ":")
		switch head {
		case "miu":
			g.events <- frame
			return false
		default:
			g.events <- frame
		}
		return true
	}

	return g.SyncSend(cb, "miu", "\r\n")
}

// PropagateEvent propagates the WebSocket frame back to the listener goroutine.
//
// It is utilized when the [Group.SyncSend] callback receives unwanted frames.
//
// Args:
//   - frame: The WebSocket frame to propagate.
func (g *Group) PropagateEvent(frame string) {
	g.events <- frame
}

// GetContext returns the [context.Context] of the group chat.
//
// Returns:
//   - context.Context: The context of the group chat.
func (g *Group) GetContext() context.Context {
	return g.context
}

// wsOnError handles WebSocket errors that occur during communication.
//
// It attempts to reconnect if the connection is still active.
// If the reconnection fails, it disconnects and dispatches the [OnGroupLeft] event.
//
// Args:
//   - err: The error that occurred.
func (g *Group) wsOnError(err error) {
	close(g.events)
	close(g.takeOver)
	if g.Connected {
		if g.Reconnect() == nil {
			log.Debug().Str("Name", g.Name).Msg("Reconnected")
			event := &Event{
				Type:  OnGroupReconnected,
				Group: g,
			}
			g.App.dispatchEvent(event)
			return
		}
	}
	g.Disconnect()
	log.Debug().Str("Name", g.Name).Msg("Disconnected")
	g.App.Groups.Del(g.Name)

	event := &Event{
		Type:  OnGroupLeft,
		Group: g,
	}
	g.App.dispatchEvent(event)
}

// wsOnFrame handles incoming WebSocket frames.
//
// Args:
//   - frame: The WebSocket frame to handle.
func (g *Group) wsOnFrame(frame string) {
	defer func() {
		if err := recover(); err != nil {
			log.Error().Str("Name", g.Name).Str("Frame", frame).Msgf("Error: %s", err)

			event := &Event{
				Type:  OnError,
				Group: g,
				Error: err,
			}
			g.App.dispatchEvent(event)
		}
	}()

	head, data, _ := strings.Cut(frame, ":")
	switch head {
	case "":
		// pong
	case "v":
		g.eventVersion(data)
	case "ok":
		g.eventOK(data)
	case "i":
		g.eventMessageHistory(data)
	case "inited":
		g.eventInited(data)
	case "n":
		g.eventParticipantCount(data)
	case "b":
		g.eventMessage(data)
	case "u":
		g.eventMessageUpdate(data)
	case "end_fw", "end_nlp":
		g.eventRestrictUpdate(data)
	case "participant":
		g.eventParticipant(data)
	case "groupflagsupdate":
		g.eventFlagsUpdate(data)
	case "annc":
		g.eventAnnouncement(data)
	case "mods":
		g.eventModerators(data)
	case "delete":
		g.eventMessageDelete(data)
	case "deleteall":
		g.eventMessageDeleteAll(data)
	case "clearall":
		g.eventMessageClearAll(data)
	case "blocked":
		g.eventUserBlocked(data)
	case "unblocked":
		g.eventUserUnblocked(data)
	case "allunblocked":
		g.eventAllUserUnblocked(data)
	case "updgroupinfo":
		g.eventUpdateGroupInfo(data)
	case "miu", "updateprofile":
		g.eventUpdateUserProfile(data)
	case "show_fw", "show_tb", "tb", "show_nlp", "show_nlp_tb", "nlptb":
		fallthrough
	case "msglexceeded", "ratelimited", "mustlogin", "proxybanned", "verificationrequired":
		fallthrough
	case "gparticipants", "getratelimit", "ratelimitset", "getannc", "groupflagstoggled":
		fallthrough
	case "premium", "blocklist", "bansearchresult", "badbansearchstring", "unblocklist":
		fallthrough
	case "badalias", "aliasok", "badlogin", "pwdok", "logoutok":
		fallthrough
	case "bw", "ubw":
		fallthrough
	case "addmoderr", "updatemoderr", "removemoderr":
		fallthrough
	case "modactions", "gotmore", "nomore":
		// This occurs when the `g.SyncSend` fails to capture these events.
		// I'm leaving this here for debugging purposes.
		log.Debug().Str("Name", g.Name).Str("Frame", frame).Msg("Uncaptured")
	default:
		// I'm not familiar with the purpose of these events, but I discovered them in the HTML source code.
		// "g_participants", Similar to "gparticipants"?
		// "chatango",
		// "p",
		// "cbw",
		// "notifysettings",
		// "setnotifysettings",
		// "checkemail_notify",
		// "limitexceeded", This event might be triggered when there are too many open connections to the server.
		// "verificationchanged", Related to "verificationrequired"?
		log.Debug().Str("Name", g.Name).Str("Frame", frame).Msg("Unknown")
	}
}

// eventVersion handles the version event.
func (g *Group) eventVersion(data string) {
	major, minor, _ := strings.Cut(data, ":")
	// Lowest compatible version, disconnect if > 15
	g.Version[0], _ = strconv.Atoi(major)
	// Current version, shows warning if > 15
	g.Version[1], _ = strconv.Atoi(minor)
}

// eventOK handles the OK event.
func (g *Group) eventOK(data string) {
	fields := strings.SplitN(data, ":", 8)
	g.Owner = fields[0]
	g.SessionID = fields[1]
	g.UserID, _ = strconv.Atoi(fields[1][:8])
	g.LoggedIn = fields[2] == "M" // "C" if anon
	g.LoginTime, _ = utils.ParseTime(fields[4])

	if fields[3] != "" {
		g.LoginName = fields[3]
	} else {
		g.LoginName = utils.GetAnonName(int(g.LoginTime.Unix()), g.UserID)
	}

	g.TimeDiff = time.Since(g.LoginTime)
	g.LoginIp = fields[5]

	var username, flag string
	var flagInt int64
	if fields[6] != "" {
		for _, perm := range strings.Split(fields[6], ";") {
			username, flag, _ = strings.Cut(perm, ",")
			flagInt, _ = strconv.ParseInt(flag, 10, 64)
			g.Moderators.Set(username, flagInt)
		}
	}

	g.Flag, _ = strconv.ParseInt(fields[7], 10, 64)

	if g.App.Config.EnablePM {
		go g.SetBackground(true)
	}

	event := &Event{
		Type:  OnGroupJoined,
		Group: g,
	}
	g.App.dispatchEvent(event)
}

// eventMessageHistory handles the message history event.
func (g *Group) eventMessageHistory(data string) {
	message := ParseGroupMessage(data, g)
	g.Messages.SetFront(message.ID, message)

	event := &Event{
		Type:    OnMessageHistory,
		Group:   g,
		Message: message,
		User:    message.User,
	}
	g.App.dispatchEvent(event)
}

// eventInited handles the initialized event.
func (g *Group) eventInited(string) {
	// TODO: loading previous messages?
	var offset int
	var count int
	var nomore bool
	var err error
	for histLen := g.Messages.Len(); histLen < MAX_MESSAGE_HISTORY && err == nil && !nomore; histLen += count {
		count, nomore, err = g.getMoreHistory(offset, utils.Min(20, MAX_MESSAGE_HISTORY-histLen))
		offset++
	}
}

// eventParticipantCount handles the participant count change event.
func (g *Group) eventParticipantCount(data string) {
	g.ParticipantCount, _ = strconv.ParseInt(data, 16, 64)

	event := &Event{
		Type:  OnParticipantCountChange,
		Group: g,
	}
	g.App.dispatchEvent(event)
}

// eventMessage handles the message event.
func (g *Group) eventMessage(data string) {
	message := ParseGroupMessage(data, g)
	if id, ok := g.TempMessageIds.Get(message.ID); ok {
		g.TempMessageIds.Del(message.ID)
		message.ID = id
		g.Messages.Set(message.ID, message)
		g.Messages.TrimFront(MAX_MESSAGE_HISTORY)

		event := &Event{
			Type:    OnMessage,
			Group:   g,
			Message: message,
			User:    message.User,
		}
		g.App.dispatchEvent(event)
	} else {
		g.TempMessages.Set(message.ID, message)
	}
}

// eventMessageUpdate handles the message update event.
func (g *Group) eventMessageUpdate(data string) {
	oldID, newID, _ := strings.Cut(data, ":")
	if message, ok := g.TempMessages.Get(oldID); ok {
		message.ID = newID
		g.Messages.Set(newID, message)
		g.Messages.TrimFront(MAX_MESSAGE_HISTORY)
		g.TempMessages.Del(oldID)

		event := &Event{
			Type:    OnMessage,
			Group:   g,
			Message: message,
			User:    message.User,
		}
		g.App.dispatchEvent(event)
	} else {
		g.TempMessageIds.Set(oldID, newID)
	}
}

// eventRestrictUpdate handles the restrict update event.
func (g *Group) eventRestrictUpdate(data string) {
	dur, _ := time.ParseDuration(data + "m")
	g.Restrict = time.Now().Add(dur)
}

// eventParticipant handles the participant event.
func (g *Group) eventParticipant(data string) {
	fields := strings.SplitN(data, ":", 7)
	var user *models.User
	t, _ := utils.ParseTime(fields[6])
	userID, _ := strconv.Atoi(fields[2])

	if fields[3] != "None" {
		user = &models.User{Name: fields[3]}
	} else if fields[4] != "None" {
		user = &models.User{Name: fields[4], IsAnon: true}
	} else {
		user = &models.User{Name: utils.GetAnonName(int(t.Unix()), userID), IsAnon: true}
	}

	user.IsSelf = userID == g.UserID && user.Name == g.LoginName

	p := &models.Participant{
		ParticipantID: fields[1],
		UserID:        userID,
		User:          user,
		Time:          t,
	}

	event := &Event{
		Group: g,
		User:  user,
	}

	switch fields[0] {
	case "1":
		g.Participants.Set(fields[1], p)
		event.Type = OnJoin
		event.Participant = p
		if p.User.IsAnon {
			g.AnonCount++
		} else {
			g.UserCount++
		}
	case "2":
		oldParticipant, ok := g.Participants.Get(fields[1])
		g.Participants.Set(fields[1], p)
		if !p.User.IsAnon {
			event.Type = OnLogin
			event.Participant = p
			g.AnonCount--
			g.UserCount++
		} else if ok && !oldParticipant.User.IsAnon {
			event.Type = OnLogout
			event.Participant = oldParticipant
			g.AnonCount++
			g.UserCount--
		}
	case "0":
		g.Participants.Del(fields[1])
		event.Type = OnLeave
		event.Participant = p
		if p.User.IsAnon {
			g.AnonCount--
		} else {
			g.UserCount--
		}
	}

	g.App.dispatchEvent(event)
}

// eventFlagsUpdate handles the flags update event.
func (g *Group) eventFlagsUpdate(data string) {
	newFlag, _ := strconv.ParseInt(data, 10, 64)
	// Compute the changes
	added, removed := utils.ComputeFlagChanges(g.Flag, newFlag)

	event := &Event{
		Type:        OnFlagUpdate,
		Group:       g,
		FlagAdded:   added,
		FlagRemoved: removed,
	}
	g.Flag = newFlag
	g.App.dispatchEvent(event)
}

// eventAnnouncement handles the announcement event.
func (g *Group) eventAnnouncement(data string) {
	event := &Event{
		Type:    OnAnnouncement,
		Group:   g,
		Message: ParseAnnouncement(data, g),
	}
	g.App.dispatchEvent(event)
}

// eventModerators handles the moderators event.
func (g *Group) eventModerators(data string) {
	var (
		newMods                          = make(map[string]int64)
		events                           []*Event
		user                             *models.User
		username, uname, flag            string
		newFlag, oldFlag, added, removed int64
		ok, selfRemoved                  bool
		event                            *Event
	)

	// Process moderator addition and changes
	for _, entry := range strings.Split(data, ":") {
		username, flag, _ = strings.Cut(entry, ",")
		user = &models.User{Name: username, IsSelf: strings.EqualFold(username, g.LoginName)}
		newFlag, _ = strconv.ParseInt(flag, 10, 64)
		newMods[username] = newFlag

		// Check if user is already a moderator
		if oldFlag, ok = g.Moderators.Get(username); ok {
			if oldFlag == newFlag {
				// No changes, skip to next entry
				continue
			}

			// Compute the changes
			added, removed = utils.ComputeFlagChanges(oldFlag, newFlag)
			event = &Event{
				Type:             OnModeratorUpdated,
				Group:            g,
				User:             user,
				ModGrantedAccess: added,
				ModRevokedAccess: removed,
			}
		} else {
			event = &Event{
				Type:  OnModeratorAdded,
				Group: g,
				User:  user,
			}
			if user.IsSelf {
				defer func() {
					g.Messages.Clear()
					g.TempMessages.Clear()
					g.TempMessageIds.Clear()
					g.Send("reload_init_batch", "\r\n")
				}()
			}
		}
		events = append(events, event)
	}

	// Process moderator removal
	cb := func(username string, oldFlag int64) bool {
		for uname = range newMods {
			if uname == username {
				return false
			}
		}

		// When it reaches this scope, it means the username has been removed.
		user = &models.User{Name: username, IsSelf: strings.EqualFold(username, g.LoginName)}
		event = &Event{
			Type:  OnModeratorRemoved,
			Group: g,
			User:  user,
		}
		events = append(events, event)
		if user.IsSelf {
			selfRemoved = true
		}
		return false
	}
	g.Moderators.Range(cb)

	if selfRemoved {
		defer func() {
			g.Messages.Clear()
			g.TempMessages.Clear()
			g.TempMessageIds.Clear()
			g.Moderators.Clear()
			g.Send("reload_init_batch", "\r\n")
		}()
	} else {
		// Update the g.Moderators
		g.Moderators.Lock()
		g.Moderators.M = newMods
		g.Moderators.Unlock()
	}

	for _, event = range events {
		g.App.dispatchEvent(event)
	}
}

// eventMessageDelete handles the message delete event.
func (g *Group) eventMessageDelete(data string) {
	msg, ok := g.Messages.Get(data)
	if ok {
		g.Messages.Del(data)

		event := &Event{
			Type:    OnMessageDelete,
			Group:   g,
			Message: msg,
			User:    msg.User,
		}
		g.App.dispatchEvent(event)
	}
}

// eventMessageDeleteAll handles the message delete all event.
func (g *Group) eventMessageDeleteAll(data string) {
	for _, id := range strings.Split(data, ":") {
		g.eventMessageDelete(id)
	}
}

// eventMessageClearAll handles the message clear all event.
func (g *Group) eventMessageClearAll(data string) {
	switch data {
	case "ok":
		g.Messages.Clear()
		g.TempMessages.Clear()
		g.TempMessageIds.Clear()

		event := &Event{
			Type:  OnClearAll,
			Group: g,
		}
		g.App.dispatchEvent(event)
	case "error":
		// This event is fired when this account does not have permission to clear all.
	}
}

// eventUserBlocked handles the user blocked event.
func (g *Group) eventUserBlocked(data string) {
	fields := strings.SplitN(data, ":", 5)
	t, _ := utils.ParseTime(fields[4])
	blocked := &models.Blocked{
		ModerationID: fields[0],
		IP:           fields[1],
		Target:       fields[2],
		Blocker:      fields[3],
		Time:         t,
	}

	event := &Event{
		Type:    OnUserBanned,
		Group:   g,
		Blocked: blocked,
	}
	g.App.dispatchEvent(event)
}

// eventUserUnblocked handles the user unblocked event.
func (g *Group) eventUserUnblocked(data string) {
	fields := strings.SplitN(data, ":", 5)
	t, _ := utils.ParseTime(fields[4])
	unblocked := &models.Unblocked{
		ModerationID: fields[0],
		IP:           fields[1],
		Target:       fields[2],
		Unblocker:    fields[3],
		Time:         t,
	}

	event := &Event{
		Type:      OnUserUnbanned,
		Group:     g,
		Unblocked: unblocked,
	}
	g.App.dispatchEvent(event)
}

// eventAllUserUnblocked handles the all user unblocked event.
func (g *Group) eventAllUserUnblocked(string) {
	event := &Event{
		Type:  OnAllUserUnbanned,
		Group: g,
	}
	g.App.dispatchEvent(event)
}

// eventUpdateGroupInfo handles the update group info event.
func (g *Group) eventUpdateGroupInfo(data string) {
	title, message, _ := strings.Cut(data, ":")
	groupinfo := &models.GroupInfo{
		Title:        title,
		OwnerMessage: message,
	}

	event := &Event{
		Type:      OnUpdateGroupInfo,
		Group:     g,
		GroupInfo: groupinfo,
	}
	g.App.dispatchEvent(event)
}

// eventUpdateGroupInfo handles the update user profile event.
func (g *Group) eventUpdateUserProfile(data string) {
	event := &Event{
		Type:  OnUpdateUserProfile,
		Group: g,
		User:  &models.User{Name: data, IsSelf: strings.EqualFold(data, g.LoginName)},
	}
	g.App.dispatchEvent(event)
}
