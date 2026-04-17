package handler

import (
	"fmt"
	"shop-demo/model"
	"shop-demo/service"

	"github.com/gin-gonic/gin"
)

type PageHandler struct {
	userService *service.UserService
	userHandler *UserHandler
}

func NewPageHandler(userService *service.UserService, userHandler *UserHandler) *PageHandler {
	return &PageHandler{
		userService: userService,
		userHandler: userHandler,
	}
}

func (h *PageHandler) Index(c *gin.Context) {
	user := h.userHandler.GetCurrentUser(c)
	c.HTML(200, "index.html", gin.H{
		"User":     user,
		"Packages": model.DefaultPackages,
	})
}

func (h *PageHandler) Login(c *gin.Context) {
	user := h.userHandler.GetCurrentUser(c)
	if user != nil {
		c.Redirect(302, "/")
		return
	}
	c.HTML(200, "login.html", nil)
}

func (h *PageHandler) Register(c *gin.Context) {
	user := h.userHandler.GetCurrentUser(c)
	if user != nil {
		c.Redirect(302, "/")
		return
	}
	c.HTML(200, "register.html", nil)
}

func (h *PageHandler) Recharge(c *gin.Context) {
	user := h.userHandler.GetCurrentUser(c)
	if user == nil {
		c.Redirect(302, "/login")
		return
	}

	packageID := c.Query("package_id")
	amountStr := c.Query("amount")

	var pkg *model.RechargePackage

	if packageID != "" {
		pkg = model.GetPackageByID(packageID)
	} else if amountStr != "" {
		var amount float64
		if _, err := fmt.Sscanf(amountStr, "%f", &amount); err == nil && amount >= 1 && amount <= 10000 {
			pkg = model.NewCustomPackage(amount)
		}
	}

	if pkg == nil {
		c.String(404, "充值套餐不存在或金额无效")
		return
	}

	c.HTML(200, "recharge.html", gin.H{
		"User":    user,
		"Package": pkg,
	})
}
