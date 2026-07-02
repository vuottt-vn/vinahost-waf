package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/webappfirewall/waf/internal/database"
	"github.com/webappfirewall/waf/internal/models"
)

func ListLogs(c *gin.Context) {
	var filter models.AuditLogFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 200 {
		filter.PageSize = 50
	}

	query := database.DB.Model(&models.AuditLog{})

	if filter.ClientIP != "" {
		query = query.Where("client_ip = ?", filter.ClientIP)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.Method != "" {
		query = query.Where("method = ?", filter.Method)
	}
	if filter.StatusCode > 0 {
		query = query.Where("status_code = ?", filter.StatusCode)
	}
	if filter.StartDate != "" {
		if t, err := time.Parse("2006-01-02", filter.StartDate); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if filter.EndDate != "" {
		if t, err := time.Parse("2006-01-02", filter.EndDate); err == nil {
			query = query.Where("created_at <= ?", t.Add(24*time.Hour))
		}
	}

	var total int64
	query.Count(&total)

	var logs []models.AuditLog
	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Order("created_at DESC").
		Offset(offset).Limit(filter.PageSize).
		Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  logs,
		"total": total,
		"page":  filter.Page,
		"pages": (total + int64(filter.PageSize) - 1) / int64(filter.PageSize),
	})
}

func GetLog(c *gin.Context) {
	id := c.Param("id")
	var log models.AuditLog
	if err := database.DB.First(&log, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Log not found"})
		return
	}
	c.JSON(http.StatusOK, log)
}

func ClearLogs(c *gin.Context) {
	days := 30 // default
	if d := c.Query("days"); d != "" {
		// try to parse
		if _, err := time.ParseDuration(d + "h"); err == nil {
			// use parsed value
		}
	}

	threshold := time.Now().AddDate(0, 0, -days)
	result := database.DB.Where("created_at < ?", threshold).Delete(&models.AuditLog{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logs cleared",
		"deleted": result.RowsAffected,
	})
}
