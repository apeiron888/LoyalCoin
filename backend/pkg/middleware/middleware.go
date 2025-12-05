package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/loyalcoin/backend/internal/auth"
	"github.com/loyalcoin/backend/internal/models"
	"github.com/loyalcoin/backend/pkg/logger"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"code":    "401_UNAUTHORIZED",
				"message": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Extract Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"code":    "401_UNAUTHORIZED",
				"message": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			logger.Warn("Invalid JWT token", map[string]interface{}{
				"error": err.Error(),
			})
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"code":    "401_UNAUTHORIZED",
				"message": "Invalid token",
			})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("wallet_address", claims.WalletAddress)

		c.Next()
	}
}

// RequireRole middleware checks if user has required role
func RequireRole(roles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"status":  "error",
				"code":    "403_FORBIDDEN",
				"message": "Role not found in token",
			})
			c.Abort()
			return
		}

		role := userRole.(models.Role)

		// Check if user has one of the required roles
		hasRole := false
		for _, requiredRole := range roles {
			if role == requiredRole {
				hasRole = true
				break
			}
		}

		if hasRole {
			c.Next()
			return
		}

		logger.Audit("UNAUTHORIZED_ACCESS_ATTEMPT", c.GetString("user_id"), map[string]interface{}{
			"user_role":      role,
			"required_roles": roles,
			"path":           c.Request.URL.Path,
		})

		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"code":    "403_FORBIDDEN",
			"message": "Insufficient permissions",
		})
		c.Abort()
	}
}

// CORSMiddleware handles CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequestLoggerMiddleware logs all requests
func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		logger.Info("HTTP Request", map[string]interface{}{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"ip":     c.ClientIP(),
		})

		c.Next()

		// Log response
		logger.Info("HTTP Response", map[string]interface{}{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"status": c.Writer.Status(),
		})
	}
}
