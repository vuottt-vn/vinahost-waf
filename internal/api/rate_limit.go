package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/webappfirewall/waf/internal/ratelimit"
)

// RateLimiter instance (set during initialization)
var GlobalRateLimiter *ratelimit.RateLimiter

// GetRateLimitStats returns current rate limit statistics
func GetRateLimitStats(c *gin.Context) {
	if GlobalRateLimiter == nil {
		c.JSON(http.StatusOK, gin.H{
			"enabled": false,
			"stats":   make(map[string]int),
		})
		return
	}

	stats := GlobalRateLimiter.GetStats()
	c.JSON(http.StatusOK, gin.H{
		"enabled": true,
		"stats":   stats,
		"total":   len(stats),
	})
}

// GetIPRateLimit returns rate limit info for a specific IP
func GetIPRateLimit(c *gin.Context) {
	ip := c.Param("ip")
	
	if GlobalRateLimiter == nil {
		c.JSON(http.StatusOK, gin.H{
			"ip":        ip,
			"enabled":   false,
			"remaining": 0,
		})
		return
	}

	remaining := GlobalRateLimiter.GetRemaining(ip)
	retryAfter := GlobalRateLimiter.GetRetryAfter(ip)

	c.JSON(http.StatusOK, gin.H{
		"ip":         ip,
		"enabled":    true,
		"remaining":  remaining,
		"retry_after_seconds": retryAfter.Seconds(),
	})
}

// ResetIPRateLimit resets rate limit for a specific IP
func ResetIPRateLimit(c *gin.Context) {
	ip := c.Param("ip")
	
	if GlobalRateLimiter == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Rate limiter not enabled"})
		return
	}

	GlobalRateLimiter.Reset(ip)
	c.JSON(http.StatusOK, gin.H{
		"message": "Rate limit reset for IP",
		"ip":      ip,
	})
}

// ResetAllRateLimits resets all rate limits
func ResetAllRateLimits(c *gin.Context) {
	if GlobalRateLimiter == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Rate limiter not enabled"})
		return
	}

	GlobalRateLimiter.ResetAll()
	c.JSON(http.StatusOK, gin.H{
		"message": "All rate limits reset",
	})
}
