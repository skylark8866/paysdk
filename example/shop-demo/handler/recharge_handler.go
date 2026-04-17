package handler

import (
	"context"
	"shop-demo/middleware"
	"shop-demo/pkg/response"
	"shop-demo/service"
	"time"

	xgdnpay "github.com/skylark8866/paysdk"

	"github.com/gin-gonic/gin"
)

type RechargeHandler struct {
	rechargeService *service.RechargeService
	userHandler     *UserHandler
}

func NewRechargeHandler(rechargeService *service.RechargeService, userHandler *UserHandler) *RechargeHandler {
	return &RechargeHandler{
		rechargeService: rechargeService,
		userHandler:     userHandler,
	}
}

type CreateOrderRequest struct {
	PackageID string `json:"package_id" binding:"required"`
}

func (h *RechargeHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	userID := middleware.GetUserID(c)
	username := middleware.GetUsername(c)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	result, err := h.rechargeService.CreateOrder(ctx, userID, username, req.PackageID)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, gin.H{
		"order_no":     result.OrderNo,
		"pay_order_no": result.PayOrderNo,
		"pay_url":      result.PayURL,
		"code_url":     result.CodeURL,
		"pay_amount":   result.PayAmount,
		"bonus_amount": result.BonusAmount,
	})
}

func (h *RechargeHandler) GetStatus(c *gin.Context) {
	orderNo := c.Query("order_no")
	if orderNo == "" {
		response.BadRequest(c, "缺少订单号")
		return
	}

	order, err := h.rechargeService.GetOrder(orderNo)
	if err != nil {
		response.NotFound(c, "订单不存在")
		return
	}

	response.Success(c, gin.H{
		"order_no":     order.OrderNo,
		"status":       order.Status,
		"pay_amount":   order.PayAmount,
		"bonus_amount": order.BonusAmount,
		"paid_at":      formatTime(order.PaidAt),
	})
}

func (h *RechargeHandler) HandleCallback(req *xgdnpay.NotifyRequest) error {
	return h.rechargeService.HandlePaymentCallback(req.OutOrderNo, req.Status)
}

func (h *RechargeHandler) History(c *gin.Context) {
	user := h.userHandler.GetCurrentUser(c)
	if user == nil {
		c.Redirect(302, "/login")
		return
	}

	orders, _ := h.rechargeService.GetUserOrders(user.ID, 20)
	logs, _ := h.rechargeService.GetUserBalanceLogs(user.ID, 20)

	c.HTML(200, "history.html", gin.H{
		"User":   user,
		"Orders": orders,
		"Logs":   logs,
	})
}

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}
