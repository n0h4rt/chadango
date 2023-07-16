package chadango

import (
	"context"
	"time"

	"golang.org/x/net/websocket"
)

// WebSocket represents a WebSocket connection.
// It implements `golang.org/x/net/websocket` under the hood and wraps it into a channel,
// allowing it to be select-able along with other channels.
type WebSocket struct {
	Connected bool
	Events    chan string
	OnError   func(error)

	url        string
	client     *websocket.Conn
	runCtx     context.Context
	cancelFunc context.CancelFunc
}

// Connect establishes a WebSocket connection to the specified URL.
func (w *WebSocket) Connect(url string) (err error) {
	if w.Connected {
		return
	}

	w.url = url
	w.client, err = websocket.Dial(url, "", WEBSOCKET_ORIGIN)
	if err != nil {
		return err
	}

	w.Connected = true
	w.Events = make(chan string, EVENT_BUFFER_SIZE)
	return
}

// Close closes the WebSocket connection.
func (w *WebSocket) Close() {
	if w.Connected {
		w.Connected = false
		if w.cancelFunc != nil {
			w.cancelFunc()
		}
		w.client.Close()
	}
}

// Sustain starts pumping events and keeps the WebSocket connection alive.
func (w *WebSocket) Sustain(ctx context.Context) {
	w.runCtx, w.cancelFunc = context.WithCancel(ctx)
	go w.pumpEvent()
	go w.keepAlive()
}

// pumpEvent pumps incoming events to the Events channel.
func (w *WebSocket) pumpEvent() {
	defer func() {
		w.Close()
		close(w.Events)
	}()

	var msg string
	var err error
	for {
		if msg, err = w.Recv(); err != nil {
			if w.OnError != nil {
				w.OnError(err)
			}
			return
		}
		w.Events <- msg
	}
}

// keepAlive sends periodic ping messages to keep the WebSocket connection alive.
func (w *WebSocket) keepAlive() {
	ticker := time.NewTicker(PING_INTERVAL)
	defer ticker.Stop()

	// This is added as a precaution in case the parent context is canceled before calling `w.Close()`.
	defer w.Close()

	for {
		select {
		case <-ticker.C:
			if w.Send("\r\n") != nil {
				return
			}
		case <-w.runCtx.Done():
			return
		}
	}
}

// Send sends a message over the WebSocket connection.
func (w *WebSocket) Send(msg string) (err error) {
	if w.Connected {
		err = websocket.Message.Send(w.client, msg)
	} else {
		err = ErrNotConnected
	}
	return
}

// Recv receives a message from the WebSocket connection.
func (w *WebSocket) Recv() (msg string, err error) {
	if w.Connected {
		err = websocket.Message.Receive(w.client, &msg)
	} else {
		err = ErrNotConnected
	}
	return
}
