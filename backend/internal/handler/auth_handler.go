package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/seymourrisey/payflow-simulator/internal/dto"
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
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	result, err := h.authService.Register(c.Context(), &req)
	if err != nil {
		if err.Error() == "email already registered" {
			return response.Conflict(c, err.Error())
		}
		return response.InternalError(c, err)
	}

	return response.Created(c, "Registration successful", result)
}

// POST /api/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	result, err := h.authService.Login(c.Context(), &req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.OK(c, "Login successful", result)
}
