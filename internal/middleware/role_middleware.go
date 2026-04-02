package middleware

import (
	"github.com/M4yankkkk/findash/internal/models"
	"github.com/M4yankkkk/findash/pkg/utils"
	"github.com/gin-gonic/gin"
)

// RequireRole returns a middleware that only allows users with one of the
// specified roles to proceed. Must be used after Authenticate middleware.
func RequireRole(allowed ...models.Role) gin.HandlerFunc {
	// Build a set for O(1) lookup
	roleSet := make(map[models.Role]struct{}, len(allowed))
	for _, r := range allowed {
		roleSet[r] = struct{}{}
	}

	return func(c *gin.Context) {
		roleStr, exists := c.Get(ContextKeyRole)
		if !exists {
			utils.Unauthorized(c, "not authenticated")
			c.Abort()
			return
		}

		role := models.Role(roleStr.(string))
		if _, ok := roleSet[role]; !ok {
			utils.Forbidden(c, "you do not have permission to perform this action")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin is a convenience wrapper for admin-only routes.
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(models.RoleAdmin)
}

// RequireManagerOrAdmin is a convenience wrapper for manager+ routes.
func RequireManagerOrAdmin() gin.HandlerFunc {
	return RequireRole(models.RoleManager, models.RoleAdmin)
}
