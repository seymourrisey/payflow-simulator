package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"

	"github.com/seymourrisey/payflow-simulator/config"
	"github.com/seymourrisey/payflow-simulator/pkg/response"
)

type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return response.Unauthorized(c)
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(config.App.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			return response.Unauthorized(c)
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			return response.Unauthorized(c)
		}

		// Simpan user info ke context, bisa diakses handler berikutnya
		c.Locals("userID", claims.UserID)
		c.Locals("email", claims.Email)

		return c.Next()
	}
}

// helper ambil userId dari conext
func GetUserID(c *fiber.Ctx) string {
	return c.Locals("userID").(string)
}
