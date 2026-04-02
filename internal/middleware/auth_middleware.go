package middleware

import (
	"strings"

	"github.com/M4yankkkk/findash/pkg/utils"
	"github.com/gin-gonic/gin"
)

const (
	// ContextKeyUserID is the gin context key for the authenticated user's ID.
	ContextKeyUserID = "userID"
	// ContextKeyRole is the gin context key for the authenticated user's role.
	ContextKeyRole = "role"
	// ContextKeyEmail is the gin context key for the authenticated user's email.
	ContextKeyEmail = "email"
)

// Authenticate is a Gin middleware that validates the Bearer JWT in the
// Authorization header and injects the claims into the request context.
func Authenticate(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Unauthorized(c, "authorization header is required")
			c.Abort()
			return
		}

		// Expect: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.Unauthorized(c, "authorization header format must be: Bearer <token>")
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(parts[1], jwtSecret)
		if err != nil {
			utils.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		// Inject claims into context for downstream handlers
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyRole, string(claims.Role))
		c.Set(ContextKeyEmail, claims.Email)

		c.Next()
	}
}
