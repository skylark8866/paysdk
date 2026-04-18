package sse

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GinOption func(*ginConfig)

type ginConfig struct {
	channelParam    string
	channelHeader   string
	channelFunc     func(*gin.Context) string
	beforeSubscribe func(*gin.Context, string) error
	onConnect       func(*gin.Context, string)
	onDisconnect    func(*gin.Context, string)
	sendConnect     bool
}

func WithChannelParam(name string) GinOption {
	return func(c *ginConfig) {
		c.channelParam = name
	}
}

func WithChannelHeader(name string) GinOption {
	return func(c *ginConfig) {
		c.channelHeader = name
	}
}

func WithChannelFunc(fn func(*gin.Context) string) GinOption {
	return func(c *ginConfig) {
		c.channelFunc = fn
	}
}

func WithBeforeSubscribe(fn func(*gin.Context, string) error) GinOption {
	return func(c *ginConfig) {
		c.beforeSubscribe = fn
	}
}

func WithOnConnect(fn func(*gin.Context, string)) GinOption {
	return func(c *ginConfig) {
		c.onConnect = fn
	}
}

func WithOnDisconnect(fn func(*gin.Context, string)) GinOption {
	return func(c *ginConfig) {
		c.onDisconnect = fn
	}
}

func WithConnectMessage() GinOption {
	return func(c *ginConfig) {
		c.sendConnect = true
	}
}

func (h *Hub) GinHandler(opts ...GinOption) gin.HandlerFunc {
	cfg := &ginConfig{
		channelParam: DefaultChannelParam,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return func(c *gin.Context) {
		channel := h.resolveGinChannel(c, cfg)
		if channel == "" {
			c.JSON(http.StatusBadRequest, gin.H{RespFieldError: ErrMsgChannelRequired})
			return
		}

		if err := ValidateChannel(channel); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{RespFieldError: err.Error()})
			return
		}

		if cfg.beforeSubscribe != nil {
			if err := cfg.beforeSubscribe(c, channel); err != nil {
				c.JSON(http.StatusForbidden, gin.H{RespFieldError: err.Error()})
				return
			}
		}

		client, err := h.TrySubscribe(channel)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{RespFieldError: err.Error()})
			return
		}
		defer h.Unsubscribe(client)

		if cfg.onConnect != nil {
			cfg.onConnect(c, channel)
		}
		if cfg.onDisconnect != nil {
			defer cfg.onDisconnect(c, channel)
		}

		if cfg.sendConnect {
			h.serveGinSSEManual(c, client)
		} else {
			h.serveGinSSEStream(c, client)
		}
	}
}

func (h *Hub) resolveGinChannel(c *gin.Context, cfg *ginConfig) string {
	if cfg.channelFunc != nil {
		return cfg.channelFunc(c)
	}

	if cfg.channelHeader != "" {
		if ch := c.GetHeader(cfg.channelHeader); ch != "" {
			return ch
		}
	}

	if cfg.channelParam != "" {
		if ch := c.Param(cfg.channelParam); ch != "" {
			return ch
		}
		if ch := c.Query(cfg.channelParam); ch != "" {
			return ch
		}
	}

	return ""
}

func (h *Hub) serveGinSSEStream(c *gin.Context, client *Client) {
	setSSEHeaders(c.Writer)

	c.Stream(func(w io.Writer) bool {
		select {
		case <-c.Request.Context().Done():
			return false
		case <-client.Done():
			return false
		case msg, ok := <-client.Send:
			if !ok {
				return false
			}
			w.Write(msg)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			return true
		}
	})
}

func (h *Hub) serveGinSSEManual(c *gin.Context, client *Client) {
	setSSEHeaders(c.Writer)

	c.Writer.Write(SSEEventConnected)
	c.Writer.(http.Flusher).Flush()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-client.Done():
			return
		case msg, ok := <-client.Send:
			if !ok {
				return
			}
			c.Writer.Write(msg)
			c.Writer.(http.Flusher).Flush()
		}
	}
}
