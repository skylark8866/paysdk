package sse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Message struct {
	ID    string      `json:"id,omitempty"`
	Event EventName   `json:"event,omitempty"`
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
	var buf bytes.Buffer

	if m.ID != "" {
		buf.WriteString("id: ")
		buf.WriteString(m.ID)
		buf.WriteString("\n")
	}

	if m.Event != "" {
		buf.WriteString("event: ")
		buf.WriteString(string(m.Event))
		buf.WriteString("\n")
	}

	dataBytes, _ := json.Marshal(m.Data)
	buf.WriteString("data: ")
	buf.Write(dataBytes)
	buf.WriteString("\n\n")

	return buf.Bytes()
}

func encodeJSON(v interface{}) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return formatSSE("", "", data), nil
}

func formatSSE(id string, event string, data []byte) []byte {
	var buf bytes.Buffer

	if id != "" {
		buf.WriteString("id: ")
		buf.WriteString(id)
		buf.WriteString("\n")
	}

	if event != "" {
		buf.WriteString("event: ")
		buf.WriteString(event)
		buf.WriteString("\n")
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		buf.WriteString("data: ")
		buf.WriteString(line)
		buf.WriteString("\n")
	}
	buf.WriteString("\n")

	return buf.Bytes()
}

func FormatEvent(event EventName, data []byte) []byte {
	return formatSSE("", string(event), data)
}

func FormatData(data []byte) []byte {
	return formatSSE("", "", data)
}

func FormatJSON(v interface{}) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return FormatData(data), nil
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
