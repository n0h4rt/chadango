package chadango

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// Private represents a private message with various properties and state.
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
// It returns an error if the connection cannot be established.
func (p *Private) Connect(ctx context.Context) (err error) {
	if p.Connected {
		return ErrAlreadyConnected
	}

	p.context, p.cancelCtx = context.WithCancel(ctx)

	err = p.connect()
	if err != nil {
		p.cancelCtx()
	}

	p.Connected = true
	return
}

func (p *Private) connect() (err error) {
	var ok bool
	if p.token, ok = p.App.API.GetCookie("auth.chatango.com"); !ok {
		return ErrLoginFailed
	}

	log.Debug().Str("Name", p.Name).Msg("Connecting")

	defer func() {
		if err != nil {
			log.Debug().Str("Name", p.Name).Err(err).Msg("Connect failed")
			p.ws.Close()
		}
	}()

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

	// Random responses may be received; expecting either OK or DENIED within 5 loops.
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
		}
		p.events <- frame
	}
	return ErrBadLogin

OK:
	p.ws.Sustain(p.context)
	go p.listen()
	log.Debug().Str("Name", p.Name).Msg("Connected")
	return
}

// listen listens for incoming messages and events on the WebSocket connection.
func (p *Private) listen() {
	var frame string
	var release context.Context
	for {
		select {
		case <-p.context.Done():
			return
		case frame = <-p.events:
			if frame == EndFrame {
				return
			}
			go p.wsOnFrame(frame)
		case frame = <-p.ws.Events:
			if frame == EndFrame {
				return
			}
			go p.wsOnFrame(strings.TrimRight(frame, "\r\n\x00"))
		case release = <-p.takeOver:
		inner:
			for {
				select {
				case <-p.context.Done():
					return
				case <-release.Done():
					break inner
				case frame = <-p.events:
					if frame == EndFrame {
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
func (p *Private) Reconnect() (err error) {
	p.ws.Close()

	// Reinitialize the API.
	if err = p.App.API.Initialize(); err != nil {
		log.Error().Str("Name", "API").Err(err).Msg("Could not initialize the API")
		return ErrLoginFailed
	}

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

// Send will join the `args` with a ":" separator and then send it to the server asynchronously.
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

// SyncSendWithTimeout will send the `args` and wait until receiving the correct reply or until timeout.
// First, a `p.takeOver` request will be made and it will wait until the listener goroutine catches it.
// Then, the `args` will be sent to the server.
// Each time a frame is received, the `callback` function is invoked and passed the frame.
// The `callback` should return `true` if a correct frame is acquired, and `false` otherwise.
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
	for {
		select {
		case <-ctx.Done():
			return ErrTimeout
		case frame = <-p.ws.Events:
			if frame == EndFrame {
				p.events <- frame
				return ErrConnectionClosed
			}
			if callback(strings.TrimRight(frame, "\r\n\x00")) {
				return
			}
		}
	}
}

// SyncSend will send the `args` and wait until receiving the correct reply or until timeout (default to 5 seconds).
// For more information, refer to the documentation of `p.SyncSendWithTimeout`.
func (p *Private) SyncSend(cb func(string) bool, text ...string) error {
	return p.SyncSendWithTimeout(cb, SYNC_SEND_TIMEOUT, text...)
}

// SendMessage sends a private message with the specified text to the username.
func (p *Private) SendMessage(username, text string) (err error) {
	cb := func(frame string) bool {
		head, _, _ := strings.Cut(frame, ":")
		switch head {
		case "show_fw":
			err = ErrFloodWarning
			return true
		case "toofast":
			err = ErrFloodBanned
			return true
		case "show_offline_limit":
			// The maximum number of unread messages is 51.
			// show_offline_limit:nekonyan
			err = ErrOfflineLimit
			return true
		default:
			p.events <- frame
		}
		return false
	}
	username = strings.ToLower(username)
	text = fmt.Sprintf(`<n%s/><m v="1"><g x%02ds%s="%s">%s</g></m>`, p.NameColor, p.TextSize, p.TextColor, p.TextFont, text)
	// The "msg" command does not produce a response, so we wait for 0.5 seconds to handle any potential errors that may occur.
	// The potential errors include "show_fw", "toofast", and "show_offline_limit".
	if err = p.SyncSendWithTimeout(cb, 500*time.Millisecond, "msg", username, text, "\r\n"); err == ErrTimeout {
		return nil
	}

	// Notifies the PM server that the client has just been active.
	p.WentActive()
	return
}

// Track retrieves the online status of the username.
func (p *Private) Track(username string) (status UserStatus, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "track":
			fields := strings.SplitN(data, ":", 3)
			status = UserStatus{
				User: &User{Name: fields[0]},
				Info: fields[2],
			}
			switch fields[2] {
			case "offline":
				status.Time, _ = ParseTime(fields[1])
			case "online", "app":
				status.Idle, _ = time.ParseDuration(fields[1] + "m")
			case "invalid":
				err = ErrInvalidUsername
			}
			return true
		default:
			p.events <- frame
		}
		return false
	}
	err = p.SyncSend(cb, "track", username, "\r\n")
	return
}

// GetSettings retrieves the current settings.
func (p *Private) GetSettings() (setting PrivateSetting, err error) {
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
			return true
		default:
			p.events <- frame
		}
		return false
	}
	err = p.SyncSend(cb, "settings", "\r\n")
	return
}

// SetSettings updates the settings with the provided values.
func (p *Private) SetSettings(setting PrivateSetting) (err error) {
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
func (p *Private) GetFriendList() (friendlist []UserStatus, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "wl":
			fields := strings.Split(data, ":")
			var status UserStatus
			for i := 0; i < len(fields); i += 4 {
				status = UserStatus{User: &User{Name: fields[i]}}
				switch fields[i+2] {
				case "off":
					status.Info = "offline"
					status.Time, _ = ParseTime(fields[i+1])
				case "on", "app":
					status.Info = "online"
					status.Idle, _ = time.ParseDuration(fields[i+3] + "m")
				}
				friendlist = append(friendlist, status)
			}
			return true
		default:
			p.events <- frame
		}
		return false
	}
	err = p.SyncSend(cb, "wl", "\r\n")
	return
}

// AddFriend adds a username to the friend list.
func (p *Private) AddFriend(username string) (status UserStatus, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "wladd":
			fields := strings.SplitN(data, ":", 3)
			status = UserStatus{User: &User{Name: fields[0]}}
			switch fields[1] {
			case "off":
				status.Info = "offline"
				status.Time, _ = ParseTime(fields[2])
			case "on", "app":
				status.Info = "online"
				status.Idle, _ = time.ParseDuration(fields[2] + "m")
			}
			return true
		default:
			p.events <- frame
		}
		return false
	}
	err = p.SyncSend(cb, "wladd", username, "\r\n")
	return
}

// RemoveFriend removes the username from the friend list.
func (p *Private) RemoveFriend(username string) (err error) {
	cb := func(frame string) bool {
		head, _, _ := strings.Cut(frame, ":")
		switch head {
		case "wldelete":
			return true
		default:
			p.events <- frame
		}
		return false
	}
	err = p.SyncSend(cb, "wldelete", username, "\r\n")
	return
}

// GetBlocked retrieves the list of blocked users.
func (p *Private) GetBlocked() (users []*User, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "block_list":
			for _, username := range strings.Split(data, ":") {
				users = append(users, &User{Name: username})
			}
			return true
		default:
			p.events <- frame
		}
		return false
	}
	err = p.SyncSend(cb, "getblock", "\r\n")
	return
}

// Block blocks the user with the specified username.
func (p *Private) Block(username string) (err error) {
	cb := func(frame string) bool {
		head, _, _ := strings.Cut(frame, ":")
		switch head {
		case "blocked":
			return true
		default:
			p.events <- frame
		}
		return false
	}
	err = p.SyncSend(cb, "block", username, "\r\n")
	return
}

// Unblock unblocks the user with the specified username.
func (p *Private) Unblock(username string) (err error) {
	cb := func(frame string) bool {
		head, _, _ := strings.Cut(frame, ":")
		switch head {
		case "unblocked":
			return true
		default:
			p.events <- frame
		}
		return false
	}
	err = p.SyncSend(cb, "unblock", username, "\r\n")
	return
}

// ConnectUser opens a chat session with the username.
func (p *Private) ConnectUser(username string) (status UserStatus, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "connect":
			fields := strings.SplitN(data, ":", 3)
			status = UserStatus{User: &User{Name: fields[0]}, Info: fields[2]}
			switch fields[2] {
			case "offline":
				status.Time, _ = ParseTime(fields[2])
			case "online", "app":
				status.Idle, _ = time.ParseDuration(fields[2] + "m")
			}
			return true
		default:
			p.events <- frame
		}
		return false
	}
	err = p.SyncSend(cb, "connect", username, "\r\n")
	return
}

// DisconnectUser closes the chat session with the username.
func (p *Private) DisconnectUser(username string) (err error) {
	return p.Send("disconnect", username, "\r\n")
}

// GetPresence retrieves the status of multiple usernames.
// If the user is offline, the corresponding `UserStatus.Time` is not accurate.
func (p *Private) GetPresence(usernames []string) (statuslist []UserStatus, err error) {
	cb := func(frame string) bool {
		head, data, _ := strings.Cut(frame, ":")
		switch head {
		case "presence":
			fields := strings.Split(data, ":")
			var status UserStatus
			var dur time.Duration
			for i := 0; i < len(fields); i += 3 {
				status = UserStatus{User: &User{Name: fields[i]}, Info: fields[i+2]}
				switch fields[i+2] {
				case "offline":
					dur, _ = time.ParseDuration(fields[i+1] + "m")
					status.Time = time.Now().Add(-dur)
				case "online", "app":
					status.Idle, _ = time.ParseDuration(fields[i+1] + "m")
				}
				statuslist = append(statuslist, status)
			}
			return true
		default:
			p.events <- frame
		}
		return false
	}
	err = p.SyncSend(cb, "statuslist", strings.Join(usernames, ";"), "\r\n")
	return
}

// ProfileRefresh notifies the server to refresh the profile.
func (p *Private) ProfileRefresh() error {
	cb := func(frame string) bool {
		head, _, _ := strings.Cut(frame, ":")
		switch head {
		case "miu":
			p.events <- frame
			return true
		default:
			p.events <- frame
		}
		return false
	}
	return p.SyncSend(cb, "miu", "\r\n")
}

// WentIdle notifies the server that the user went idle.
func (p *Private) WentIdle() {
	p.IsIdle = true
	p.Send("idle", "0", "\r\n")
}

// WentActive notifies the server that the user went active.
func (p *Private) WentActive() {
	// Stops the previous timer.
	if !p.idleTimer.Stop() {
		// It appears that the callback has already been fired. Draining the channel.
		<-p.idleTimer.C
	}
	if p.IsIdle {
		p.Send("idle", "1", "\r\n")
		p.IsIdle = false
	}
	// Setup a new idle timer.
	p.idleTimer = time.AfterFunc(60*time.Second, p.WentIdle)
}

// PropagateEvent propagates the WebSocket frame back to the listener goroutine.
// It is utilized when the `p.SyncSend` callback receives unwanted frames.
func (p *Private) PropagateEvent(frame string) {
	p.events <- frame
}

// GetContext returns the context of the private chat.
func (p *Private) GetContext() context.Context {
	return p.context
}

// wsOnError handles WebSocket errors that occur during communication.
func (p *Private) wsOnError(e error) {
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
			p.App.Dispatch(event)
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
	p.App.Dispatch(event)
}

// wsOnFrame handles incoming WebSocket frames.
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
			p.App.Dispatch(event)
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
		// This occurs when the `p.SyncSendWithTimeout` fails to capture these events.
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
// It also saves the time differences between the client and the server (serverTime - clientTime).
func (p *Private) eventServerTime(data string) {
	p.LoginTime, _ = ParseTime(data)
	p.TimeDiff = time.Since(p.LoginTime)
}

// eventOK handles the OK event.
func (p *Private) eventOK() {
	// Send the idle command 1 minute after the connection is established.
	p.idleTimer = time.AfterFunc(60*time.Second, p.WentIdle)

	// if p.App.Config.EnableBG {
	// 	go p.SetBackground(true)
	// }

	event := &Event{
		Type:      OnPrivateConnected,
		Private:   p,
		IsPrivate: true,
	}
	p.App.Dispatch(event)
}

// eventSellerName handles the seller name event with the session ID.
func (p *Private) eventSellerName(data string) {
	p.LoginName, p.SessionID, _ = strings.Cut(data, ":")
}

// eventKickedOff handles the kicked off event.
// This event is triggered when the same account initiates another PM session.
func (p *Private) eventKickedOff() {
	p.Disconnect()
	event := &Event{
		Type:      OnPrivateKickedOff,
		Private:   p,
		IsPrivate: true,
	}
	p.App.Dispatch(event)
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
	p.App.Dispatch(event)
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
	p.App.Dispatch(event)
}

// eventFriendOnline handles the friend online event.
func (p *Private) eventFriendOnline(data string) {
	username, _, _ := strings.Cut(data, ":")
	event := &Event{
		Type:      OnPrivateFriendOnline,
		Private:   p,
		IsPrivate: true,
		User:      &User{Name: username},
	}
	p.App.Dispatch(event)
}

// eventFriendOnlineApp handles the friend online (app) event.
func (p *Private) eventFriendOnlineApp(data string) {
	username, _, _ := strings.Cut(data, ":")
	event := &Event{
		Type:      OnPrivateFriendOnlineApp,
		Private:   p,
		IsPrivate: true,
		User:      &User{Name: username},
	}
	p.App.Dispatch(event)
}

// eventFriendOffline handles the friend offline event.
func (p *Private) eventFriendOffline(data string) {
	username, _, _ := strings.Cut(data, ":")
	event := &Event{
		Type:      OnPrivateFriendOffline,
		Private:   p,
		IsPrivate: true,
		User:      &User{Name: username},
	}
	p.App.Dispatch(event)
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
			User:      &User{Name: username},
		}
	} else {
		event = &Event{
			Type:      OnPrivateFriendIdle,
			Private:   p,
			IsPrivate: true,
			User:      &User{Name: username},
		}
	}
	p.App.Dispatch(event)
}

// eventUpdateGroupInfo handles the update user profile event.
func (p *Private) eventUpdateUserProfile(data string) {
	event := &Event{
		Type:      OnUpdateUserProfile,
		Private:   p,
		IsPrivate: true,
		User:      &User{Name: data},
	}
	p.App.Dispatch(event)
}
