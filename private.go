package chadango

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/n0h4rt/chadango/models"
	"github.com/n0h4rt/chadango/utils"
	"github.com/rs/zerolog/log"
)

// Private represents a private message with various properties and state.
//
// It provides methods for connecting, disconnecting, sending messages, retrieving user status, and managing settings.
// The [Private] struct also handles events related to private messages, friend status, and user profile updates.
type Private struct {
	App       *Application // Reference to the application.
	Name      string       // The name of the group.
	NameColor string       // The color for displaying name in the message.
	TextColor string       // The color for displaying text in the message.
	TextFont  string       // The font style for displaying text in the message.
	TextSize  int          // The font size for displaying text in the message.

	WsUrl     string               // The WebSocket URL for connecting to the PM server.
	ws        *WebSocket           // The WebSocket connection to the PM server.
	Connected bool                 // Indicates if the PM server is currently connected.
	events    chan string          // Channel for propagating events back to the listener.
	takeOver  chan context.Context // Channel for taking over the WebSocket connection.
	backoff   *Backoff             // Cancelable backoff for reconnection.
	context   context.Context      // Context for running the private chat operations.
	cancelCtx context.CancelFunc   // Function for stopping private chat operations.

	token     string        // The auth token used to connect to the PM server.
	LoginName string        // The login name of the user.
	SessionID string        // The session ID for the PM server.
	LoginTime time.Time     // The time when the user logged in.
	TimeDiff  time.Duration // The time difference between the server and client (serverTime - clientTime).

	idleTimer *time.Timer // The timer for the idle command.
	IsIdle    bool        // Indicates whether there has been no activity within 1 minute (e.g., sending a message).
}

// Connect establishes a connection to the server.
//
// Args:
//   - ctx: The context for the connection.
//
// Returns:
//   - error: An error if the connection cannot be established.
func (p *Private) Connect(ctx context.Context) (err error) {
	if p.Connected {
		return ErrAlreadyConnected
	}

	p.context, p.cancelCtx = context.WithCancel(ctx)

	log.Debug().Str("Name", p.Name).Msg("Connecting")

	defer func() {
		if err != nil {
			if p.ws != nil {
				p.ws.Close()
			}
			log.Debug().Str("Name", p.Name).Err(err).Msg("Connect failed")
		}
	}()

	err = p.connect()
	if err != nil {
		p.cancelCtx()
		return
	}

	p.Connected = true

	log.Debug().Str("Name", p.Name).Msg("Connected")

	return
}

// connect establishes a WebSocket connection to the PM server.
//
// It performs the necessary login steps using the private API, retrieves the
// authentication token, and initializes the WebSocket connection and channels.
// The function attempts to send a login request and waits for an "OK" or "DENIED"
// response within 5 attempts. If the login is successful, it sustains the connection
// and starts listening for incoming events.
//
// Returns:
//   - error: An error if the connection cannot be established.
func (p *Private) connect() (err error) {
	if err = privateAPI.Login(); err != nil {
		return
	}

	var ok bool
	if p.token, ok = privateAPI.GetCookie("auth.chatango.com"); !ok {
		return ErrLoginFailed
	}

	p.ws = &WebSocket{
		OnError: p.wsOnError,
	}
	if err = p.ws.Connect(p.WsUrl); err != nil {
		return
	}

	// Initializing channels.
	p.events = make(chan string, EVENT_BUFFER_SIZE)
	p.takeOver = make(chan context.Context)

	var frame, head string

	if err = p.Send("tlogin", p.token, "2", p.App.Config.SessionID, "\x00"); err != nil {
		return
	}

	// Random responses may be received, expecting either "OK" or "DENIED" within 5 loops.
	for i := 0; i < 5; i++ {
		if frame, err = p.ws.Recv(); err != nil {
			return
		}
		frame = strings.TrimRight(frame, "\r\n\x00")
		head, _, _ = strings.Cut(frame, ":")
		switch head {
		case "OK":
			p.events <- frame
			goto OK
		case "DENIED":
			return ErrBadLogin
		default:
			p.events <- frame
		}
	}

	return ErrBadLogin

OK:
	p.ws.Sustain(p.context)
	go p.listen()

	return
}

// listen listens for incoming messages and events on the WebSocket connection.
func (p *Private) listen() {
	var frame string
	var ok bool
	var release context.Context
	for {
		select {
		case <-p.context.Done():
			return
		case frame, ok = <-p.events:
			if !ok {
				return
			}
			go p.wsOnFrame(frame)
		case frame, ok = <-p.ws.Events:
			if !ok {
				return
			}
			go p.wsOnFrame(strings.TrimRight(frame, "\r\n\x00"))
		case release, ok = <-p.takeOver:
			if !ok {
				return
			}
		inner:
			for {
				select {
				case <-p.context.Done():
					return
				case <-release.Done():
					break inner
				case frame, ok = <-p.events:
					if !ok {
						return
					}
					go p.wsOnFrame(frame)
				}
			}
		}
	}
}

// Disconnect gracefully closes the connection to the PM server.
func (p *Private) Disconnect() {
	if p.backoff != nil {
		p.backoff.Cancel()
	}

	if !p.Connected {
		return
	}

	p.Connected = false
	p.cancelCtx()
	p.ws.Close()
}

// Reconnect attempts to reconnect to the PM server.
//
// Returns:
//   - error: An error if the reconnection fails.
func (p *Private) Reconnect() (err error) {
	p.ws.Close()

	// Reinitialize the API.
	initAPI(p.App.Config.Username, p.App.Config.Password, p.App.context)

	p.backoff = &Backoff{
		Duration:    BASE_BACKOFF_DUR,
		MaxDuration: MAX_BACKOFF_DUR,
	}
	defer func() {
		p.backoff = nil
	}()
	for retries := 0; retries < MAX_RETRIES && !p.backoff.Sleep(p.context); retries++ {
		if err = p.connect(); err == nil {
			return
		}
	}

	// Either canceled or reached the maximum retries.
	return ErrRetryEnds
}

// Send will join the [args] with a ":" separator and then send it to the server asynchronously.
//
// Note:
//   - A terminator should be included in the last [args].
//   - The terminator can be "\r\n" or "\x00" depending on the command.
//
// Args:
//   - args: The arguments to send to the server.
//
// Returns:
//   - error: An error if the sending fails.
func (p *Private) Send(args ...string) error {
	if !p.ws.Connected {
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

	return p.ws.Send(command + terminator)
}

// SyncSendWithTimeout will send the [args] and wait until receiving the correct reply or until timeout.
//
// Args:
//   - callback: The function to call for each received frame.
//   - timeout: The timeout duration for the operation.
//   - args: The arguments to send to the server.
//
// Returns:
//   - error: An error if the operation fails.
func (p *Private) SyncSendWithTimeout(callback func(string) bool, timeout time.Duration, args ...string) (err error) {
	ctx, cancel := context.WithTimeout(p.context, timeout)
	defer cancel()

	// Make a takeover request to allow the listener to relinquish the connection.
	select {
	case <-ctx.Done():
		return ErrTimeout
	case p.takeOver <- ctx:
	}

	if err = p.Send(args...); err != nil {
		return
	}

	var frame string
	var ok bool
	for {
		select {
		case <-ctx.Done():
			return ErrTimeout
		case frame, ok = <-p.ws.Events:
			if !ok {
				close(p.events)
				return ErrConnectionClosed
			}
			if !callback(strings.TrimRight(frame, "\r\n\x00")) {
				return
			}
		}
	}
}

// SyncSend will send the [args] and wait until receiving the correct reply or until timeout (default to 5 seconds).
//
// Args:
//   - cb: The function to call for each received frame.
//   - args: The arguments to send to the server.
//
// Returns:
//   - error: An error if the operation fails.
func (p *Private) SyncSend(cb func(string) bool, args ...string) error {
	return p.SyncSendWithTimeout(cb, SYNC_SEND_TIMEOUT, args...)
}

// SendMessage sends a private message to the specified username with the given text and optional arguments.
//
// It returns an error if any occurs during the message sending process.
// The text can include formatting placeholders (%s, %d, etc.), and optional arguments can be provided to fill in these placeholders.
// The function also handles a flood warning, a flood ban, and the maximum number of unread messages (51) has been reached.
// The function replaces newlines with the `<br/>` HTML tag to format the message properly.
//
// Args:
//   - username: The username to send the message to.
//   - text: The message text.
//   - a: Optional arguments to fill in placeholders in the message text.
//
// Returns:
//   - error: An error if any occurs during the message sending process.\
func (p *Private) SendMessage(username, text string, a ...any) (err error) {
	cb := func(frame string) bool {
		head, _, _ := strings.Cut(frame, ":")
		switch head {
		case "show_fw":
			err = ErrFloodWarning
			return false
		case "toofast":
			err = ErrFloodBanned
			return false
		case "show_offline_limit":
			// The maximum number of unread messages is 51.
			// show_offline_limit:nekonyan
			err = ErrOfflineLimit
			return false
		default:
			p.events <- frame
		}
		return true
	}

	username = strings.ToLower(username)

	text = fmt.Sprintf(text, a...)
	text = fmt.Sprintf(`<n%s/><m v="1"><g x%02ds%s="%s">%s</g></m>`, p.NameColor, p.TextSize, p.TextColor, p.TextFont, text)

	// Replacing newlines with the `<br/>` tag.
	text = strings.ReplaceAll(text, "\r\n", "<br/>")
	text = strings.ReplaceAll(text, "\n", "<br/>")

	// The "msg" command does not produce a response, so we wait for 0.5 seconds to handle any potential errors that may occur.
	// The potential errors include "show_fw", "toofast", and "show_offline_limit".
	if err2 := p.SyncSendWithTimeout(cb, 500*time.Millisecond, "msg", username, text, "\r\n"); err != nil {
		return
	} else if err2 != ErrTimeout {
		return err2
	}

	// Notifies the PM server that the client has just been active.
	p.WentActive()

	return
}

// Track retrieves the online status of the username.
//
// Args:
//   - username: The username to track.
//
// Returns:
//   - UserStatus: The online status of the username.
//   - error: An error if the operation fails.
func (p *Private) Track(username string) (status models.UserStatus, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "track":
			fields := strings.SplitN(data, ":", 3)
			status.User = &models.User{Name: fields[0]}
			status.Info = fields[2]
			switch fields[2] {
			case "offline":
				status.Time, _ = utils.ParseTime(fields[1])
			case "online", "app":
				status.Idle, _ = time.ParseDuration(fields[1] + "m")
			case "invalid":
			}
			return false
		default:
			p.events <- frame
		}
		return true
	}

	err = p.SyncSend(cb, "track", username, "\r\n")

	return
}

// GetSettings retrieves the current settings.
//
// Returns:
//   - PMSetting: The current settings.
//   - error: An error if the operation fails.
func (p *Private) GetSettings() (setting models.PMSetting, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "settings":
			fields := strings.Split(data, ":")
			for i := 0; i < len(fields); i += 2 {
				switch fields[i] {
				case "disable_idle_time":
					setting.DisableIdleTime = fields[i+1] == "on"
				case "allow_anon":
					setting.AllowAnon = fields[i+1] == "on"
				case "email_offline_msg":
					setting.EmailOfflineMsg = fields[i+1] == "on"
				}
			}
			return false
		default:
			p.events <- frame
		}
		return true
	}

	err = p.SyncSend(cb, "settings", "\r\n")

	return
}

// SetSettings updates the settings with the provided values.
//
// Args:
//   - setting: The new settings to apply.
//
// Returns:
//   - error: An error if the operation fails.
func (p *Private) SetSettings(setting models.PMSetting) (err error) {
	onoff := func(state bool) string {
		if state {
			return "on"
		}
		return "off"
	}

	if err = p.Send("setsettings", "disable_idle_time", onoff(setting.DisableIdleTime), "\r\n"); err != nil {
		return
	}

	if err = p.Send("setsettings", "allow_anon", onoff(setting.AllowAnon), "\r\n"); err != nil {
		return
	}

	if err = p.Send("setsettings", "email_offline_msg", onoff(setting.EmailOfflineMsg), "\r\n"); err != nil {
		return
	}

	return
}

// GetFriendList retrieves the list of friends with their corresponding status.
//
// Returns:
//   - []UserStatus: The list of friends and their status.
//   - error: An error if the operation fails.
func (p *Private) GetFriendList() (friendlist []models.UserStatus, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "wl":
			fields := strings.Split(data, ":")
			var status models.UserStatus
			for i := 0; i < len(fields); i += 4 {
				status = models.UserStatus{User: &models.User{Name: fields[i]}}
				switch fields[i+2] {
				case "off":
					status.Info = "offline"
					status.Time, _ = utils.ParseTime(fields[i+1])
				case "on":
					status.Info = "online"
					status.Idle, _ = time.ParseDuration(fields[i+3] + "m")
				case "app":
					status.Info = "app"
					status.Idle, _ = time.ParseDuration(fields[i+3] + "m")
				}
				friendlist = append(friendlist, status)
			}
			return false
		default:
			p.events <- frame
		}
		return true
	}

	err = p.SyncSend(cb, "wl", "\r\n")

	return
}

// AddFriend adds a username to the friend list.
//
// Args:
//   - username: The username to add.
//
// Returns:
//   - UserStatus: The status of the added friend.
//   - error: An error if the operation fails.
func (p *Private) AddFriend(username string) (status models.UserStatus, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "wladd":
			fields := strings.SplitN(data, ":", 3)
			status.User = &models.User{Name: fields[0]}
			switch fields[1] {
			case "off":
				status.Info = "offline"
				status.Time, _ = utils.ParseTime(fields[2])
			case "on":
				status.Info = "online"
				status.Idle, _ = time.ParseDuration(fields[2] + "m")
			case "app":
				status.Info = "app"
				status.Idle, _ = time.ParseDuration(fields[2] + "m")
			}
			return false
		default:
			p.events <- frame
		}
		return true
	}

	err = p.SyncSend(cb, "wladd", username, "\r\n")

	return
}

// RemoveFriend removes the username from the friend list.
//
// Args:
//   - username: The username to remove.
//
// Returns:
//   - error: An error if the operation fails.
func (p *Private) RemoveFriend(username string) (err error) {
	cb := func(frame string) bool {
		head, _, _ := strings.Cut(frame, ":")
		switch head {
		case "wldelete":
			return false
		default:
			p.events <- frame
		}
		return true
	}

	err = p.SyncSend(cb, "wldelete", username, "\r\n")

	return
}

// GetBlocked retrieves the list of blocked users.
//
// Returns:
//   - users: A slice of blocked users.
//   - error: An error if the operation fails.
func (p *Private) GetBlocked() (users []*models.User, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "block_list":
			for _, username := range strings.Split(data, ":") {
				users = append(users, &models.User{Name: username})
			}
			return false
		default:
			p.events <- frame
		}
		return true
	}

	err = p.SyncSend(cb, "getblock", "\r\n")

	return
}

// Block blocks the user with the specified username.
//
// Args:
//   - username: The username to block.
//
// Returns:
//   - error: An error if the operation fails.
func (p *Private) Block(username string) (err error) {
	cb := func(frame string) bool {
		head, _, _ := strings.Cut(frame, ":")
		switch head {
		case "blocked":
			return false
		default:
			p.events <- frame
		}
		return true
	}

	err = p.SyncSend(cb, "block", username, "\r\n")

	return
}

// Unblock unblocks the user with the specified username.
//
// Args:
//   - username: The username to unblock.
//
// Returns:
//   - error: An error if the operation fails.
func (p *Private) Unblock(username string) (err error) {
	cb := func(frame string) bool {
		head, _, _ := strings.Cut(frame, ":")
		switch head {
		case "unblocked":
			return false
		default:
			p.events <- frame
		}
		return true
	}

	err = p.SyncSend(cb, "unblock", username, "\r\n")

	return
}

// ConnectUser opens a chat session with the username.
//
// Args:
//   - username: The username to connect with.
//
// Returns:
//   - UserStatus: The status of the connected user.
//   - error: An error if the operation fails.
func (p *Private) ConnectUser(username string) (status models.UserStatus, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "connect":
			fields := strings.SplitN(data, ":", 3)
			status.User = &models.User{Name: fields[0]}
			status.Info = fields[2]
			switch fields[2] {
			case "offline":
				status.Time, _ = utils.ParseTime(fields[2])
			case "online", "app":
				status.Idle, _ = time.ParseDuration(fields[2] + "m")
			}
			return false
		default:
			p.events <- frame
		}
		return true
	}

	err = p.SyncSend(cb, "connect", username, "\r\n")

	return
}

// DisconnectUser closes the chat session with the username.
//
// Args:
//   - username: The username to disconnect from.
//
// Returns:
//   - error: An error if the operation fails.
func (p *Private) DisconnectUser(username string) (err error) {
	return p.Send("disconnect", username, "\r\n")
}

// GetPresence retrieves the status of multiple usernames.
// If the user is offline, the corresponding [UserStatus.Time] is not accurate.
//
// Args:
//   - usernames: A slice of usernames to retrieve the status for.
//
// Returns:
//   - []UserStatus: A slice of `UserStatus` objects representing the status of each username.
//   - error: An error if the operation fails.
func (p *Private) GetPresence(usernames []string) (statuslist []models.UserStatus, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "presence":
			fields := strings.Split(data, ":")
			var status models.UserStatus
			var dur time.Duration
			for i := 0; i < len(fields); i += 3 {
				status = models.UserStatus{User: &models.User{Name: fields[i]}, Info: fields[i+2]}
				switch fields[i+2] {
				case "offline":
					dur, _ = time.ParseDuration(fields[i+1] + "m")
					status.Time = time.Now().Add(-dur)
				case "online", "app":
					status.Idle, _ = time.ParseDuration(fields[i+1] + "m")
				}
				statuslist = append(statuslist, status)
			}
			return false
		default:
			p.events <- frame
		}
		return true
	}

	err = p.SyncSend(cb, "statuslist", strings.Join(usernames, ";"), "\r\n")

	return
}

// ProfileRefresh notifies the server to refresh the profile.
//
// Returns:
//   - error: An error if the operation fails.
func (p *Private) ProfileRefresh() error {
	cb := func(frame string) bool {
		head, _, _ := strings.Cut(frame, ":")
		switch head {
		case "miu":
			p.events <- frame
			return false
		default:
			p.events <- frame
		}
		return true
	}

	return p.SyncSend(cb, "miu", "\r\n")
}

// WentIdle notifies the server that the user went idle.
//
// Returns:
//   - error: An error if the operation fails.
func (p *Private) WentIdle() error {
	p.IsIdle = true
	return p.Send("idle", "0", "\r\n")
}

// WentActive notifies the server that the user went active.
//
// Returns:
//   - error: An error if the operation fails.
func (p *Private) WentActive() (err error) {
	// Stops the previous timer.
	if !p.idleTimer.Stop() {
		// It appears that the callback has already been fired. Draining the channel.
		<-p.idleTimer.C
	}
	if p.IsIdle {
		err = p.Send("idle", "1", "\r\n")
		if err != nil {
			return
		}
		p.IsIdle = false
	}
	// Setup a new idle timer.
	p.idleTimer = time.AfterFunc(60*time.Second, func() { p.WentIdle() })

	return
}

// PropagateEvent propagates the WebSocket frame back to the listener goroutine.
//
// It is utilized when the [Private.SyncSend] callback receives unwanted frames.
//
// Args:
//   - frame: The WebSocket frame to propagate.
func (p *Private) PropagateEvent(frame string) {
	p.events <- frame
}

// GetContext returns the [context.Context] of the private chat.
//
// Returns:
//   - context.Context: The context of the private chat.
func (p *Private) GetContext() context.Context {
	return p.context
}

// wsOnError handles WebSocket errors that occur during communication.
//
// It attempts to reconnect if the connection is still active.
// If the reconnection fails, it disconnects and dispatches the [OnPrivateDisconnected] event.
//
// Args:
//   - err: The WebSocket error.
func (p *Private) wsOnError(err error) {
	close(p.events)
	close(p.takeOver)
	if p.Connected {
		if p.Reconnect() == nil {
			log.Debug().Str("Name", p.Name).Msg("Reconnected")

			event := &Event{
				Type:      OnPrivateReconnected,
				Private:   p,
				IsPrivate: true,
			}
			p.App.dispatchEvent(event)
			return
		}
	}
	p.Disconnect()
	log.Debug().Str("Name", p.Name).Msg("Disconnected")

	event := &Event{
		Type:      OnPrivateDisconnected,
		Private:   p,
		IsPrivate: true,
	}
	p.App.dispatchEvent(event)
}

// wsOnFrame handles incoming WebSocket frames.
//
// It parses the frame and dispatches the corresponding event.
//
// Args:
//   - frame: The WebSocket frame to handle.
func (p *Private) wsOnFrame(frame string) {
	defer func() {
		if err := recover(); err != nil {
			log.Error().Str("Name", p.Name).Str("Frame", frame).Msgf("Error: %s", err)

			event := &Event{
				Type:      OnError,
				Private:   p,
				IsPrivate: true,
				Error:     err,
			}
			p.App.dispatchEvent(event)
		}
	}()

	head, data, _ := strings.Cut(frame, ":")
	switch head {
	case "":
		// pong
	case "time":
		p.eventServerTime(data)
	case "OK":
		p.eventOK()
	case "seller_name":
		p.eventSellerName(data)
	case "kickingoff":
		p.eventKickedOff()
	case "msg":
		p.eventMessage(data)
	case "msgoff":
		p.eventOfflineMessage(data)
	case "wlonline":
		p.eventFriendOnline(data)
	case "wloffline":
		p.eventFriendOffline(data)
	case "wlapp":
		p.eventFriendOnlineApp(data)
	case "idleupdate":
		p.eventIdleUpdate(data)
	case "miu":
		p.eventUpdateUserProfile(data)
	case "show_fw", "toofast", "show_offline_limit":
		fallthrough
	case "track", "settings", "wl", "wladd", "wldelete":
		fallthrough
	case "block_list", "blocked", "unblocked":
		fallthrough
	case "connect", "presence":
		// This occurs when the [Private.SyncSendWithTimeout] fails to capture these events.
		// I'm leaving this here for debugging purposes.
		log.Debug().Str("Name", p.Name).Str("Frame", frame).Msg("Uncaptured")
	default:
		// I'm not familiar with the purpose of these events, but I discovered them in the HTML source code.
		// "reload_profile", Similar to "miu"?
		// "firstlogin", It is possible that the event is triggered after logging into a newly created or purchased account.
		// "lowversion", This event might be utilized by the message catcher that sends the "version" command to the server.
		// "status", Similar to "track"?
		log.Debug().Str("Name", p.Name).Str("Frame", frame).Msg("Unknown")
	}
}

// eventServerTime handles the server time event.
//
// It also saves the time differences between the client and the server (serverTime - clientTime).
func (p *Private) eventServerTime(data string) {
	p.LoginTime, _ = utils.ParseTime(data)
	p.TimeDiff = time.Since(p.LoginTime)
}

// eventOK handles the OK event.
//
// It dispatches the [OnPrivateConnected] event and sets up an idle timer.
func (p *Private) eventOK() {
	// Send the idle command 1 minute after the connection is established.
	p.idleTimer = time.AfterFunc(60*time.Second, func() { p.WentIdle() })

	// if p.App.Config.EnableBG {
	// 	go p.SetBackground(true)
	// }

	event := &Event{
		Type:      OnPrivateConnected,
		Private:   p,
		IsPrivate: true,
	}
	p.App.dispatchEvent(event)
}

// eventSellerName handles the seller name event with the session ID.
func (p *Private) eventSellerName(data string) {
	p.LoginName, p.SessionID, _ = strings.Cut(data, ":")
}

// eventKickedOff handles the kicked off event.
//
// This event is triggered when the same account initiates another PM session.
func (p *Private) eventKickedOff() {
	p.Disconnect()

	event := &Event{
		Type:      OnPrivateKickedOff,
		Private:   p,
		IsPrivate: true,
	}
	p.App.dispatchEvent(event)
}

// eventMessage handles the message event.
func (p *Private) eventMessage(data string) {
	message := ParsePrivateMessage(data, p)

	event := &Event{
		Type:      OnPrivateMessage,
		Private:   p,
		IsPrivate: true,
		Message:   message,
		User:      message.User,
	}
	p.App.dispatchEvent(event)
}

// eventOfflineMessage handles the offline message event.
func (p *Private) eventOfflineMessage(data string) {
	message := ParsePrivateMessage(data, p)

	event := &Event{
		Type:      OnPrivateOfflineMessage,
		Private:   p,
		IsPrivate: true,
		Message:   message,
		User:      message.User,
	}
	p.App.dispatchEvent(event)
}

// eventFriendOnline handles the friend online event.
func (p *Private) eventFriendOnline(data string) {
	username, _, _ := strings.Cut(data, ":")

	event := &Event{
		Type:      OnPrivateFriendOnline,
		Private:   p,
		IsPrivate: true,
		User:      &models.User{Name: username},
	}
	p.App.dispatchEvent(event)
}

// eventFriendOnlineApp handles the friend online (app) event.
func (p *Private) eventFriendOnlineApp(data string) {
	username, _, _ := strings.Cut(data, ":")

	event := &Event{
		Type:      OnPrivateFriendOnlineApp,
		Private:   p,
		IsPrivate: true,
		User:      &models.User{Name: username},
	}
	p.App.dispatchEvent(event)
}

// eventFriendOffline handles the friend offline event.
func (p *Private) eventFriendOffline(data string) {
	username, _, _ := strings.Cut(data, ":")

	event := &Event{
		Type:      OnPrivateFriendOffline,
		Private:   p,
		IsPrivate: true,
		User:      &models.User{Name: username},
	}
	p.App.dispatchEvent(event)
}

// eventIdleUpdate handles the friend idle update event.
func (p *Private) eventIdleUpdate(data string) {
	username, onoff, _ := strings.Cut(data, ":")

	var event *Event
	if onoff == "0" {
		event = &Event{
			Type:      OnPrivateFriendActive,
			Private:   p,
			IsPrivate: true,
			User:      &models.User{Name: username},
		}
	} else {
		event = &Event{
			Type:      OnPrivateFriendIdle,
			Private:   p,
			IsPrivate: true,
			User:      &models.User{Name: username},
		}
	}
	p.App.dispatchEvent(event)
}

// eventUpdateGroupInfo handles the update user profile event.
func (p *Private) eventUpdateUserProfile(data string) {
	event := &Event{
		Type:      OnUpdateUserProfile,
		Private:   p,
		IsPrivate: true,
		User:      &models.User{Name: data},
	}
	p.App.dispatchEvent(event)
}
