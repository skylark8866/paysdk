package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	xgdnpay "xgdn-pay"
)

var (
	client     *xgdnpay.Client
	orders     = make(map[string]*Order)
	ordersLock sync.RWMutex
	templates  *template.Template
	appSecret  string
	baseURL    string
)

type Product struct {
	ID    string
	Name  string
	Price float64
	Desc  string
}

type Order struct {
	OrderNo    string
	OutOrderNo string
	ProductID  string
	Amount     float64
	Status     int
	PaidAt     string
	CreatedAt  time.Time
}

var products = []Product{
	{ID: "p001", Name: "测试商品A", Price: 0.01, Desc: "这是一个测试商品，价格0.01元"},
	{ID: "p002", Name: "测试商品B", Price: 0.02, Desc: "这是另一个测试商品，价格0.02元"},
	{ID: "p003", Name: "测试商品C", Price: 0.05, Desc: "这是第三个测试商品，价格0.05元"},
}

func main() {
	appID := getEnv("XGDN_APP_ID", "")
	appSecret = getEnv("XGDN_APP_SECRET", "")
	baseURL = getEnv("XGDN_BASE_URL", "https://pay.xgdn.net")

	client = xgdnpay.NewClient(
		appID,
		appSecret,
		xgdnpay.WithBaseURL(baseURL),
		xgdnpay.WithTimeout(30*time.Second),
	)

	var err error
	templates, err = template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatal("模板加载失败:", err)
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/pay", payHandler)
	http.HandleFunc("/result", resultHandler)
	http.HandleFunc("/api/create-order", createOrderAPI)
	http.HandleFunc("/api/order-status", orderStatusAPI)
	http.HandleFunc("/api/callback/pay", payCallbackHandler)
	http.HandleFunc("/api/events", sseHandler)

	port := getEnv("PORT", "8080")
	fmt.Printf("商城服务启动在 http://localhost:%s\n", port)
	fmt.Println("使用 Ctrl+C 停止服务")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Products": products,
	}
	renderTemplate(w, "index.html", data)
}

func payHandler(w http.ResponseWriter, r *http.Request) {
	productID := r.URL.Query().Get("product_id")
	if productID == "" {
		http.Error(w, "缺少商品ID", 400)
		return
	}

	var product *Product
	for i := range products {
		if products[i].ID == productID {
			product = &products[i]
			break
		}
	}
	if product == nil {
		http.Error(w, "商品不存在", 404)
		return
	}

	data := map[string]interface{}{
		"Product": product,
		"BaseURL": baseURL,
	}
	renderTemplate(w, "pay.html", data)
}

func resultHandler(w http.ResponseWriter, r *http.Request) {
	orderNo := r.URL.Query().Get("order_no")
	if orderNo == "" {
		http.Error(w, "缺少订单号", 400)
		return
	}

	ordersLock.RLock()
	order, exists := orders[orderNo]
	ordersLock.RUnlock()

	if !exists {
		http.Error(w, "订单不存在", 404)
		return
	}

	data := map[string]interface{}{
		"Order": order,
	}
	renderTemplate(w, "result.html", data)
}

func createOrderAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var req struct {
		ProductID string `json:"product_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "参数错误", 400)
		return
	}

	var product *Product
	for i := range products {
		if products[i].ID == req.ProductID {
			product = &products[i]
			break
		}
	}
	if product == nil {
		http.Error(w, "商品不存在", 404)
		return
	}

	outOrderNo := fmt.Sprintf("SHOP_%d", time.Now().UnixNano())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	notifyURL := fmt.Sprintf("http://localhost:%s/api/callback/pay", getEnv("PORT", "8080"))
	if baseURL := os.Getenv("PUBLIC_URL"); baseURL != "" {
		notifyURL = baseURL + "/api/callback/pay"
	}

	order, err := client.CreateOrder(ctx, &xgdnpay.CreateOrderRequest{
		OutOrderNo: outOrderNo,
		Amount:     product.Price,
		Title:      product.Name,
		PayType:    xgdnpay.PayTypeNative,
		NotifyURL:  notifyURL,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("创建订单失败: %v", err), 500)
		return
	}

	localOrder := &Order{
		OrderNo:    order.OrderNo,
		OutOrderNo: outOrderNo,
		ProductID:  product.ID,
		Amount:     product.Price,
		Status:     0,
		CreatedAt:  time.Now(),
	}

	ordersLock.Lock()
	orders[order.OrderNo] = localOrder
	ordersLock.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code": 0,
		"data": map[string]interface{}{
			"order_no":   order.OrderNo,
			"pay_url":    order.PayURL,
			"code_url":   order.CodeURL,
			"amount":     product.Price,
			"product_id": product.ID,
		},
	})
}

func orderStatusAPI(w http.ResponseWriter, r *http.Request) {
	orderNo := r.URL.Query().Get("order_no")
	if orderNo == "" {
		http.Error(w, "缺少订单号", 400)
		return
	}

	// 优先从 XGDN 平台查询订单状态
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	orderInfo, err := client.QueryOrder(ctx, orderNo)
	if err == nil && orderInfo != nil {
		// 更新本地订单状态
		ordersLock.Lock()
		if localOrder, exists := orders[orderNo]; exists {
			localOrder.Status = int(orderInfo.Status)
			localOrder.PaidAt = orderInfo.PaidAt
		}
		ordersLock.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"order_no": orderInfo.OrderNo,
				"status":   orderInfo.Status,
				"paid_at":  orderInfo.PaidAt,
			},
		})
		return
	}

	// 如果平台查询失败，回退到本地查询
	ordersLock.RLock()
	order, exists := orders[orderNo]
	ordersLock.RUnlock()

	if !exists {
		http.Error(w, "订单不存在", 404)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code": 0,
		"data": map[string]interface{}{
			"order_no": order.OrderNo,
			"status":   order.Status,
			"paid_at":  order.PaidAt,
		},
	})
}

func payCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var req xgdnpay.NotifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "参数错误", 400)
		return
	}

	if err := xgdnpay.VerifyNotify(&req, appSecret, 300); err != nil {
		http.Error(w, "签名验证失败", 400)
		return
	}

	ordersLock.Lock()
	if order, exists := orders[req.OrderNo]; exists {
		order.Status = int(req.Status)
		if req.Status == xgdnpay.OrderStatusPaid {
			order.PaidAt = req.PaidAt
		}
	}
	ordersLock.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    0,
		"message": "success",
	})
}

func sseHandler(w http.ResponseWriter, r *http.Request) {
	orderNo := r.URL.Query().Get("order_no")
	if orderNo == "" {
		http.Error(w, "缺少订单号", 400)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "不支持 SSE", 500)
		return
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeout := time.After(5 * time.Minute)

	for {
		select {
		case <-ticker.C:
			ordersLock.RLock()
			order, exists := orders[orderNo]
			ordersLock.RUnlock()

			if !exists {
				fmt.Fprintf(w, "data: {\"error\":\"订单不存在\"}\n\n")
				flusher.Flush()
				return
			}

			data, _ := json.Marshal(map[string]interface{}{
				"order_no": order.OrderNo,
				"status":   order.Status,
				"paid_at":  order.PaidAt,
			})
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()

			if order.Status == 1 {
				return
			}

		case <-timeout:
			return

		case <-r.Context().Done():
			return
		}
	}
}

func renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
