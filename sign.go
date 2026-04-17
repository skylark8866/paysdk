package xgdnpay

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func GenerateTimestamp() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

func GenerateNonce() string {
	return generateNonce()
}

func (c *Client) buildSignedRequest(data interface{}) (*SignedRequest, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal data failed: %w", err)
	}

	sortedData, err := SortJSON(dataBytes)
	if err != nil {
		return nil, fmt.Errorf("sort json failed: %w", err)
	}

	timestamp := GenerateTimestamp()
	nonce := GenerateNonce()
	sign := CalculateSignV2(c.appID, timestamp, nonce, sortedData, c.appSecret)

	return &SignedRequest{
		AppID:     c.appID,
		Timestamp: timestamp,
		Nonce:     nonce,
		Data:      json.RawMessage(sortedData),
		Sign:      sign,
	}, nil
}

func CalculateSignV2(appID, timestamp, nonce, sortedData, appSecret string) string {
	var sb strings.Builder
	sb.Grow(len(appID) + len(sortedData) + len(nonce) + len(timestamp) + len(appSecret) + 64)
	sb.WriteString("app_id=")
	sb.WriteString(appID)
	sb.WriteString("&data=")
	sb.WriteString(sortedData)
	sb.WriteString("&nonce=")
	sb.WriteString(nonce)
	sb.WriteString("&timestamp=")
	sb.WriteString(timestamp)
	sb.WriteString("&app_secret=")
	sb.WriteString(appSecret)

	hash := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(hash[:])
}

func SortJSON(data []byte) (string, error) {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return "", err
	}
	sorted, err := json.Marshal(sortMap(m))
	if err != nil {
		return "", err
	}
	return string(sorted), nil
}

func sortMap(m map[string]interface{}) map[string]interface{} {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := make(map[string]interface{}, len(m))
	for _, k := range keys {
		v := m[k]
		switch val := v.(type) {
		case map[string]interface{}:
			result[k] = sortMap(val)
		case []interface{}:
			result[k] = sortSlice(val)
		default:
			result[k] = v
		}
	}
	return result
}

func sortSlice(s []interface{}) []interface{} {
	result := make([]interface{}, len(s))
	for i, v := range s {
		switch val := v.(type) {
		case map[string]interface{}:
			result[i] = sortMap(val)
		case []interface{}:
			result[i] = sortSlice(val)
		default:
			result[i] = v
		}
	}
	return result
}

func VerifySign(params map[string]string, appSecret string, maxDelay int64) error {
	sign, ok := params["sign"]
	if !ok || sign == "" {
		return fmt.Errorf(ErrMsgSignNotFound)
	}

	timestampStr, ok := params["timestamp"]
	if !ok || timestampStr == "" {
		return fmt.Errorf(ErrMsgTimestampNotFound)
	}

	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return fmt.Errorf(ErrMsgTimestampInvalid)
	}

	delay := time.Now().Unix() - timestamp
	if delay < 0 {
		delay = -delay
	}
	if delay > maxDelay {
		return fmt.Errorf(ErrMsgRequestExpired)
	}

	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.Grow(len(keys) * 64)
	for _, k := range keys {
		v := params[k]
		if v != "" {
			sb.WriteString(k)
			sb.WriteString("=")
			sb.WriteString(v)
			sb.WriteString("&")
		}
	}
	sb.WriteString("app_secret=")
	sb.WriteString(appSecret)

	hash := sha256.Sum256([]byte(sb.String()))
	calculatedSign := hex.EncodeToString(hash[:])

	if calculatedSign != sign {
		return fmt.Errorf(ErrMsgSignVerifyFail)
	}

	return nil
}
