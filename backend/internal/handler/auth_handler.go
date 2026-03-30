package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/seymourrisey/payflow-simulator/internal/dto"
	"github.com/seymourrisey/payflow-simulator/internal/middleware"
	"github.com/seymourrisey/payflow-simulator/internal/service"
	"github.com/seymourrisey/payflow-simulator/pkg/response"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	result, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		if err.Error() == "email already registered" {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Created(c, "Registration successful", result)
}

// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	result, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, "Login successful", result)
}

// POST /api/auth/logout — JWT stateless: token dihapus di sisi client
// Backend hanya konfirmasi logout berhasil + return info user yang logout
func (h *AuthHandler) Logout(c *gin.Context) {
	userID := middleware.GetUserID(c)
	email, _ := c.Get("email")

	response.OK(c, "Logout successful", gin.H{
		"user_id": userID,
		"email":   email,
		"message": "Token has been invalidated on client side",
	})
}
