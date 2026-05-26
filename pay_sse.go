package xgdnpay

import (
	"context"

	"github.com/skylark8866/paysdk/sse"
)

type PaySSE struct {
	hub        *sse.Hub
	subscriber *sse.Subscriber
	cancel     context.CancelFunc
}

type PaySSEOption func(*paySSEConfig)

type paySSEConfig struct {
	baseURL string
	hubOpts []sse.HubOption
	subOpts []sse.SubscriberOption
}

func WithSSEBaseURL(url string) PaySSEOption {
	return func(c *paySSEConfig) {
		c.baseURL = url
	}
}

func WithSSEHubOpts(opts ...sse.HubOption) PaySSEOption {
	return func(c *paySSEConfig) {
		c.hubOpts = append(c.hubOpts, opts...)
	}
}

func NewPaySSE(opts ...PaySSEOption) *PaySSE {
	sdkCfg := loadSDKConfig()

	cfg := &paySSEConfig{
		baseURL: sdkCfg.Payment.BaseURL,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	hub := sse.NewHub(cfg.hubOpts...)
	ctx, cancel := context.WithCancel(context.Background())
	go hub.Run(ctx)

	subOpts := cfg.subOpts
	if cfg.baseURL != "" {
		subOpts = append(subOpts, sse.WithBaseURL(cfg.baseURL))
	}
	subscriber := sse.NewSubscriber(subOpts...)

	return &PaySSE{
		hub:        hub,
		subscriber: subscriber,
		cancel:     cancel,
	}
}

func (p *PaySSE) Hub() *sse.Hub {
	return p.hub
}

func (p *PaySSE) Subscribe(orderNo string, callback sse.PayNotifyCallback) {
	p.subscriber.Subscribe(orderNo, callback)
}

func (p *PaySSE) Shutdown() {
	p.subscriber.Close()
	p.cancel()
}