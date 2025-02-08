package handlers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"aimodels-prices/models"
)

func generateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func GetAuthStatus(c *gin.Context) {
	cookie, err := c.Cookie("session")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not logged in"})
		return
	}

	db := c.MustGet("db").(*sql.DB)
	var session models.Session
	err = db.QueryRow("SELECT id, user_id, expires_at, created_at, updated_at, deleted_at FROM session WHERE id = ?", cookie).Scan(
		&session.ID, &session.UserID, &session.ExpiresAt, &session.CreatedAt, &session.UpdatedAt, &session.DeletedAt)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	if session.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired"})
		return
	}

	user, err := session.GetUser(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	c.Set("user", user)
	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

func Login(c *gin.Context) {
	// 开发环境下使用测试账号
	if gin.Mode() != gin.ReleaseMode {
		db := c.MustGet("db").(*sql.DB)

		// 创建测试用户(如果不存在)
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM user WHERE username = 'admin'").Scan(&count)
		if err != nil || count == 0 {
			_, err = db.Exec("INSERT INTO user (username, email, role) VALUES (?, ?, ?)",
				"admin", "admin@test.com", "admin")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create test user"})
				return
			}
		}

		// 获取用户ID
		var userID uint
		err = db.QueryRow("SELECT id FROM user WHERE username = 'admin'").Scan(&userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
			return
		}

		// 创建会话
		sessionID := generateSessionID()
		expiresAt := time.Now().Add(24 * time.Hour)
		_, err = db.Exec("INSERT INTO session (id, user_id, expires_at) VALUES (?, ?, ?)",
			sessionID, userID, expiresAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
			return
		}

		// 设置cookie
		c.SetCookie("session", sessionID, int(24*time.Hour.Seconds()), "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"message": "Logged in successfully"})
		return
	}

	// 生产环境使用 Discourse SSO
	discourseURL := os.Getenv("DISCOURSE_URL")
	if discourseURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Discourse URL not configured"})
		return
	}

	// 生成随机 nonce
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate nonce"})
		return
	}
	nonceStr := hex.EncodeToString(nonce)

	// 构建 payload
	payload := url.Values{}
	payload.Set("nonce", nonceStr)
	payload.Set("return_sso_url", fmt.Sprintf("https://aimodels-prices.q58.pro/api/auth/callback"))

	// Base64 编码
	payloadStr := base64.StdEncoding.EncodeToString([]byte(payload.Encode()))

	// 计算签名
	ssoSecret := os.Getenv("DISCOURSE_SSO_SECRET")
	if ssoSecret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SSO secret not configured"})
		return
	}

	h := hmac.New(sha256.New, []byte(ssoSecret))
	h.Write([]byte(payloadStr))
	sig := hex.EncodeToString(h.Sum(nil))

	// 构建重定向 URL
	redirectURL := fmt.Sprintf("%s/session/sso_provider?sso=%s&sig=%s",
		discourseURL, url.QueryEscape(payloadStr), sig)

	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func Logout(c *gin.Context) {
	cookie, err := c.Cookie("session")
	if err == nil {
		db := c.MustGet("db").(*sql.DB)
		db.Exec("DELETE FROM session WHERE id = ?", cookie)
	}

	c.SetCookie("session", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func GetUser(c *gin.Context) {
	cookie, err := c.Cookie("session")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not logged in"})
		return
	}

	db := c.MustGet("db").(*sql.DB)
	var session models.Session
	if err := db.QueryRow("SELECT id, user_id, expires_at, created_at, updated_at, deleted_at FROM session WHERE id = ?", cookie).Scan(
		&session.ID, &session.UserID, &session.ExpiresAt, &session.CreatedAt, &session.UpdatedAt, &session.DeletedAt); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	user, err := session.GetUser(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

func AuthCallback(c *gin.Context) {
	sso := c.Query("sso")
	sig := c.Query("sig")

	if sso == "" || sig == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing parameters"})
		return
	}

	// 获取 SSO 密钥
	ssoSecret := os.Getenv("DISCOURSE_SSO_SECRET")
	if ssoSecret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SSO secret not configured"})
		return
	}

	// 验证签名
	h := hmac.New(sha256.New, []byte(ssoSecret))
	h.Write([]byte(sso))
	computedSig := hex.EncodeToString(h.Sum(nil))
	if computedSig != sig {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature"})
		return
	}

	// 解码 SSO payload
	payload, err := base64.StdEncoding.DecodeString(sso)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid SSO payload"})
		return
	}

	// 解析 payload
	values, err := url.ParseQuery(string(payload))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload format"})
		return
	}

	// 获取用户信息
	username := values.Get("username")
	email := values.Get("email")
	groups := values.Get("groups")
	admin := values.Get("admin")         // Discourse 管理员标志
	moderator := values.Get("moderator") // Discourse 版主标志
	if username == "" || email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user information"})
		return
	}

	// 判断用户角色
	role := "user"
	// 如果是管理员、版主或属于 admins 组，都赋予管理权限
	if admin == "true" || moderator == "true" || (groups != "" && strings.Contains(groups, "admins")) {
		role = "admin"
	}

	db := c.MustGet("db").(*sql.DB)

	// 检查用户是否存在
	var user models.User
	err = db.QueryRow("SELECT id, username, email, role FROM user WHERE email = ?", email).Scan(
		&user.ID, &user.Username, &user.Email, &user.Role)

	if err == sql.ErrNoRows {
		// 创建新用户
		result, err := db.Exec(`
			INSERT INTO user (username, email, role) 
			VALUES (?, ?, ?)`,
			username, email, role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
		userID, _ := result.LastInsertId()
		user = models.User{
			ID:       uint(userID),
			Username: username,
			Email:    email,
			Role:     role,
		}
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	} else {
		// 更新现有用户的角色（如果需要）
		if user.Role != role {
			_, err = db.Exec("UPDATE user SET role = ? WHERE id = ?", role, user.ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user role"})
				return
			}
			user.Role = role
		}
	}

	// 创建会话
	sessionID := generateSessionID()
	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = db.Exec("INSERT INTO session (id, user_id, expires_at) VALUES (?, ?, ?)",
		sessionID, user.ID, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// 设置 cookie
	c.SetCookie("session", sessionID, int(24*time.Hour.Seconds()), "/", "", false, true)

	// 重定向到前端
	c.Redirect(http.StatusTemporaryRedirect, "/")
}
