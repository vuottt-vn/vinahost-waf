package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/webappfirewall/waf/internal/api"
	"github.com/webappfirewall/waf/internal/auth"
	"github.com/webappfirewall/waf/internal/config"
	"github.com/webappfirewall/waf/internal/database"
	"github.com/webappfirewall/waf/internal/engine"
	"github.com/webappfirewall/waf/internal/models"
	"github.com/webappfirewall/waf/internal/ratelimit"
)

func main() {
	cfg := config.Load()

	// Initialize auth
	auth.Init(cfg.JWT.Secret)

	// Initialize database
	if err := database.Init(&cfg.Database); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Seed default admin
	api.SeedAdmin(cfg)

	// Initialize WAF engine (for rule validation and reload capability)
	wafEngine, err := engine.NewEngine(&cfg.WAF)
	if err != nil {
		log.Fatalf("Failed to initialize WAF engine: %v", err)
	}

	// Set up engine reloader
	api.WAFEngineReloader = func() error {
		var rules []models.Rule
		database.DB.Where("is_enabled = ?", true).Find(&rules)
		return wafEngine.Reload(rules)
	}

	// Initialize rate limiter for API
	if cfg.RateLimit.Enabled {
		api.GlobalRateLimiter = ratelimit.NewRateLimiter(
			cfg.RateLimit.MaxRequests,
			time.Duration(cfg.RateLimit.WindowSeconds)*time.Second,
		)
		log.Printf("[API] Rate limiter enabled: %d requests per %d seconds",
			cfg.RateLimit.MaxRequests, cfg.RateLimit.WindowSeconds)
	}

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(api.CORSMiddleware())

	// Auth routes (public)
	authHandler := api.NewAuthHandler(cfg)
	authRoutes := r.Group("/api/auth")
	{
		authRoutes.POST("/login", authHandler.Login)
		authRoutes.POST("/refresh", authHandler.RefreshToken)
	}

	// Protected routes
	protected := r.Group("/api")
	protected.Use(api.AuthMiddleware())
	{
		protected.GET("/auth/me", authHandler.GetMe)

		// Dashboard (all users)
		protected.GET("/dashboard/stats", api.GetDashboardStats)
		protected.GET("/dashboard/realtime", api.GetRealtimeStats)

		// Rules (read for users, write for admins)
		protected.GET("/rules", api.ListRules)

		// Logs (read for users, clear for admins)
		protected.GET("/logs", api.ListLogs)
		protected.GET("/logs/:id", api.GetLog)

		// Targets (read for users)
		protected.GET("/targets", api.ListTargets)

		// IP Management (read for users)
		protected.GET("/ips", api.ListIPs)
		protected.GET("/rate-limit/stats", api.GetRateLimitStats)
		protected.GET("/rate-limit/:ip", api.GetIPRateLimit)

		// Settings (read for users)
		protected.GET("/settings", api.GetSettings)
	}

	// Admin-only routes
	admin := r.Group("/api")
	admin.Use(api.AuthMiddleware(), api.AdminMiddleware())
	{
		// Users management
		admin.GET("/users", api.ListUsers)
		admin.POST("/users", api.CreateUser)
		admin.PUT("/users/:id", api.UpdateUser)
		admin.DELETE("/users/:id", api.DeleteUser)

		// Rules management
		admin.POST("/rules", api.CreateRule)
		admin.PUT("/rules/:id", api.UpdateRule)
		admin.DELETE("/rules/:id", api.DeleteRule)
		admin.POST("/rules/:id/toggle", api.ToggleRule)
		admin.POST("/rules/reload", api.ReloadEngine)

		// Targets management
		admin.POST("/targets", api.CreateTarget)
		admin.PUT("/targets/:id", api.UpdateTarget)
		admin.DELETE("/targets/:id", api.DeleteTarget)

		// Logs management
		admin.DELETE("/logs", api.ClearLogs)

		// IP Management
		admin.POST("/ips", api.CreateIPEntry)
		admin.PUT("/ips/:id", api.UpdateIPEntry)
		admin.DELETE("/ips/:id", api.DeleteIPEntry)
		admin.POST("/ips/bulk-import", api.BulkImportIPs)
		admin.POST("/rate-limit/reset/:ip", api.ResetIPRateLimit)
		admin.POST("/rate-limit/reset-all", api.ResetAllRateLimits)

		// Settings
		admin.PUT("/settings", api.UpdateSettings)
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	addr := fmt.Sprintf(":%d", cfg.Server.APIPort)
	log.Printf("API Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start API server: %v", err)
	}
}
