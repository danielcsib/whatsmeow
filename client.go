// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package whatsmeow implements a WhatsApp web multiice client.
package whatsmeow

import (
	au.fimau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"go.mau.fi/whatsmeow/util/log"
)

// EventHandler is a function that can handle events from WhatsApp.
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

// dispatch sends an event to all registered event handlers.
func (cli *Client) dispatch(evt interface{}) {
	cli.eventHandlersLock.RLock()
	handlers := cli.eventHandlers
	cli.eventHandlersLock.RUnlock()
	for _, handler := range handlers {
		handler.fn(evt)
	}
}

// IsConnected returns true if the client is currently connected to WhatsApp.
func (cli *Client) IsConnected() bool {
	return atomic.LoadInt32(&cli.connectionState) == connectionStateConnected
}

// IsLoggedIn returns true if the client has valid credentials stored.
func (cli *Client) IsLoggedIn() bool {
	return cli.Store != nil && cli.Store.ID != nil
}

// GetJID returns the JID of the currently logged-in user, or an empty JID if not logged in.
func (cli *Client) GetJID() types.JID {
	if cli.Store == nil || cli.Store.ID == nil {
		return types.EmptyJID
	}
	return *cli.Store.ID
}

// Disconnect disconnects the client from WhatsApp and cleans up resources.
func (cli *Client) Disconnect() {
	if atomic.CompareAndSwapInt32(&cli.connectionState, connectionStateConnected, connectionStateDisconnected) {
		cli.cancel()
		cli.dispatch(&events.Disconnected{})
		cli.Log.Infof("Disconnected from WhatsApp")
	}
}
