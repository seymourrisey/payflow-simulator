package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func OK(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Created(c *gin.Context, message string, data any) {
	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func BadRequest(c *gin.Context, err string) {
	c.JSON(http.StatusBadRequest, APIResponse{
		Success: false,
		Message: "Bad request",
		Error:   err,
	})
}

func Unauthorized(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, APIResponse{
		Success: false,
		Message: "Unauthorized",
		Error:   "Invalid or expired token",
	})
}

func NotFound(c *gin.Context, resource string) {
	c.JSON(http.StatusNotFound, APIResponse{
		Success: false,
		Message: resource + " not found",
	})
}

func InternalError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, APIResponse{
		Success: false,
		Message: "Internal server error",
		Error:   err.Error(),
	})
}

func Conflict(c *gin.Context, err string) {
	c.JSON(http.StatusConflict, APIResponse{
		Success: false,
		Message: "Conflict",
		Error:   err,
	})
}
