package chadango

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

// Application represents the main application.
type Application struct {
	Config        *Config                 // Config holds the configuration for the application.
	Persistence   *Persistence            // Persistence manages data persistence for the application.
	API           *API                    // API provides access to various Chatango APIs used by the application.
	Private       Private                 // Private represents the private chat functionality of the application.
	Groups        SyncMap[string, *Group] // Groups stores the groups the application is connected to.
	handlers      []Handler               // handlers contains the registered event handlers for the application.
	interruptChan chan os.Signal          // interruptChan receives interrupt signals to gracefully stop the application.
	context       context.Context         // Context for running the application.
	cancelCtx     context.CancelFunc      // Function for stopping the application.
	initialized   bool                    // initialized indicates whether the application has been initialized.
	mu            sync.Mutex              // mu is a mutex used for locking access to critical sections.
}

// AddHandler adds a new handler to the application.
// It returns the `*Application` to allow for nesting.
func (app *Application) AddHandler(handler Handler) *Application {
	app.mu.Lock()
	defer app.mu.Unlock()

	if ch, ok := handler.(*CommandHandler); ok {
		ch.app = app
	}
	app.handlers = append(app.handlers, handler)
	return app
}

// RemoveHandler removes a handler from the application.
func (app *Application) RemoveHandler(handler Handler) {
	app.mu.Lock()
	defer app.mu.Unlock()

	// Find and remove the handler from the collection
	for i, h := range app.handlers {
		if h == handler {
			app.handlers = append(app.handlers[:i], app.handlers[i+1:]...)
			break
		}
	}
}

// Dispatch dispatches an event to the appropriate handler.
func (app *Application) Dispatch(event *Event) {
	var context *Context

	for _, handler := range app.handlers {
		if handler.Check(event) {
			if context == nil {
				context = &Context{
					App:     app,
					BotData: app.Persistence.GetBotData(),
				}
				if event.IsPrivate && event.User != nil && !event.User.IsAnon {
					context.ChatData = app.Persistence.GetChatData(strings.ToLower(event.User.Name))
				} else {
					context.ChatData = app.Persistence.GetChatData(event.Group.Name)
				}
			}
			handler.Invoke(event, context)
			break
		}
	}
}

// Initialize initializes the application.
func (app *Application) Initialize() {
	if app.Persistence == nil {
		app.Persistence = new(Persistence)
	}
	app.Persistence.Initialize()
	app.API = &API{
		Username: app.Config.Username,
		Password: app.Config.Password,
	}
	app.API.Initialize()
	app.Groups = NewSyncMap[string, *Group]()
	app.checkConfig()
	app.initialized = true
}

// checkConfig checks certain configurations and assigns default values if they are left unset.
func (app *Application) checkConfig() {
	if app.Config.AnonName == "" {
		app.Config.AnonName = "anon0001"
	}
	if app.Config.NameColor == "" {
		app.Config.NameColor = DEFAULT_COLOR
	}
	if app.Config.TextColor == "" {
		app.Config.TextColor = DEFAULT_COLOR
	}
	if app.Config.TextFont == "" {
		app.Config.TextFont = DEFAULT_TEXT_FONT
	}
	if app.Config.TextSize == 0 {
		app.Config.TextSize = DEFAULT_TEXT_SIZE
	}
}

// Start starts the application.
func (app *Application) Start(ctx context.Context) {
	if !app.initialized {
		panic("the application is not initialized")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	app.context, app.cancelCtx = context.WithCancel(ctx)
	for _, groupName := range app.Config.Groups {
		go app.JoinGroup(groupName)
	}
	if app.Config.EnablePM {
		go app.ConnectPM()
	}
	app.interruptChan = make(chan os.Signal, 1)
	signal.Notify(app.interruptChan, os.Interrupt, syscall.SIGTERM)
}

// Park waits for the application to stop or receive an interrupt signal.
func (app *Application) Park() {
	select {
	case <-app.context.Done():
	case <-app.interruptChan:
		app.Stop()
	}
}

// Stop stops the application.
func (app *Application) Stop() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		app.DisconnectPM()
		wg.Done()
	}()
	cb := func(_ string, group *Group) bool {
		wg.Add(1)
		go func() {
			group.Disconnect()
			wg.Done()
		}()
		return false
	}
	app.Groups.Range(cb)
	wg.Wait()
	app.Persistence.Close()
	app.cancelCtx()
}

// JoinGroup joins a group in the application.
func (app *Application) JoinGroup(groupName string) error {
	groupName = strings.ToLower(groupName)
	if _, ok := app.Groups.Get(groupName); ok {
		return ErrAlreadyConnected
	}
	if isGroup, err := app.API.IsGroup(groupName); err != nil || !isGroup {
		return ErrNotAGroup
	}
	group := Group{
		App:       app,
		Name:      groupName,
		WsUrl:     GetServer(groupName),
		AnonName:  app.Config.AnonName,
		NameColor: app.Config.NameColor,
		TextColor: app.Config.TextColor,
		TextFont:  app.Config.TextFont,
		TextSize:  app.Config.TextSize,
		SessionID: app.Config.SessionID,
		LoggedIn:  app.Config.Password != "",
	}
	if err := group.Connect(app.context); err != nil {
		return err
	}
	app.Groups.Set(groupName, &group)
	return nil
}

// LeaveGroup leaves a group in the application.
func (app *Application) LeaveGroup(groupName string) error {
	groupName = strings.ToLower(groupName)
	if group, ok := app.Groups.Get(groupName); ok {
		// app.Groups.Del(groupName) // Group deletion is handled by the `Group.wsOnError`.
		group.Disconnect()
		return nil
	}
	return ErrNotConnected
}

// ConnectPM connects to private messages.
func (app *Application) ConnectPM() error {
	app.Private.App = app
	app.Private.Name = "Private"
	app.Private.WsUrl = PM_SERVER
	app.Private.NameColor = app.Config.NameColor
	app.Private.TextColor = app.Config.TextColor
	app.Private.TextFont = app.Config.TextFont
	app.Private.TextSize = app.Config.TextSize
	app.Private.SessionID = app.Config.SessionID
	return app.Private.Connect(app.context)
}

// DisconnectPM disconnects from private messages.
func (app *Application) DisconnectPM() {
	app.Private.Disconnect()
}

func New(config *Config) *Application {
	return &Application{Config: config}
}
