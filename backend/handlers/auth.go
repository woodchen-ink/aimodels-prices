package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
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
		c.SetCookie("session", sessionID, int(24*time.Hour.Seconds()), "/", "aimodels-prices.q58.club", true, true)
		c.JSON(http.StatusOK, gin.H{"message": "Logged in successfully"})
		return
	}

	// 生产环境使用 OAuth 2.0
	clientID := os.Getenv("OAUTH_CLIENT_ID")
	redirectURI := os.Getenv("OAUTH_REDIRECT_URI")
	authorizeURL := os.Getenv("OAUTH_AUTHORIZE_URL")

	if clientID == "" || redirectURI == "" || authorizeURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OAuth configuration not found"})
		return
	}

	// 构建授权 URL
	authURL := fmt.Sprintf("%s?response_type=code&client_id=%s&redirect_uri=%s",
		authorizeURL,
		url.QueryEscape(clientID),
		url.QueryEscape(redirectURI))

	// 返回授权 URL 而不是直接重定向
	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
	})
}

func Logout(c *gin.Context) {
	cookie, err := c.Cookie("session")
	if err == nil {
		db := c.MustGet("db").(*sql.DB)
		db.Exec("DELETE FROM session WHERE id = ?", cookie)
	}

	c.SetCookie("session", "", -1, "/", "aimodels-prices.q58.club", true, true)
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
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing authorization code"})
		return
	}

	// 获取访问令牌
	tokenURL := os.Getenv("OAUTH_TOKEN_URL")
	clientID := os.Getenv("OAUTH_CLIENT_ID")
	clientSecret := os.Getenv("OAUTH_CLIENT_SECRET")
	redirectURI := os.Getenv("OAUTH_REDIRECT_URI")

	// 构建请求体
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("grant_type", "authorization_code")

	// 发送请求获取访问令牌
	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get access token"})
		return
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse token response"})
		return
	}

	// 使用访问令牌获取用户信息
	userURL := os.Getenv("OAUTH_USER_URL")
	req, err := http.NewRequest("GET", userURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user info request"})
		return
	}

	req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	client := &http.Client{}
	userResp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer userResp.Body.Close()

	var userInfo struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		Username  string `json:"username"`
		Admin     bool   `json:"admin"`
		AvatarURL string `json:"avatar_url"`
		Name      string `json:"name"`
	}

	if err := json.NewDecoder(userResp.Body).Decode(&userInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user info"})
		return
	}

	db := c.MustGet("db").(*sql.DB)

	// 检查用户是否存在
	var user models.User
	err = db.QueryRow("SELECT id, username, email, role FROM user WHERE email = ?", userInfo.Email).Scan(
		&user.ID, &user.Username, &user.Email, &user.Role)

	role := "user"
	if userInfo.Admin {
		role = "admin"
	}

	if err == sql.ErrNoRows {
		// 创建新用户
		result, err := db.Exec(`
			INSERT INTO user (username, email, role) 
			VALUES (?, ?, ?)`,
			userInfo.Username, userInfo.Email, role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
		userID, _ := result.LastInsertId()
		user = models.User{
			ID:       uint(userID),
			Username: userInfo.Username,
			Email:    userInfo.Email,
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
	c.SetCookie("session", sessionID, int(24*time.Hour.Seconds()), "/", "aimodels-prices.q58.club", true, true)

	// 重定向到前端
	c.Redirect(http.StatusTemporaryRedirect, "https://aimodels-prices.q58.club")
}
