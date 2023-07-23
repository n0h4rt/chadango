package chadango

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// Group represents a chat group with various properties and state.
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

	Participants     SyncMap[string, *Participant] // Map of participants in the group. Invoke `g.GetParticipantsStart` to initiate the participant feeds.
	ParticipantCount int64                         // The total count of participants in the group.
	UserCount        int                           // The count of registered users in the group.
	AnonCount        int                           // The count of anonymous users in the group.
}

func (g *Group) initFields() {
	g.Moderators = NewSyncMap[string, int64]()
	g.Messages = NewOrderedSyncMap[string, *Message]()
	g.TempMessages = NewSyncMap[string, *Message]()
	g.TempMessageIds = NewSyncMap[string, string]()
	g.Participants = NewSyncMap[string, *Participant]()
}

// Connect establishes a connection to the server.
// It returns an error if the connection cannot be established.
func (g *Group) Connect(ctx context.Context) (err error) {
	if g.Connected {
		return ErrAlreadyConnected
	}

	g.context, g.cancelCtx = context.WithCancel(ctx)

	log.Debug().Str("Name", g.Name).Msg("Connecting")

	err = g.connect()
	if err != nil {
		g.cancelCtx()
	}

	g.Connected = true

	return
}

func (g *Group) connect() (err error) {
	defer func() {
		if err != nil {
			if g.ws != nil {
				g.ws.Close()
			}

			log.Debug().Str("Name", g.Name).Msg("Connect failed")
		}
	}()

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

	log.Debug().Str("Name", g.Name).Msg("Connected")

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

// Reconnect attempts to reconnect to the server.
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

// Send will join the `args` with a ":" separator and then send it to the server asynchronously.
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

// SyncSend will send the `args` and wait until receiving the correct reply or until timeout.
// First, a `g.takeOver` request will be made and it will wait until the listener goroutine catches it.
// Then, the `args` will be sent to the server.
// Each time a frame is received, the `callback` function is invoked and passed the frame.
// The `callback` should return `false` if a correct frame is acquired, and `true` otherwise.
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

// SyncSend will send the `args` and wait until receiving the correct reply or until timeout (default to 5 seconds).
// For more information, refer to the documentation of `g.SyncSendWithTimeout`.
func (g *Group) SyncSend(cb func(string) bool, text ...string) error {
	return g.SyncSendWithTimeout(cb, SYNC_SEND_TIMEOUT, text...)
}

// SendMessage sends a message to the group with the specified text and optional arguments.
// It returns a pointer to the sent `*Message` and an error (if any occurred).
// The text can include formatting placeholders (%s, %d, etc.), and optional arguments can be provided to fill in these placeholders.
// The function also handles flood warnings, restricted messages, spam warnings, rate limiting, and other Chatango-specific events.
// If the group is logged in, the message text is styled with the group's name color, text size, text color, and text font.
// If the group is anonymous, the message text is modified with the anonymous seed based on the group's `AnonName` and `UserID`.
// The function replaces newlines with the `<br/>` HTML tag to format the message properly.
func (g *Group) SendMessage(text string, a ...any) (msg *Message, err error) {
	// The received "b" and "u" frames should be returned to `g.Events`.
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

	// I'm not sure what it is for, but it gets sent back to the client when `climited` occurs.
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
		text = fmt.Sprintf(`<n%d/>%s`, CreateAnonSeed(g.AnonName, g.UserID), text)
	}

	// Replacing newlines with the `<br/>` tag.
	text = strings.ReplaceAll(text, "\r\n", "<br/>")
	text = strings.ReplaceAll(text, "\n", "<br/>")

	if err2 := g.SyncSend(cb, "bm", randomString, fmt.Sprintf("%d", g.Channel), text, "\r\n"); err == nil && err2 != nil {
		err = err2
	}

	return
}

// SendMessage sends the chunked `text` with a size of `chunkSize` and returns the sent `[]*Message`.
// In the event of an error, the already sent messages will be returned along with the error for the unsent message.
func (g *Group) SendMessageChunked(text string, chunkSize int) (msgs []*Message, err error) {
	var msg *Message
	for _, chunk := range SplitTextIntoChunks(text, chunkSize) {
		if msg, err = g.SendMessage(chunk); err != nil {
			return
		}
		msgs = append(msgs, msg)
	}

	return
}

// IsRestricted checks if the group is restricted.
// The restriction can originate from either a flood ban or a rate limit.
func (g *Group) IsRestricted() bool {
	return g.Restrict.After(time.Now()) || g.RateLimited.After(time.Now())
}

// GetParticipantsStart initiates the `participant` event feeds and returns the current participants.
// The `participant` event will be triggered when there is user activity, such as joining, logging out, logging in, or leaving.
func (g *Group) GetParticipantsStart() (p *SyncMap[string, *Participant], err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "gparticipants":
			g.Participants.Clear()
			anoncount, entries, _ := strings.Cut(data, ":")
			g.AnonCount, _ = strconv.Atoi(anoncount)

			var fields []string
			var user *User
			var t time.Time
			var participant *Participant
			for _, entry := range strings.Split(entries, ";") {
				fields = strings.SplitN(entry, ":", 6)
				t, _ = ParseTime(fields[1])
				userID, _ := strconv.Atoi(fields[2])
				if fields[3] != "None" {
					user = &User{Name: fields[3]}
				} else if fields[4] != "None" {
					user = &User{Name: fields[4], IsAnon: true}
				} else {
					user = &User{Name: GetAnonName(int(t.Unix()), userID), IsAnon: true}
				}
				user.IsSelf = userID == g.UserID && user.Name == g.LoginName
				participant = &Participant{
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
// This will leave g.Participants, g.UserCount, g.AnonCount out of date.
func (g *Group) GetParticipantsStop() error {
	return g.Send("gparticipants", "stop", "\r\n")
}

// GetRateLimit retrieves the rate limit settings for the group.
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
func (g *Group) SetAnnouncement(annc string, enable bool, interval time.Duration) error {
	return g.Send("updateannouncement", BoolZeroOrOne(enable), fmt.Sprintf("%.0f", interval.Seconds()), "\r\n")
}

// UpdateGroupFlag updates the group's flag by adding and removing specific flags.
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
// This function would activate server validation for the premium status.
// For example, the message background and media won't activate before this command is sent.
func (g *Group) GetPremiumInfo() (flag int, expire time.Time, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "premium":
			fl, ti, _ := strings.Cut(data, ":")
			flag, _ = strconv.Atoi(fl)
			expire, _ = ParseTime(ti)
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
// If enable is true, it enables the background feature for the group.
// If enable is false, it disables the background feature for the group.
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

	return g.Send("msgbg", BoolZeroOrOne(enable), "\r\n")
}

// SetMedia sets the media status of the group.
// If enable is true, it enables the media feature for the group.
// If enable is false, it disables the media feature for the group.
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

	return g.Send("msgmedia", BoolZeroOrOne(enable), "\r\n")
}

// GetBanList retrieves a list of blocked users (ban list) for the group.
// The offset can be set to zero time to retrieve the newest result.
// The returned order is from newer to older.
func (g *Group) GetBanList(offset time.Time, ammount int) (banList []Blocked, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "blocklist":
			var fields []string
			var target string
			var t time.Time
			var banned Blocked
			for _, entry := range strings.Split(data, ";") {
				fields = strings.SplitN(entry, ":", 5)
				target = fields[2]
				if target == "" {
					target = "anon"
				}
				t, _ = ParseTime(fields[3])
				banned = Blocked{
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
// The query can be either a user name or an IP address.
func (g *Group) SearchBannedUser(query string) (banned Blocked, ok bool, err error) {
	query = strings.TrimSpace(query)
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
			banned = Blocked{
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

	err = g.SyncSend(cb, "searchban", query, "\r\n")

	return
}

// BanUser bans the user associated with the specified message.
func (g *Group) BanUser(message *Message) (err error) {
	// The received frame should be returned to `g.Events`.
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
// The `offset` is taken from the earliest time in the previous result or zero `time.Time`.
// The `amount` corresponds to the desired number of results.
func (g *Group) GetUnbanList(offset time.Time, ammount int) (unbanList []Unblocked, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "unblocklist":
			var fields []string
			var target string
			var t time.Time
			var unbanned Unblocked
			for _, entry := range strings.Split(data, ";") {
				fields = strings.SplitN(entry, ":", 5)
				target = fields[2]
				if target == "" {
					target = "anon"
				}
				t, _ = ParseTime(fields[3])
				unbanned = Unblocked{
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
func (g *Group) UnbanUser(blocked *Blocked) (err error) {
	// The received frame should be returned to `g.Events`.
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
func (g *Group) UnbanAll() (err error) {
	// The received frame should be returned to `g.Events`.
	cb := func(frame string) bool {
		head, _, _ := strings.Cut(frame, ":")
		switch head {
		case "allunblocked":
			// allunblocked:2 (length of unblocked users)
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
// If the password is an empty string, it will log in as the named anon instead.
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

			g.LoginName = GetAnonName(int(g.LoginTime.Unix()), g.UserID)
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
func (g *Group) GetBanWords() (banWord BanWord, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "bw":
			partial, exact, _ := strings.Cut(data, ":")
			banWord = BanWord{WholeWords: exact, Words: partial}
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
func (g *Group) SetBanWords(banWord BanWord) (err error) {
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
func (g *Group) Delete(message *Message) (err error) {
	// The received frame should be returned to `g.Events`.
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
func (g *Group) DeleteAll(message *Message) (err error) {
	// The received frame should be returned to `g.Events`.
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
func (g *Group) ClearAll() (err error) {
	// The received frame should be returned to `g.Events`.
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
func (g *Group) AddModerator(username string, access int64) (err error) {
	// The received frame should be returned to `g.Events`.
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
func (g *Group) UpdateModerator(username string, access int64) (err error) {
	// The received frame should be returned to `g.Events`.
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
func (g *Group) RemoveModerator(username string) (err error) {
	// The received frame should be returned to `g.Events`.
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
// The `dir` can be either "prev" to go earlier or "next" to go to the latest.
// The `offset` corresponds to the ID of the earlier-1 or latest+1 `ModAction.ID`.
func (g *Group) GetModActions(dir string, offset int) (modactions []*ModAction, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "modactions":
			modactions = ParseModActions(data)
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
// This is useful for banning a user by their username.
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
// It takes an offset and the amount of history messages to fetch as parameters.
// It returns the count of received messages, a boolean indicating if there are no more messages to fetch,
// and an error if the operation encounters any issues.
func (g *Group) getMoreHistory(offset, amount int) (count int, nomore bool, err error) {
	// The received "i" frame should be returned to `g.Events`.
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
// It is utilized when the `g.SyncSend` callback receives unwanted frames.
func (g *Group) PropagateEvent(frame string) {
	g.events <- frame
}

// GetContext returns the context of the private chat.
func (g *Group) GetContext() context.Context {
	return g.context
}

// wsOnError handles WebSocket errors that occur during communication.
func (g *Group) wsOnError(e error) {
	close(g.events)
	close(g.takeOver)
	if g.Connected {
		if g.Reconnect() == nil {
			log.Debug().Str("Name", g.Name).Msg("Reconnected")
			event := &Event{
				Type:  OnGroupReconnected,
				Group: g,
			}
			g.App.Dispatch(event)
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
	g.App.Dispatch(event)
}

// wsOnFrame handles incoming WebSocket frames.
func (g *Group) wsOnFrame(frame string) {
	defer func() {
		if err := recover(); err != nil {
			log.Error().Str("Name", g.Name).Str("Frame", frame).Msgf("Error: %s", err)

			event := &Event{
				Type:  OnError,
				Group: g,
				Error: err,
			}
			g.App.Dispatch(event)
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
	case "miu":
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
	g.LoginTime, _ = ParseTime(fields[4])

	if fields[3] != "" {
		g.LoginName = fields[3]
	} else {
		g.LoginName = GetAnonName(int(g.LoginTime.Unix()), g.UserID)
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
	g.App.Dispatch(event)
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
	g.App.Dispatch(event)
}

// eventInited handles the initialized event.
func (g *Group) eventInited(data string) {
	// TODO: loading previous messages?
	var offset int
	var count int
	var nomore bool
	var err error
	for histLen := g.Messages.Len(); histLen < MAX_MESSAGE_HISTORY && err == nil && !nomore; histLen += count {
		count, nomore, err = g.getMoreHistory(offset, Min(20, MAX_MESSAGE_HISTORY-histLen))
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
	g.App.Dispatch(event)
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
		g.App.Dispatch(event)
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
		g.App.Dispatch(event)
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
	var user *User
	t, _ := ParseTime(fields[6])
	userID, _ := strconv.Atoi(fields[2])

	if fields[3] != "" {
		user = &User{Name: fields[3]}
	} else if fields[4] != "None" {
		user = &User{Name: fields[4], IsAnon: true}
	} else {
		user = &User{Name: GetAnonName(int(t.Unix()), userID), IsAnon: true}
	}

	user.IsSelf = userID == g.UserID && user.Name == g.LoginName

	p := &Participant{
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

	g.App.Dispatch(event)
}

// eventFlagsUpdate handles the flags update event.
func (g *Group) eventFlagsUpdate(data string) {
	newFlag, _ := strconv.ParseInt(data, 10, 64)
	// Compute the changes
	added, removed := ComputeFlagChanges(g.Flag, newFlag)

	event := &Event{
		Type:        OnFlagUpdate,
		Group:       g,
		FlagAdded:   added,
		FlagRemoved: removed,
	}
	g.Flag = newFlag
	g.App.Dispatch(event)
}

// eventAnnouncement handles the announcement event.
func (g *Group) eventAnnouncement(data string) {
	event := &Event{
		Type:    OnAnnouncement,
		Group:   g,
		Message: ParseAnnouncement(data, g),
	}
	g.App.Dispatch(event)
}

// eventModerators handles the moderators event.
func (g *Group) eventModerators(data string) {
	var (
		newMods                          = make(map[string]int64)
		events                           []*Event
		user                             *User
		username, uname, flag            string
		newFlag, oldFlag, added, removed int64
		ok, selfRemoved                  bool
		event                            *Event
	)

	// Process moderator addition and changes
	for _, entry := range strings.Split(data, ":") {
		username, flag, _ = strings.Cut(entry, ",")
		user = &User{Name: username, IsSelf: strings.EqualFold(username, g.LoginName)}
		newFlag, _ = strconv.ParseInt(flag, 10, 64)
		newMods[username] = newFlag

		// Check if user is already a moderator
		if oldFlag, ok = g.Moderators.Get(username); ok {
			if oldFlag == newFlag {
				// No changes, skip to next entry
				continue
			}

			// Compute the changes
			added, removed = ComputeFlagChanges(oldFlag, newFlag)
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
		user = &User{Name: username, IsSelf: strings.EqualFold(username, g.LoginName)}
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
		g.App.Dispatch(event)
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
		g.App.Dispatch(event)
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
		g.App.Dispatch(event)
	case "error":
		// This event is fired when this account does not have permission to clear all.
	}
}

// eventUserBlocked handles the user blocked event.
func (g *Group) eventUserBlocked(data string) {
	fields := strings.SplitN(data, ":", 5)
	t, _ := ParseTime(fields[4])
	blocked := &Blocked{
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
	g.App.Dispatch(event)
}

// eventUserUnblocked handles the user unblocked event.
func (g *Group) eventUserUnblocked(data string) {
	fields := strings.SplitN(data, ":", 5)
	t, _ := ParseTime(fields[4])
	unblocked := &Unblocked{
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
	g.App.Dispatch(event)
}

// eventAllUserUnblocked handles the all user unblocked event.
func (g *Group) eventAllUserUnblocked(data string) {
	event := &Event{
		Type:  OnAllUserUnbanned,
		Group: g,
	}
	g.App.Dispatch(event)
}

// eventUpdateGroupInfo handles the update group info event.
func (g *Group) eventUpdateGroupInfo(data string) {
	title, message, _ := strings.Cut(data, ":")
	groupinfo := &GroupInfo{
		Title:        title,
		OwnerMessage: message,
	}

	event := &Event{
		Type:      OnUpdateGroupInfo,
		Group:     g,
		GroupInfo: groupinfo,
	}
	g.App.Dispatch(event)
}

// eventUpdateGroupInfo handles the update user profile event.
func (g *Group) eventUpdateUserProfile(data string) {
	event := &Event{
		Type:  OnUpdateUserProfile,
		Group: g,
		User:  &User{Name: data, IsSelf: strings.EqualFold(data, g.LoginName)},
	}
	g.App.Dispatch(event)
}
