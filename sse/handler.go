package sse

import (
	"net/http"
)

type HandlerOption func(*handlerConfig)

type handlerConfig struct {
	channelParam    string
	channelHeader   string
	channelFunc     func(*http.Request) string
	beforeSubscribe func(*http.Request, string) error
	onConnect       func(*http.Request, string)
	onDisconnect    func(*http.Request, string)
}

func WithHandlerChannelParam(name string) HandlerOption {
	return func(c *handlerConfig) {
		c.channelParam = name
	}
}

func WithHandlerChannelHeader(name string) HandlerOption {
	return func(c *handlerConfig) {
		c.channelHeader = name
	}
}

func WithHandlerChannelFunc(fn func(*http.Request) string) HandlerOption {
	return func(c *handlerConfig) {
		c.channelFunc = fn
	}
}

func WithHandlerBeforeSubscribe(fn func(*http.Request, string) error) HandlerOption {
	return func(c *handlerConfig) {
		c.beforeSubscribe = fn
	}
}

func WithHandlerOnConnect(fn func(*http.Request, string)) HandlerOption {
	return func(c *handlerConfig) {
		c.onConnect = fn
	}
}

func WithHandlerOnDisconnect(fn func(*http.Request, string)) HandlerOption {
	return func(c *handlerConfig) {
		c.onDisconnect = fn
	}
}

func (h *Hub) Handler(opts ...HandlerOption) http.HandlerFunc {
	cfg := &handlerConfig{
		channelParam: DefaultChannelParam,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		channel := h.resolveHandlerChannel(r, cfg)
		if channel == "" {
			writeJSONError(w, http.StatusBadRequest, ErrMsgChannelRequired)
			return
		}

		if err := ValidateChannel(channel); err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}

		if cfg.beforeSubscribe != nil {
			if err := cfg.beforeSubscribe(r, channel); err != nil {
				writeJSONError(w, http.StatusForbidden, err.Error())
				return
			}
		}

		client, err := h.TrySubscribe(channel)
		if err != nil {
			writeJSONError(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		defer h.Unsubscribe(client)

		if cfg.onConnect != nil {
			cfg.onConnect(r, channel)
		}
		if cfg.onDisconnect != nil {
			defer cfg.onDisconnect(r, channel)
		}

		h.serveHandlerSSE(w, r, client)
	}
}

func (h *Hub) resolveHandlerChannel(r *http.Request, cfg *handlerConfig) string {
	if cfg.channelFunc != nil {
		return cfg.channelFunc(r)
	}

	if cfg.channelHeader != "" {
		if ch := r.Header.Get(cfg.channelHeader); ch != "" {
			return ch
		}
	}

	if cfg.channelParam != "" {
		if ch := r.URL.Query().Get(cfg.channelParam); ch != "" {
			return ch
		}
	}

	return ""
}

func (h *Hub) serveHandlerSSE(w http.ResponseWriter, r *http.Request, client *Client) {
	setSSEHeaders(w)

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, ErrMsgStreamNotSupport)
		return
	}

	w.Write(SSEEventConnected)
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-client.Done():
			return
		case msg, ok := <-client.Send:
			if !ok {
				return
			}
			w.Write(msg)
			flusher.Flush()
		}
	}
}
