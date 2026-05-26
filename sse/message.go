package sse

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/skylark8866/paysdk/sse/protocol"
)

type SSEMessage interface {
	EventName() EventName
	ToJSON() ([]byte, error)
}

type Message struct {
	ID    string    `json:"id,omitempty"`
	Event EventName `json:"event,omitempty"`
	Data  interface{} `json:"data"`
}

func NewMessage(data interface{}) *Message {
	return &Message{Data: data}
}

func (m *Message) SetID(id string) *Message {
	m.ID = id
	return m
}

func (m *Message) SetEvent(event EventName) *Message {
	m.Event = event
	return m
}

func (m *Message) Bytes() []byte {
	dataBytes, _ := json.Marshal(m.Data)
	return protocol.FormatEvent(m.Event, dataBytes)
}

func FormatEvent(event EventName, data []byte) []byte {
	return protocol.FormatEvent(event, data)
}

func FormatData(data []byte) []byte {
	return protocol.FormatData(data)
}

func FormatJSON(v interface{}) ([]byte, error) {
	return protocol.FormatJSON(v)
}

func GenerateID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

func ValidateChannel(channel string) error {
	if channel == "" {
		return fmt.Errorf(ErrMsgChannelEmpty)
	}
	if len(channel) > MaxChannelLength {
		return fmt.Errorf(ErrMsgChannelTooLong)
	}
	return nil
}
