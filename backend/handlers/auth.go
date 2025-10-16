package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"aimodels-prices/database"
	"aimodels-prices/models"
)

// generateSessionID 生成随机会话ID
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

	var session models.Session
	if err := database.DB.Preload("User").Where("id = ?", cookie).First(&session).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	if session.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired"})
		return
	}

	c.Set("user", &session.User)
	c.JSON(http.StatusOK, gin.H{
		"user": session.User,
	})
}

func Login(c *gin.Context) {
	// 生产环境使用 OAuth 2.0
	clientID := os.Getenv("OAUTH_CLIENT_ID")
	redirectURI := os.Getenv("OAUTH_REDIRECT_URI")
	authorizeURL := "https://connect.czl.net/oauth2/authorize" // 固定授权URL

	if clientID == "" || redirectURI == "" {
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
		// 删除会话
		database.DB.Where("id = ?", cookie).Delete(&models.Session{})
	}

	c.SetCookie("session", "", -1, "/", "ai-prices.sunai.net", true, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func GetUser(c *gin.Context) {
	cookie, err := c.Cookie("session")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not logged in"})
		return
	}

	var session models.Session
	if err := database.DB.Preload("User").Where("id = ?", cookie).First(&session).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	if session.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": session.User,
	})
}

func AuthCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing authorization code"})
		return
	}

	// 获取访问令牌
	tokenURL := "https://connect.czl.net/api/oauth2/token" // 固定令牌URL
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
	userURL := "https://connect.czl.net/api/oauth2/userinfo" // 固定用户信息URL
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
		ID        int    `json:"id"`
		Email     string `json:"email"`
		Username  string `json:"username"`
		Avatar    string `json:"avatar"`
		Name      string `json:"name"`
		Groups    string `json:"groups"`
		Upstreams []struct {
			ID               int                    `json:"id"`
			UpstreamID       int                    `json:"upstream_id"`
			UpstreamName     string                 `json:"upstream_name"`
			UpstreamType     string                 `json:"upstream_type"`
			UpstreamIcon     string                 `json:"upstream_icon"`
			UpstreamUserID   string                 `json:"upstream_user_id"`
			UpstreamUsername string                 `json:"upstream_username"`
			UpstreamEmail    string                 `json:"upstream_email"`
			UpstreamAvatar   string                 `json:"upstream_avatar"`
			ProviderData     map[string]interface{} `json:"provider_data"`
		} `json:"upstreams"`
	}

	if err := json.NewDecoder(userResp.Body).Decode(&userInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user info"})
		return
	}

	// 添加调试日志
	fmt.Printf("收到OAuth用户信息: ID=%v, Username=%s, Email=%s, Groups=%s\n",
		userInfo.ID, userInfo.Username, userInfo.Email, userInfo.Groups)

	// 检查用户是否存在
	var user models.User
	result := database.DB.Where("email = ?", userInfo.Email).First(&user)

	groups := userInfo.Groups
	if groups == "" {
		groups = "t0"
	}

	if result.Error != nil {
		// 创建新用户
		user = models.User{
			Username: userInfo.Username,
			Email:    userInfo.Email,
			Groups:   groups,
		}
		if err := database.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
		fmt.Printf("创建新用户: %s, Groups=%s\n", userInfo.Username, groups)
	} else {
		// 每次登录都更新用户的权限组
		user.Groups = groups
		if err := database.DB.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}
		fmt.Printf("更新用户权限: %s, Groups=%s\n", userInfo.Username, groups)
	}

	// 创建会话
	sessionID := generateSessionID()
	expiresAt := time.Now().Add(24 * time.Hour)
	session := models.Session{
		ID:        sessionID,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
	}
	if err := database.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// 设置 cookie
	c.SetCookie("session", sessionID, int(24*time.Hour.Seconds()), "/", "ai-prices.sunai.net", true, true)

	// 重定向到前端
	c.Redirect(http.StatusTemporaryRedirect, "https://ai-prices.sunai.net")
}
