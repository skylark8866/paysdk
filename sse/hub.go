package sse

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Hub struct {
	clients       map[string]map[*Client]struct{}
	register      chan *Client
	unregister    chan *Client
	broadcast     chan broadcastMessage
	mu            sync.RWMutex
	bufferSize    int
	keepAlive     time.Duration
	maxClients    int
	maxPerChannel int
}

type broadcastMessage struct {
	channel string
	data    []byte
}

type HubOption func(*Hub)

func WithKeepAlive(d time.Duration) HubOption {
	return func(h *Hub) {
		h.keepAlive = d
	}
}

func WithHubBufferSize(size int) HubOption {
	return func(h *Hub) {
		h.bufferSize = size
	}
}

func WithMaxClients(n int) HubOption {
	return func(h *Hub) {
		h.maxClients = n
	}
}

func WithMaxPerChannel(n int) HubOption {
	return func(h *Hub) {
		h.maxPerChannel = n
	}
}

func NewHub(opts ...HubOption) *Hub {
	h := &Hub{
		clients:       make(map[string]map[*Client]struct{}),
		register:      make(chan *Client, DefaultHubRegisterBuf),
		unregister:    make(chan *Client, DefaultHubUnregBuf),
		broadcast:     make(chan broadcastMessage, DefaultHubBroadcastBuf),
		bufferSize:    DefaultClientBufSize,
		keepAlive:     DefaultKeepAlive * time.Second,
		maxClients:    DefaultMaxClients,
		maxPerChannel: DefaultMaxPerChannel,
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *Hub) Run(ctx context.Context) {
	keepAliveTicker := time.NewTicker(h.keepAlive)
	defer keepAliveTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.closeAll()
			return

		case client := <-h.register:
			h.mu.Lock()
			if h.totalClients() >= h.maxClients {
				h.mu.Unlock()
				client.Close()
				continue
			}
			if h.clients[client.Channel] == nil {
				h.clients[client.Channel] = make(map[*Client]struct{})
			}
			if len(h.clients[client.Channel]) >= h.maxPerChannel {
				h.mu.Unlock()
				client.Close()
				continue
			}
			h.clients[client.Channel][client] = struct{}{}
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.Channel]; ok {
				delete(clients, client)
				if len(clients) == 0 {
					delete(h.clients, client.Channel)
				}
			}
			h.mu.Unlock()
			client.Close()

		case msg := <-h.broadcast:
			h.mu.RLock()
			clients, ok := h.clients[msg.channel]
			if !ok {
				h.mu.RUnlock()
				continue
			}
			for client := range clients {
				select {
				case client.Send <- msg.data:
				default:
				}
			}
			h.mu.RUnlock()

		case <-keepAliveTicker.C:
			h.cleanStaleClients()
		}
	}
}

func (h *Hub) cleanStaleClients() {
	h.mu.Lock()
	defer h.mu.Unlock()

	keepAlive := []byte(SSECommentKeepAlive)
	var staleClients []*Client

	for channel, clients := range h.clients {
		for client := range clients {
			select {
			case client.Send <- keepAlive:
			default:
				staleClients = append(staleClients, client)
				delete(clients, client)
			}
		}
		if len(clients) == 0 {
			delete(h.clients, channel)
		}
	}

	for _, client := range staleClients {
		client.Close()
	}
}

func (h *Hub) totalClients() int {
	total := 0
	for _, clients := range h.clients {
		total += len(clients)
	}
	return total
}

func (h *Hub) Subscribe(channel string) *Client {
	client := newClient(channel, h.bufferSize)
	h.register <- client
	return client
}

func (h *Hub) TrySubscribe(channel string) (*Client, error) {
	h.mu.RLock()
	if h.totalClients() >= h.maxClients {
		h.mu.RUnlock()
		return nil, fmt.Errorf(ErrMsgTooManyClients)
	}
	if chClients := h.clients[channel]; len(chClients) >= h.maxPerChannel {
		h.mu.RUnlock()
		return nil, fmt.Errorf(ErrMsgTooManyChannel)
	}
	h.mu.RUnlock()

	client := newClient(channel, h.bufferSize)
	h.register <- client
	return client, nil
}

func (h *Hub) Unsubscribe(client *Client) {
	h.unregister <- client
}

func (h *Hub) Broadcast(channel string, data []byte) {
	h.broadcast <- broadcastMessage{channel: channel, data: data}
}

func (h *Hub) BroadcastJSON(channel string, v interface{}) error {
	data, err := encodeJSON(v)
	if err != nil {
		return err
	}
	h.Broadcast(channel, data)
	return nil
}

func (h *Hub) ClientCount(channel string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[channel])
}

func (h *Hub) TotalClients() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.totalClients()
}

func (h *Hub) ChannelCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) closeAll() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, clients := range h.clients {
		for client := range clients {
			client.Close()
		}
	}
	h.clients = make(map[string]map[*Client]struct{})
}
