package handler

import (
	"shop-demo/middleware"
	"shop-demo/model"
	"shop-demo/pkg/response"
	"shop-demo/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	user, token, err := h.userService.Register(req.Username, req.Password)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	c.SetCookie("token", token, 86400*7, "/", "", false, true)
	response.Success(c, gin.H{
		"user_id":  user.ID,
		"username": user.Username,
		"balance":  user.Balance,
	})
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	user, token, err := h.userService.Login(req.Username, req.Password)
	if err != nil {
		response.Error(c, 401, err.Error())
		return
	}

	c.SetCookie("token", token, 86400*7, "/", "", false, true)
	response.Success(c, gin.H{
		"user_id":  user.ID,
		"username": user.Username,
		"balance":  user.Balance,
	})
}

func (h *UserHandler) Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
	response.Success(c, nil)
}

func (h *UserHandler) GetInfo(c *gin.Context) {
	userID := middleware.GetUserID(c)
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		response.InternalError(c, "获取用户信息失败")
		return
	}

	response.Success(c, gin.H{
		"user_id":  user.ID,
		"username": user.Username,
		"balance":  user.Balance,
	})
}

func (h *UserHandler) GetCurrentUser(c *gin.Context) *model.User {
	cookie, err := c.Cookie("token")
	if err != nil {
		return nil
	}

	claims, err := h.userService.ParseToken(cookie)
	if err != nil {
		return nil
	}

	user, _ := h.userService.GetUserByID(claims.UserID)
	return user
}
