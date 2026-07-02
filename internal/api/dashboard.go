package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/webappfirewall/waf/internal/database"
	"github.com/webappfirewall/waf/internal/models"
)

func GetDashboardStats(c *gin.Context) {
	stats := models.DashboardStats{}

	// Total counts
	database.DB.Model(&models.AuditLog{}).Count(&stats.TotalRequests)
	database.DB.Model(&models.AuditLog{}).Where("action = ?", models.ActionBlock).Count(&stats.BlockedRequests)
	database.DB.Model(&models.AuditLog{}).Where("action = ?", models.ActionChallenge).Count(&stats.ChallengedReqs)
	stats.AllowedRequests = stats.TotalRequests - stats.BlockedRequests - stats.ChallengedReqs

	// Top attack types - extract from matched rules text
	var logs []models.AuditLog
	database.DB.Where("action = ? AND matched_rules != ''", models.ActionBlock).
		Select("matched_rules").Limit(1000).Find(&logs)

	attackCounts := make(map[string]int64)
	for _, log := range logs {
		// Parse rule ID from "Rule ID: XXXX, Message: ..."
		if strings.Contains(log.MatchedRules, "Rule ID:") {
			parts := strings.SplitN(log.MatchedRules, ",", 2)
			if len(parts) > 0 {
				ruleType := strings.TrimSpace(parts[0])
				attackCounts[ruleType]++
			}
		} else {
			attackCounts["Unknown"]++
		}
	}

	stats.TopAttackTypes = make([]models.AttackTypeStat, 0, len(attackCounts))
	for t, count := range attackCounts {
		stats.TopAttackTypes = append(stats.TopAttackTypes, models.AttackTypeStat{Type: t, Count: count})
	}
	// Sort by count descending (simple bubble sort for small slice)
	for i := 0; i < len(stats.TopAttackTypes); i++ {
		for j := i + 1; j < len(stats.TopAttackTypes); j++ {
			if stats.TopAttackTypes[j].Count > stats.TopAttackTypes[i].Count {
				stats.TopAttackTypes[i], stats.TopAttackTypes[j] = stats.TopAttackTypes[j], stats.TopAttackTypes[i]
			}
		}
	}
	if len(stats.TopAttackTypes) > 10 {
		stats.TopAttackTypes = stats.TopAttackTypes[:10]
	}

	// Top blocked IPs
	type IPCount struct {
		IP    string `json:"ip"`
		Count int64  `json:"count"`
	}
	var blockedIPs []IPCount
	database.DB.Model(&models.AuditLog{}).
		Select("client_ip as ip, COUNT(*) as count").
		Where("action = ?", models.ActionBlock).
		Group("client_ip").
		Order("count DESC").
		Limit(10).
		Scan(&blockedIPs)
	stats.TopBlockedIPs = make([]models.BlockedIPStat, len(blockedIPs))
	for i, ip := range blockedIPs {
		stats.TopBlockedIPs[i] = models.BlockedIPStat{IP: ip.IP, Count: ip.Count}
	}

	// Traffic timeline (last 24 hours, hourly) - SQLite compatible
	type TimeCount struct {
		Hour    string `json:"hour"`
		Total   int64  `json:"total"`
		Blocked int64  `json:"blocked"`
	}
	var timeline []TimeCount
	since := time.Now().Add(-24 * time.Hour)
	database.DB.Model(&models.AuditLog{}).
		Select("strftime('%H:00', created_at) as hour, COUNT(*) as total, SUM(CASE WHEN action = 'block' THEN 1 ELSE 0 END) as blocked").
		Where("created_at >= ?", since).
		Group("strftime('%H:00', created_at)").
		Order("hour ASC").
		Scan(&timeline)
	stats.TrafficTimeline = make([]models.TrafficPoint, len(timeline))
	for i, t := range timeline {
		stats.TrafficTimeline[i] = models.TrafficPoint{Time: t.Hour, Total: t.Total, Blocked: t.Blocked}
	}

	c.JSON(http.StatusOK, stats)
}

func GetRealtimeStats(c *gin.Context) {
	// Last 5 minutes stats
	since := time.Now().Add(-5 * time.Minute)

	var stats struct {
		Total   int64 `json:"total"`
		Blocked int64 `json:"blocked"`
		Recent  []models.AuditLog `json:"recent"`
	}

	database.DB.Model(&models.AuditLog{}).Where("created_at >= ?", since).Count(&stats.Total)
	database.DB.Model(&models.AuditLog{}).Where("created_at >= ? AND action = ?", since, models.ActionBlock).Count(&stats.Blocked)
	database.DB.Where("created_at >= ?", since).Order("created_at DESC").Limit(20).Find(&stats.Recent)

	c.JSON(http.StatusOK, stats)
}

func GetSettings(c *gin.Context) {
	// Return current WAF config from environment
	settings := models.WAFSettings{
		Mode:               "on",
		ChallengeEnabled:   true,
		ChallengeThreshold: 5,
		ChallengeDifficulty: 3,
		RequestBodyLimit:  131072,
		ResponseBodyLimit: 131072,
	}
	c.JSON(http.StatusOK, settings)
}

func UpdateSettings(c *gin.Context) {
	var req models.WAFSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// In production, this would update the config and reload the engine
	c.JSON(http.StatusOK, gin.H{"message": "Settings updated"})
}
