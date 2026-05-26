package sse

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/skylark8866/paysdk/sse/protocol"
)

type PayNotifyEvent struct {
	Status        string `json:"status"`
	OrderNo       string `json:"order_no"`
	TransactionID string `json:"transaction_id"`
}

type PayNotifyCallback func(event *PayNotifyEvent)

type Subscriber struct {
	baseURL    string
	httpClient *http.Client
	mu         sync.Mutex
	subs       map[string]context.CancelFunc
	closed     bool
}

const (
	defaultReconnectDelay    = 1 * time.Second
	maxReconnectDelay        = 60 * time.Second
	defaultHTTPClientTimeout = 30 * time.Second
)

type SubscriberOption func(*Subscriber)

func WithBaseURL(url string) SubscriberOption {
	return func(s *Subscriber) {
		s.baseURL = strings.TrimSuffix(url, "/")
	}
}

func (s *Subscriber) setDefaults() {
	if s.baseURL == "" {
		s.baseURL = DefaultSubscriberBaseURL
	}
}

func NewSubscriber(opts ...SubscriberOption) *Subscriber {
	s := &Subscriber{
		httpClient: &http.Client{
			Timeout: 0,
		},
		subs: make(map[string]context.CancelFunc),
	}
	for _, opt := range opts {
		opt(s)
	}
	s.setDefaults()
	return s
}

func (s *Subscriber) Subscribe(orderNo string, callback PayNotifyCallback) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		log.Printf("[SSE Subscriber] 已关闭，无法订阅订单: %s", orderNo)
		return
	}

	if _, exists := s.subs[orderNo]; exists {
		s.mu.Unlock()
		log.Printf("[SSE Subscriber] 订单已订阅，跳过: %s", orderNo)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.subs[orderNo] = cancel
	s.mu.Unlock()

	log.Printf("[SSE Subscriber] 开始订阅订单: %s, URL: %s", orderNo, s.sseURL(orderNo))
	go s.connectLoop(ctx, orderNo, callback)
}

func (s *Subscriber) Unsubscribe(orderNo string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cancel, exists := s.subs[orderNo]; exists {
		cancel()
		delete(s.subs, orderNo)
	}
}

func (s *Subscriber) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.closed = true
	for orderNo, cancel := range s.subs {
		cancel()
		delete(s.subs, orderNo)
	}
}

func (s *Subscriber) IsSubscribed(orderNo string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.subs[orderNo]
	return exists
}

func (s *Subscriber) SubCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.subs)
}

func (s *Subscriber) sseURL(orderNo string) string {
	return fmt.Sprintf("%s%s", s.baseURL, strings.Replace(protocol.SubscribePath, "{orderNo}", orderNo, 1))
}

func (s *Subscriber) connectLoop(ctx context.Context, orderNo string, callback PayNotifyCallback) {
	defer func() {
		s.mu.Lock()
		delete(s.subs, orderNo)
		s.mu.Unlock()
	}()

	delay := defaultReconnectDelay

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		connCtx, connCancel := context.WithCancel(ctx)
		err := s.connectAndRead(connCtx, orderNo, callback)
		connCancel()

		if err != nil && !strings.Contains(err.Error(), "context canceled") {
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}

		if delay < maxReconnectDelay {
			delay *= 2
		}
	}
}

func (s *Subscriber) connectAndRead(ctx context.Context, orderNo string, callback PayNotifyCallback) error {
	url := s.sseURL(orderNo)

	log.Printf("[SSE Subscriber] 连接后端SSE: %s", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Printf("[SSE Subscriber] 创建请求失败: %v", err)
		return err
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf("[SSE Subscriber] HTTP请求失败: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[SSE Subscriber] SSE连接失败, 状态码: %d", resp.StatusCode)
		return fmt.Errorf("SSE连接失败, 状态码: %d", resp.StatusCode)
	}

	log.Printf("[SSE Subscriber] SSE连接成功: %s", orderNo)

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var eventName string
	var dataLines []string

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Text()

		if line == "" {
			if protocol.EventName(eventName).IsValid() && len(dataLines) > 0 {
				data := strings.Join(dataLines, "")
				log.Printf("[SSE Subscriber] 收到事件: %s, 数据: %s", eventName, data)
				var event PayNotifyEvent
				if err := json.Unmarshal([]byte(data), &event); err == nil {
					log.Printf("[SSE Subscriber] 触发回调: orderNo=%s, status=%s", event.OrderNo, event.Status)
					callback(&event)
				} else {
					log.Printf("[SSE Subscriber] 解析事件数据失败: %v", err)
				}
			}
			eventName = ""
			dataLines = nil
			continue
		}

		if strings.HasPrefix(line, ":") {
			continue
		}

		if strings.HasPrefix(line, "event:") {
			eventName = strings.TrimSpace(line[6:])
		} else if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(line[5:]))
		}
	}

	return scanner.Err()
}
