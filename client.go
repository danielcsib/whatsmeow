// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package whatsmeow implements a client for the WhatsApp web.
package whatsmeow

import (
	"context"
	"sync"
	"sync/atomic"

	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/util/log"
	"go.mau.fi/whatsmeow/events"
)

// EventHandler is a function that can handle events received from WhatsApp.
type EventHandler func(evt interface{})

// Client is the main WhatsApp client struct.
type Client struct {
	Store   *store.Device
	Log     log.Logger

	// Event handlers registered via AddEventHandler
	eventHandlersLock sync.RWMutex
	eventHandlers     []wrappedEventHandler
	lastEventHandlerID uint32

	// Connection state
	connectionState int32 // atomic, see connectionState* constants

	// Context for managing goroutine lifecycle
	ctx    context.Context
	cancel context.CancelFunc
}

const (
	connectionStateDisconnected int32 = iota
	connectionStateConnecting
	connectionStateConnected
)

type wrappedEventHandler struct {
	fn EventHandler
	id uint32
}

// NewClient creates a new WhatsApp client with the given device store and logger.
func NewClient(deviceStore *store.Device, log log.Logger) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		Store:  deviceStore,
		Log:    log,
		ctx:    ctx,
		cancel: cancel,
	}
}

// AddEventHandler registers a new event handler function and returns its ID.
// The ID can be used to remove the handler later via RemoveEventHandler.
func (cli *Client) AddEventHandler(handler EventHandler) uint32 {
	id := atomic.AddUint32(&cli.lastEventHandlerID, 1)
	cli.eventHandlersLock.Lock()
	cli.eventHandlers = append(cli.eventHandlers, wrappedEventHandler{fn: handler, id: id})
	cli.eventHandlersLock.Unlock()
	return id
}

// RemoveEventHandler removes a previously registered event handler by its ID.
func (cli *Client) RemoveEventHandler(id uint32) bool {
	cli.eventHandlersLock.Lock()
	defer cli.eventHandlersLock.Unlock()
	for i, handler := range cli.eventHandlers {
		if handler.id == id {
			cli.eventHandlers = append(cli.eventHandlers[:i], cli.eventHandlers[i+1:]...)
			return true
		}
	}
	return false
}

// RemoveAllEventHandlers removes all registered event handlers.
// Useful for cleanup during testing or when reinitializing the client.
func (cli *Client) RemoveAllEventHandlers() {
	cli.eventHandlersLock.Lock()
	cli.eventHandlers = nil
	cli.eventHandlersLock.Unlock()
}

// dispatch sends an event to all registered event handlers.
// Handlers are called sequentially. If a handler panics, the panic is recovered
// so that remaining handlers still receive the event.
// Note: a copy of the handler slice is taken under RLock to avoid holding the
// lock while calling user-supplied functions (which may themselves call Add/RemoveEventHandler).
func (cli *Client) dispatch(evt interface{}) {
	cli.eventHandlersLock.RLock()
	// Copy the slice so we don't hold the lock during handler execution.
	handlers := make([]wrappedEventHandler, len(cli.eventHandlers))
	copy(handlers, cli.eventHandlers)
	cli.eventHandlersLock.RUnlock()
	for _, handler := range handlers {
		func() {
			defer func() {
				if r := recover(); r != nil {
					cli.Log.Errorf("Panic in event handler for %T: %v", evt, r)
				}
			}()
			handler.fn(evt)
		}()
	}
}

// ensure events package is used
var _ = events.Connected{}
