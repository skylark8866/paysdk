package xgdnpay

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func GenerateTimestamp() string {
	return fmt.Sprintf("%d", time.Now().Unix())
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
	signStr := fmt.Sprintf("app_id=%s&data=%s&nonce=%s&timestamp=%s&app_secret=%s",
		appID, sortedData, nonce, timestamp, appSecret)
	hash := sha256.Sum256([]byte(signStr))
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
	result := make(map[string]interface{})
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
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
		return fmt.Errorf("签名不存在")
	}

	timestamp, ok := params["timestamp"]
	if !ok || timestamp == "" {
		return fmt.Errorf("时间戳不存在")
	}

	ts, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		ts, err = time.Parse("2006-01-02 15:04:05", timestamp)
		if err != nil {
			return fmt.Errorf("时间戳格式错误")
		}
	}

	delay := time.Now().Unix() - ts.Unix()
	if delay < 0 {
		delay = -delay
	}
	if delay > maxDelay {
		return fmt.Errorf("请求已过期")
	}

	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		v := params[k]
		if v != "" {
			sb.WriteString(fmt.Sprintf("%s=%s&", k, v))
		}
	}
	sb.WriteString("app_secret=" + appSecret)

	hash := sha256.Sum256([]byte(sb.String()))
	calculatedSign := hex.EncodeToString(hash[:])

	if calculatedSign != sign {
		return fmt.Errorf("签名验证失败")
	}

	return nil
}
