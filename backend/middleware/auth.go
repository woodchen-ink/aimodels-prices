package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"aimodels-prices/database"
	"aimodels-prices/models"
)

// HasPermission 检查用户是否拥有指定权限级别
// 权限级别：t0 → t1 → t2 → t3 → t4 → t5 → viewer → admin
func HasPermission(user *models.User, requiredLevel string) bool {
	if user == nil || user.Groups == "" {
		return false
	}

	groups := strings.ToLower(user.Groups)
	required := strings.ToLower(requiredLevel)
	return strings.Contains(groups, required)
}

// IsModerator 检查用户是否具有审核权限（t4或admin）
func IsModerator(user *models.User) bool {
	return HasPermission(user, "t4") || HasPermission(user, "admin")
}

// IsAdmin 检查用户是否具有管理员权限
func IsAdmin(user *models.User) bool {
	return HasPermission(user, "admin")
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("session")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not logged in"})
			c.Abort()
			return
		}

		var session models.Session
		if err := database.DB.Preload("User").Where("id = ? AND expires_at > ?", cookie, time.Now()).First(&session).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
			c.Abort()
			return
		}

		c.Set("user", &session.User)
		c.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not logged in"})
			c.Abort()
			return
		}

		u, ok := user.(*models.User)
		if !ok || !IsAdmin(u) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		u, ok := user.(*models.User)
		if !ok || !IsAdmin(u) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireModerator 要求用户具有审核权限（t4或admin）
func RequireModerator() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		u, ok := user.(*models.User)
		if !ok || !IsModerator(u) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Moderator access required (t4 or admin)"})
			c.Abort()
			return
		}
		c.Next()
	}
}
