package authz

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// AuthClient instance for middleware
type AuthMiddleware struct {
	client *AuthClient
}

func NewAuthMiddleware(authServiceURL string) *AuthMiddleware {
	return &AuthMiddleware{
		client: NewAuthClient(authServiceURL),
	}
}

// Fiber Middleware - Uses HTTP validation
func (a *AuthMiddleware) FiberMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")

		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authorization token required",
			})
		}

		// Extract token
		token := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}

		// Validate via HTTP call to auth service
		response, err := a.client.ValidateTokenHTTP(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "failed to validate token",
			})
		}

		if !response.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": response.Error,
			})
		}

		// Store user ID in context
		c.Locals("user_id", response.UserID)
		return c.Next()
	}
}

// Chi Middleware - Uses HTTP validation
func (a *AuthMiddleware) ChiMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, "authorization token required", http.StatusUnauthorized)
			return
		}

		// Extract token
		token := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}

		// Validate via HTTP call to auth service
		response, err := a.client.ValidateTokenHTTP(token)
		if err != nil {
			http.Error(w, "failed to validate token", http.StatusUnauthorized)
			return
		}

		if !response.Valid {
			http.Error(w, response.Error, http.StatusUnauthorized)
			return
		}

		// Store user ID in context
		ctx := WithUserID(r.Context(), response.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper functions for getting user ID from context
func GetUserIDFromFiber(c *fiber.Ctx) (string, bool) {
	userID := c.Locals("user_id")
	if userID == nil {
		return "", false
	}
	return userID.(string), true
}

func GetUserIDFromChi(r *http.Request) (string, bool) {
	return GetUserID(r.Context())
}
