package protocol

import sdkprotocol "github.com/skylark8866/paysdk/protocol"

type EventName string

const (
	EventConnected    EventName = "connected"
	EventPayNotify    EventName = "pay_notify"
	EventRefundNotify EventName = "refund_notify"
	EventKeepAlive    EventName = "keep_alive"
)

var validEventNames = map[EventName]bool{
	EventConnected:    true,
	EventPayNotify:    true,
	EventRefundNotify: true,
	EventKeepAlive:    true,
}

func (e EventName) IsValid() bool {
	return validEventNames[e]
}

func (e EventName) String() string {
	return string(e)
}

const (
	HeaderContentType  = "Content-Type"
	HeaderCacheControl = "Cache-Control"
	HeaderConnection   = "Connection"
	HeaderACAO         = "Access-Control-Allow-Origin"

	SSEContentType  = "text/event-stream"
	SSECacheControl = "no-cache"
	SSEConnection   = "keep-alive"
	SSEAllowOrigin  = "*"

	JSONContentType = "application/json"
)

const SubscribePath = sdkprotocol.PathSSESubscribe
