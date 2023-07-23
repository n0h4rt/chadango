package chadango

import (
	"strings"
)

// Handler is an interface that defines the methods for handling events.
type Handler interface {
	Check(*Event) bool       // Check checks if the event satisfies the conditions specified by the handler.
	Invoke(*Event, *Context) // Invoke handles the event using the specified callback and context.
}

// Callback is a function type that represents a callback function for handling events.
type Callback func(*Event, *Context)

// CommandHandler is a struct that implements the Handler interface for handling command in message events.
type CommandHandler struct {
	Callback Callback     // Callback is the function that will be invoked when a command event is triggered.
	Filter   Filter       // Filter is the filter that will be applied to the events before invoking the callback.
	Commands []string     // Commands is a list of command names that this handler will respond to.
	app      *Application // app is a reference to the application where this handler is registered.
}

// Check checks if the event is a command event that matches the prefix and command.
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
	if len(fields) == 0 || (len(fields) > 0 && !Contains(ch.Commands, fields[0])) {
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
func (ch *CommandHandler) Invoke(event *Event, context *Context) {
	ch.Callback(event, context)
}

// NewCommandHandler returns a new `CommandHandler`.
func NewCommandHandler(callback Callback, filter Filter, commands ...string) Handler {
	return &CommandHandler{
		Callback: callback,
		Filter:   filter,
		Commands: commands,
	}
}

// MessageHandler is a struct that implements the Handler interface for handling message events.
type MessageHandler struct {
	Callback Callback // Callback is the function that will be invoked when a message event is triggered.
	Filter   Filter   // Filter is the filter that will be applied to the events before invoking the callback.
}

// Check checks if the event is a message event.
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
func (mh *MessageHandler) Invoke(event *Event, context *Context) {
	mh.Callback(event, context)
}

// NewMessageHandler returns a new `MessageHandler`.
func NewMessageHandler(callback Callback, filter Filter) Handler {
	return &MessageHandler{
		Callback: callback,
		Filter:   filter,
	}
}

// TypeHandler is a struct that implements the Handler interface for handling events of a specific type.
type TypeHandler struct {
	Callback Callback  // Callback is the function that will be invoked when an event of the specified type is triggered.
	Filter   Filter    // Filter is the filter that will be applied to the events before invoking the callback.
	Type     EventType // Type is the type of event that this handler will respond to.
}

// Check checks if the event is of the specified type.
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
func (th *TypeHandler) Invoke(event *Event, context *Context) {
	th.Callback(event, context)
}

// NewTypeHandler returns a new `TypeHandler`.
func NewTypeHandler(callback Callback, filter Filter, eventType EventType) Handler {
	return &TypeHandler{
		Callback: callback,
		Filter:   filter,
		Type:     eventType,
	}
}
