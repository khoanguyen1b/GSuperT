package auth

import (
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/user/gsupert/internal/common"
	"github.com/user/gsupert/internal/config"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			common.Error(c, 401, "UNAUTHORIZED", "Authorization header is required")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			common.Error(c, 401, "UNAUTHORIZED", "Authorization header must be Bearer token")
			c.Abort()
			return
		}

		claims, err := ValidateToken(parts[1], cfg.JWTSecret)
		if err != nil {
			common.Error(c, 401, "UNAUTHORIZED", "Invalid or expired token")
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func RoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			common.Error(c, 403, "FORBIDDEN", "User role not found")
			c.Abort()
			return
		}

		roleStr := userRole.(string)
		allowed := false
		for _, r := range roles {
			if r == roleStr {
				allowed = true
				break
			}
		}

		if !allowed {
			common.Error(c, 403, "FORBIDDEN", "You do not have permission to access this resource")
			c.Abort()
			return
		}

		c.Next()
	}
}
