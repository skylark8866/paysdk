package sse

import "github.com/skylark8866/paysdk/sse/protocol"

type EventName = protocol.EventName

const (
	EventConnected    = protocol.EventConnected
	EventPayNotify    = protocol.EventPayNotify
	EventRefundNotify = protocol.EventRefundNotify
	EventKeepAlive    = protocol.EventKeepAlive
)

const (
	HeaderContentType  = protocol.HeaderContentType
	HeaderCacheControl = protocol.HeaderCacheControl
	HeaderConnection   = protocol.HeaderConnection
	HeaderACAO         = protocol.HeaderACAO

	SSEContentType  = protocol.SSEContentType
	SSECacheControl = protocol.SSECacheControl
	SSEConnection   = protocol.SSEConnection
	SSEAllowOrigin  = protocol.SSEAllowOrigin

	SSECommentConnected = ": connected\n\n"
	SSECommentKeepAlive = ": keep-alive\n\n"

	JSONContentType = protocol.JSONContentType

	RespFieldCode    = "code"
	RespFieldMessage = "message"
	RespFieldError   = "error"
)

var SSEEventConnected = protocol.FormatConnectedEvent()

const (
	DefaultChannelParam    = "channel"
	DefaultHubRegisterBuf  = 100
	DefaultHubUnregBuf     = 100
	DefaultHubBroadcastBuf = 1000
	DefaultClientBufSize   = 256
	DefaultKeepAlive       = 15

	MaxChannelLength     = 128
	DefaultMaxClients    = 1000
	DefaultMaxPerChannel = 10
)

const DefaultSubscriberBaseURL = "https://pay.xgdn.net"

const (
	ErrMsgChannelEmpty     = "channel cannot be empty"
	ErrMsgChannelTooLong   = "channel too long (max 128)"
	ErrMsgChannelRequired  = "channel is required"
	ErrMsgStreamNotSupport = "streaming not supported"
	ErrMsgTooManyClients   = "too many connections"
	ErrMsgTooManyChannel   = "too many connections for this channel"
)
