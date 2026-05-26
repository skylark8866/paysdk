package protocol

import (
	"bytes"
	"strings"
)

func FormatEvent(event EventName, data []byte) []byte {
	return formatSSE("", string(event), data)
}

func FormatData(data []byte) []byte {
	return formatSSE("", "", data)
}

func FormatJSON(v interface{}) ([]byte, error) {
	data, err := marshalJSON(v)
	if err != nil {
		return nil, err
	}
	return FormatData(data), nil
}

func FormatConnectedEvent() []byte {
	return FormatEvent(EventConnected, []byte("{}"))
}

func FormatKeepAlive() []byte {
	return []byte(": keep-alive\n\n")
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
