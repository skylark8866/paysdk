package sse

const (
	HeaderContentType  = "Content-Type"
	HeaderCacheControl = "Cache-Control"
	HeaderConnection   = "Connection"
	HeaderACAO         = "Access-Control-Allow-Origin"

	SSEContentType  = "text/event-stream"
	SSECacheControl = "no-cache"
	SSEConnection   = "keep-alive"
	SSEAllowOrigin  = "*"

	SSECommentConnected = ": connected\n\n"
	SSECommentKeepAlive = ": keep-alive\n\n"

	SSEEventConnected = "event: connected\ndata: {}\n\n"

	JSONContentType = "application/json"
)

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

const (
	ErrMsgChannelEmpty     = "channel cannot be empty"
	ErrMsgChannelTooLong   = "channel too long (max 128)"
	ErrMsgChannelRequired  = "channel is required"
	ErrMsgStreamNotSupport = "streaming not supported"
	ErrMsgTooManyClients   = "too many connections"
	ErrMsgTooManyChannel   = "too many connections for this channel"
)
