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
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"aimodels-prices/config"
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
	// 生产环境使用 OAuth 2.0
	clientID := os.Getenv("OAUTH_CLIENT_ID")
	redirectURI := os.Getenv("OAUTH_REDIRECT_URI")
	authorizeURL := "https://connect.q58.club/oauth/authorize"

	if clientID == "" || redirectURI == "" || authorizeURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OAuth configuration not found"})
		return
	}

	// 构建授权 URL
	data := url.Values{}
	data.Set("response_type", "code")
	data.Set("client_id", clientID)
	data.Set("redirect_uri", redirectURI)
	data.Set("scope", "read_profile")

	authURL := authorizeURL + "?" + data.Encode()

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

	redirectURI := os.Getenv("OAUTH_REDIRECT_URI")
	parsedURL, err := url.Parse(redirectURI)
	if err != nil || parsedURL.Host == "" {
		// 如果无法解析重定向URI，使用默认域名
		c.SetCookie("session", "", -1, "/", "localhost", true, true)
	} else {
		c.SetCookie("session", "", -1, "/", parsedURL.Host, true, true)
	}
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
	tokenURL := "https://connect.q58.club/api/oauth/access_token"
	clientID := os.Getenv("OAUTH_CLIENT_ID")
	clientSecret := os.Getenv("OAUTH_CLIENT_SECRET")
	redirectURI := os.Getenv("OAUTH_REDIRECT_URI")

	// 构建请求体
	data := url.Values{}
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	// 创建请求
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token request"})
		return
	}

	// 设置 Basic Auth
	req.SetBasicAuth(clientID, clientSecret)

	// 设置请求头
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
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
	userURL := "https://connect.q58.club/api/oauth/user"
	req, err = http.NewRequest("GET", userURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user info request"})
		return
	}

	req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	client = &http.Client{}
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
	cfg := c.MustGet("config").(*config.Config)

	// 检查用户是否存在
	var user models.User
	err = db.QueryRow("SELECT id, username, email, role FROM user WHERE email = ?", userInfo.Email).Scan(
		&user.ID, &user.Username, &user.Email, &user.Role)

	// 检查用户是否在管理员列表中
	isAdmin := false
	for _, adminUsername := range cfg.AdminUsernames {
		if adminUsername == userInfo.Username {
			isAdmin = true
			break
		}
	}

	role := "user"
	if isAdmin {
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
		// 更新现有用户信息
		_, err = db.Exec("UPDATE user SET username = ?, role = ? WHERE id = ?",
			userInfo.Username, role, user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user info"})
			return
		}
		user.Username = userInfo.Username
		user.Role = role
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
	// redirectURI := os.Getenv("OAUTH_REDIRECT_URI")
	parsedURL, err := url.Parse(redirectURI)
	if err != nil || parsedURL.Host == "" {
		// 如果无法解析重定向URI，使用默认域名
		c.SetCookie("session", sessionID, int(24*time.Hour.Seconds()), "/", "localhost", true, true)
	} else {
		c.SetCookie("session", sessionID, int(24*time.Hour.Seconds()), "/", parsedURL.Host, true, true)
	}

	// 重定向到前端
	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	c.Redirect(http.StatusTemporaryRedirect, baseURL)
}
