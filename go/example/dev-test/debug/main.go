package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

func main() {
	fmt.Println("=== XGDN Pay SDK 调试测试 ===")
	fmt.Println()

	appID := "app_your_app_id"
	appSecret := "your_app_secret"
	baseURL := "http://localhost:8093"

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := generateNonce()

	businessParams := map[string]string{
		"out_order_no": fmt.Sprintf("SDK_TEST_%d", time.Now().Unix()),
		"amount":       "0.01",
		"title":        "SDK开发测试",
		"pay_type":     "NATIVE",
		"return_url":   "http://localhost:3001/result",
	}

	params := map[string]string{
		"app_id":    appID,
		"timestamp": timestamp,
		"nonce":     nonce,
	}

	for k, v := range businessParams {
		params[k] = v
	}

	sign := calculateSign(params, appSecret)
	params["sign"] = sign

	fmt.Println("=== 请求参数 ===")
	for k, v := range params {
		fmt.Printf("  %s: %s\n", k, v)
	}
	fmt.Println()

	fmt.Printf("=== 签名原文 ===\n%s\n\n", getSignString(params, appSecret))
	fmt.Printf("=== 计算签名 ===\n%s\n\n", sign)

	body, _ := json.Marshal(params)
	fmt.Printf("=== 请求体 ===\n%s\n\n", string(body))

	req, err := http.NewRequest("POST", baseURL+"/api/v1/order/create", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	fmt.Printf("=== 响应状态 ===\n%d\n\n", resp.StatusCode)

	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, respBody, "", "  ")
	fmt.Printf("=== 响应内容 ===\n%s\n", prettyJSON.String())
}

func calculateSign(params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys)+1)
	for _, k := range keys {
		if params[k] != "" {
			parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
		}
	}
	parts = append(parts, fmt.Sprintf("app_secret=%s", secret))

	joined := strings.Join(parts, "&")
	hash := sha256.Sum256([]byte(joined))
	return hex.EncodeToString(hash[:])
}

func getSignString(params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys)+1)
	for _, k := range keys {
		if params[k] != "" {
			parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
		}
	}
	parts = append(parts, fmt.Sprintf("app_secret=%s", secret))
	return strings.Join(parts, "&")
}

func generateNonce() string {
	return strconv.FormatInt(int64(time.Now().UnixNano()), 36) + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	rand.Seed(time.Now().UnixNano())
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
