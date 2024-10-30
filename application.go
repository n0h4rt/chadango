package chadango

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/n0h4rt/chadango/models"
	"github.com/n0h4rt/chadango/utils"
	"github.com/rs/zerolog/log"
)

// Application represents the main application.
//
// It provides methods for managing the application's lifecycle, including initialization, starting, stopping,
// connecting to groups and private messages, and handling events and errors.
// The application also manages data persistence and provides access to the public and private APIs.
type Application struct {
	Config        *Config                 // Config holds the configuration for the pplication.
	persistence   Persistence             // Persistence manages data persistence for the application.
	Private       Private                 // Private represents the private chat functionality of the application.
	Groups        SyncMap[string, *Group] // Groups stores the groups the application is connected to.
	eventHandlers []Handler               // eventHandlers contains the registered event handlers for the application.
	errorHandlers []Handler               // errorHandlers contains the registered error handlers for the application.
	context       context.Context         // Context for running the application.
	cancelCtx     context.CancelFunc      // Function for stopping the application.
	initialized   bool                    // initialized indicates whether the application has been initialized.
}

// AddHandler adds a new handler to the application.
//
// Args:
//   - handler: The handler to add to the application.
//
// Returns:
//   - *Application: The application instance for method chaining.
func (app *Application) AddHandler(handler Handler) *Application {
	if ch, ok := handler.(*CommandHandler); ok {
		ch.app = app
	}

	app.eventHandlers = append(app.eventHandlers, handler)

	return app
}

// RemoveHandler removes a handler from the application.
//
// Args:
//   - handler: The handler to remove from the application.
//
// Returns:
//   - *Application: The application instance for method chaining.
func (app *Application) RemoveHandler(handler Handler) *Application {
	app.eventHandlers = utils.Remove(app.eventHandlers, handler)

	return app
}

// AddErrorHandler adds a new error handler to the application.
//
// Args:
//   - handler: The error handler to add to the application.
//
// Returns:
//   - *Application: The application instance for method chaining.
func (app *Application) AddErrorHandler(handler Handler) *Application {
	if ch, ok := handler.(*CommandHandler); ok {
		ch.app = app
	}

	app.errorHandlers = append(app.errorHandlers, handler)

	return app
}

// RemoveErrorHandler removes an error handler from the application.
//
// Args:
//   - handler: The error handler to remove from the application.
//
// Returns:
//   - *Application: The application instance for method chaining.
func (app *Application) RemoveErrorHandler(handler Handler) *Application {
	app.errorHandlers = utils.Remove(app.errorHandlers, handler)

	return app
}

// UsePersistence enables the persistence layer for the application.
//
// Args:
//   - persistence: The persistence layer to use for the application.
//
// Returns:
//   - *Application: The application instance for method chaining.
func (app *Application) UsePersistence(persistence Persistence) *Application {
	app.persistence = persistence

	return app
}

// dispatchEvent dispatches an event to the appropriate handler.
//
// Args:
//   - event: The event to dispatch.
func (app *Application) dispatchEvent(event *Event) {
	var context *Context

	for _, handler := range app.eventHandlers {
		if handler.Check(event) {
			if context == nil {
				context = &Context{
					App:     app,
					BotData: app.persistence.GetBotData(),
				}

				if event.IsPrivate && event.User != nil && !event.User.IsAnon {
					context.ChatData = app.persistence.GetChatData(strings.ToLower(event.User.Name))
				} else if event.Group != nil {
					context.ChatData = app.persistence.GetChatData(event.Group.Name)
				}
			}

			func() {
				defer func() {
					if err := recover(); err != nil {
						event.Error = err

						app.dispatchError(event)
					}
				}()

				handler.Invoke(event, context)
			}()
		}
	}
}

// dispatchError dispatches an error event to the error handlers.
//
// Args:
//   - event: The event to dispatch.
func (app *Application) dispatchError(event *Event) {
	var context *Context

	for _, handler := range app.errorHandlers {
		if handler.Check(event) {
			if context == nil {
				context = &Context{
					App:     app,
					BotData: app.persistence.GetBotData(),
				}

				if event.IsPrivate && event.User != nil && !event.User.IsAnon {
					context.ChatData = app.persistence.GetChatData(strings.ToLower(event.User.Name))
				} else if event.Group != nil {
					context.ChatData = app.persistence.GetChatData(event.Group.Name)
				}
			}

			func() {
				defer func() {
					if err := recover(); err != nil {
						log.Error().
							Str("Event", event.Type.String()).
							AnErr("Origin", event.Error.(error)).
							AnErr("Current", err.(error)).
							Msg("Another error occured during handling an error.")
					}
				}()

				handler.Invoke(event, context)
			}()
		}
	}
}

// Initialize initializes the application.
//
// Returns:
//   - *Application: The application instance for method chaining.
func (app *Application) Initialize() *Application {
	if app.persistence == nil {
		// Using the GobPersistence without Filename as a dummy.
		app.persistence = new(GobPersistence)
	}
	app.persistence.Initialize()

	app.Groups = NewSyncMap[string, *Group]()
	app.checkConfig()
	app.initialized = true

	return app
}

// checkConfig checks certain configurations and assigns default values if they are left unset.
func (app *Application) checkConfig() {
	if app.Config.AnonName == "" {
		app.Config.AnonName = "anon0001"
	}
	if app.Config.NameColor == "" {
		app.Config.NameColor = models.DEFAULT_COLOR
	}
	if app.Config.TextColor == "" {
		app.Config.TextColor = models.DEFAULT_COLOR
	}
	if app.Config.TextFont == "" {
		app.Config.TextFont = models.DEFAULT_TEXT_FONT
	}
	if app.Config.TextSize == 0 {
		app.Config.TextSize = models.DEFAULT_TEXT_SIZE
	}
}

// Start starts the application.
//
// Args:
//   - ctx: The context for running the application.
//
// Returns:
//   - *Application: The application instance for method chaining.
func (app *Application) Start(ctx context.Context) *Application {
	if !app.initialized {
		panic("the application is not initialized")
	}

	if ctx == nil {
		ctx = context.Background()
	}
	app.context, app.cancelCtx = context.WithCancel(ctx)

	initAPI(app.Config.Username, app.Config.Password, ctx)

	for _, groupName := range app.Config.Groups {
		go app.JoinGroup(groupName)
	}
	if app.Config.EnablePM {
		go app.ConnectPM()
	}

	app.persistence.Runner(app.context)

	app.dispatchEvent(&Event{Type: OnStart})

	return app
}

// Park waits for the application to stop or receive an interrupt signal.
func (app *Application) Park() {
	intCh := make(chan os.Signal, 1)
	signal.Notify(intCh, os.Interrupt, syscall.SIGTERM)

	select {
	case <-app.context.Done():
	case <-intCh:
		app.Stop()
	}
}

// Stop stops the application.
func (app *Application) Stop() {
	app.dispatchEvent(&Event{Type: OnStop})

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
		return true
	}
	app.Groups.Range(cb)

	wg.Wait()
	app.persistence.Close()
	app.cancelCtx()
}

// JoinGroup joins a group in the application.
//
// Args:
//   - groupName: The name of the group to join.
//
// Returns:
//   - error: An error if the group cannot be joined.
func (app *Application) JoinGroup(groupName string) error {
	groupName = strings.ToLower(groupName)
	if _, ok := app.Groups.Get(groupName); ok {
		return ErrAlreadyConnected
	}

	if isGroup, err := publicAPI.IsGroup(groupName); err != nil || !isGroup {
		return ErrNotAGroup
	}

	group := Group{
		App:       app,
		Name:      groupName,
		WsUrl:     utils.GetServer(groupName),
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
//
// Args:
//   - groupName: The name of the group to leave.
//
// Returns:
//   - error: An error if the group cannot be left.
func (app *Application) LeaveGroup(groupName string) error {
	groupName = strings.ToLower(groupName)
	if group, ok := app.Groups.Get(groupName); ok {
		// app.Groups.Del(groupName) // Group deletion is handled by the [Group.wsOnError].
		group.Disconnect()
		return nil
	}

	return ErrNotConnected
}

// ConnectPM connects to private messages.
//
// Returns:
//   - error: An error if the connection to private messages fails.
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

// GetContext returns the [context.Context] of the application.
//
// Returns:
//   - context.Context: The context of the application.
func (app *Application) GetContext() context.Context {
	return app.context
}

// PrivateAPI returns the [PrivateAPI] used in the application.
//
// Returns:
//   - *PrivateAPI: The private API used in the application.
func (app *Application) PrivateAPI() *PrivateAPI {
	return privateAPI
}

// PublicAPI returns the [PublicAPI] used in the application.
//
// Returns:
//   - *PublicAPI: The public API used in the application.
func (app *Application) PublicAPI() *PublicAPI {
	return publicAPI
}

// New creates a new instance of the [Application] with the provided configuration.
//
// Args:
//   - config: The configuration for the application.
//
// Returns:
//   - *Application: A new instance of the [Application].
func New(config *Config) *Application {
	return &Application{
		eventHandlers: []Handler{},
		errorHandlers: []Handler{},
		Config:        config,
	}
}
