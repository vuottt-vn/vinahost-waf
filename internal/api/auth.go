package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/webappfirewall/waf/internal/auth"
	"github.com/webappfirewall/waf/internal/config"
	"github.com/webappfirewall/waf/internal/database"
	"github.com/webappfirewall/waf/internal/models"
)

type AuthHandler struct {
	cfg *config.Config
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{cfg: cfg}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "Account is disabled"})
		return
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	accessToken, err := auth.GenerateAccessToken(user.ID, user.Username, string(user.Role), h.cfg.JWT.AccessExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID, user.Username, string(user.Role), h.cfg.JWT.RefreshExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    h.cfg.JWT.AccessExpiry * 60,
		User:         &user,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := auth.ValidateToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	accessToken, err := auth.GenerateAccessToken(claims.UserID, claims.Username, claims.Role, h.cfg.JWT.AccessExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken, err := auth.GenerateRefreshToken(claims.UserID, claims.Username, claims.Role, h.cfg.JWT.RefreshExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    h.cfg.JWT.AccessExpiry * 60,
	})
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// SeedAdmin creates the default admin user if no admin exists
func SeedAdmin(cfg *config.Config) {
	var count int64
	database.DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		return
	}

	hash, err := auth.HashPassword("admin123")
	if err != nil {
		log.Printf("Failed to hash default admin password: %v", err)
		return
	}

	admin := models.User{
		Username:     "admin",
		Email:        "admin@waf.local",
		PasswordHash: hash,
		Role:         models.RoleAdmin,
		IsActive:     true,
	}

	if err := database.DB.Create(&admin).Error; err != nil {
		log.Printf("Failed to create default admin: %v", err)
		return
	}

	log.Println("[Auth] Default admin user created: admin/admin123")
}
