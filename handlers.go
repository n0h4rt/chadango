package chadango

import (
	"strings"

	"github.com/n0h4rt/chadango/utils"
)

// Handler is an interface that defines the methods for handling events.
//
// It provides methods for checking if an event satisfies the conditions specified by the handler,
// and invoking a callback function to handle the event using the specified context.
type Handler interface {
	Check(*Event) bool       // Check checks if the event satisfies the conditions specified by the handler.
	Invoke(*Event, *Context) // Invoke handles the event using the specified callback and context.
}

// Callback is a function type that represents a callback function for handling events.
//
// It takes an [*Event] and a [*Context] as arguments and performs the desired actions based on the event.
type Callback func(*Event, *Context)

// CommandHandler is a struct that implements the [Handler] interface for handling command in message events.
//
// It filters events based on a command prefix and a list of commands, and invokes a callback function when a matching command is found.
// The handler also supports filtering events using a [Filter] object.
type CommandHandler struct {
	Callback Callback     // Callback is the function that will be invoked when a command event is triggered.
	Filter   Filter       // Filter is the filter that will be applied to the events before invoking the callback.
	Commands []string     // Commands is a list of command names that this handler will respond to.
	app      *Application // app is a reference to the application where this handler is registered.
}

// Check checks if the event is a command event that matches the prefix and command.
//
// Args:
//   - event: The event to check.
//
// Returns:
//   - bool: True if the event is a command event that matches the prefix and command, false otherwise.
func (ch *CommandHandler) Check(event *Event) bool {
	if event.Type != OnMessage && event.Type != OnPrivateMessage {
		return false
	}

	if event.Message.FromSelf {
		return false
	}

	text := strings.TrimLeft(event.Message.Text, "\r\n\t ")
	if !strings.HasPrefix(text, ch.app.Config.Prefix) {
		return false
	}

	text = strings.TrimLeft(text[len(ch.app.Config.Prefix):], "\r\n\t ")
	fields := strings.Fields(text)
	if len(fields) == 0 || (len(fields) > 0 && !utils.Contains(ch.Commands, fields[0])) {
		return false
	}

	ok := true
	if ch.Filter != nil {
		ok = ch.Filter.Check(event)
	}

	if ok {
		event.Command = fields[0]
		event.Arguments = fields[1:]
		event.Argument = strings.TrimLeft(text[len(fields[0]):], "\r\n\t ")
		event.WithArgument = len(fields) > 1
	}

	return ok
}

// Invoke executes the callback function for the command event.
//
// Args:
//   - event: The event to handle.
//   - context: The context for the event.
func (ch *CommandHandler) Invoke(event *Event, context *Context) {
	ch.Callback(event, context)
}

// NewCommandHandler returns a new [CommandHandler].
//
// Args:
//   - callback: The callback function to invoke when a command event is triggered.
//   - filter: The filter to apply to the events before invoking the callback.
//   - commands: A list of command names that this handler will respond to.
//
// Returns:
//   - Handler: A new [CommandHandler] instance.
func NewCommandHandler(callback Callback, filter Filter, commands ...string) Handler {
	return &CommandHandler{
		Callback: callback,
		Filter:   filter,
		Commands: commands,
	}
}

// MessageHandler is a struct that implements the [Handler] interface for handling message events.
//
// It filters events based on a [Filter] object and invokes a callback function when a matching message event is found.
type MessageHandler struct {
	Callback Callback // Callback is the function that will be invoked when a message event is triggered.
	Filter   Filter   // Filter is the filter that will be applied to the events before invoking the callback.
}

// Check checks if the event is a message event.
//
// Args:
//   - event: The event to check.
//
// Returns:
//   - bool: True if the event is a message event, false otherwise.
func (mh *MessageHandler) Check(event *Event) bool {
	if event.Type != OnMessage && event.Type != OnPrivateMessage {
		return false
	}

	if event.Message.FromSelf {
		return false
	}

	ok := true
	if mh.Filter != nil {
		ok = mh.Filter.Check(event)
	}

	return ok
}

// Invoke executes the callback function for the message event.
//
// Args:
//   - event: The event to handle.
//   - context: The context for the event.
func (mh *MessageHandler) Invoke(event *Event, context *Context) {
	mh.Callback(event, context)
}

// NewMessageHandler returns a new [MessageHandler].
//
// Args:
//   - callback: The callback function to invoke when a message event is triggered.
//   - filter: The filter to apply to the events before invoking the callback.
//
// Returns:
//   - Handler: A new [MessageHandler] instance.
func NewMessageHandler(callback Callback, filter Filter) Handler {
	return &MessageHandler{
		Callback: callback,
		Filter:   filter,
	}
}

// TypeHandler is a struct that implements the [Handler] interface for handling events of a specific type.
//
// It filters events based on a specific event type and a [Filter] object, and invokes a callback function when a matching event is found.
type TypeHandler struct {
	Callback Callback  // Callback is the function that will be invoked when an event of the specified type is triggered.
	Filter   Filter    // Filter is the filter that will be applied to the events before invoking the callback.
	Type     EventType // Type is the type of event that this handler will respond to.
}

// Check checks if the event is of the specified type.
//
// Args:
//   - event: The event to check.
//
// Returns:
//   - bool: True if the event is of the specified type, false otherwise.
func (th *TypeHandler) Check(event *Event) bool {
	if th.Type&event.Type == 0 {
		return false
	}

	ok := true
	if th.Filter != nil {
		ok = th.Filter.Check(event)
	}

	return ok
}

// Invoke executes the callback function for the event of the specified type.
//
// Args:
//   - event: The event to handle.
//   - context: The context for the event.
func (th *TypeHandler) Invoke(event *Event, context *Context) {
	th.Callback(event, context)
}

// NewTypeHandler returns a new [TypeHandler].
//
// Args:
//   - callback: The callback function to invoke when an event of the specified type is triggered.
//   - filter: The filter to apply to the events before invoking the callback.
//   - eventType: The type of event that this handler will respond to.
//
// Returns:
//   - Handler: A new [TypeHandler] instance.
func NewTypeHandler(callback Callback, filter Filter, eventType EventType) Handler {
	return &TypeHandler{
		Callback: callback,
		Filter:   filter,
		Type:     eventType,
	}
}
