package sse

type Client struct {
	Channel string
	Send    chan []byte
	done    chan struct{}
}

func newClient(channel string, bufferSize int) *Client {
	return &Client{
		Channel: channel,
		Send:    make(chan []byte, bufferSize),
		done:    make(chan struct{}),
	}
}

func (c *Client) Close() {
	select {
	case <-c.done:
	default:
		close(c.done)
		close(c.Send)
	}
}

func (c *Client) Done() <-chan struct{} {
	return c.done
}
