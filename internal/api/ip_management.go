package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/webappfirewall/waf/internal/database"
	"github.com/webappfirewall/waf/internal/models"
)

func ListIPs(c *gin.Context) {
	var filter models.IPFilter
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

	query := database.DB.Model(&models.IPEntry{}).Preload("Creator")

	if filter.ListType != "" {
		query = query.Where("list_type = ?", filter.ListType)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.Search != "" {
		query = query.Where("ip_address LIKE ? OR reason LIKE ?", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	var total int64
	query.Count(&total)

	var entries []models.IPEntry
	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Order("created_at DESC").
		Offset(offset).Limit(filter.PageSize).
		Find(&entries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch IP entries"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  entries,
		"total": total,
		"page":  filter.Page,
		"pages": (total + int64(filter.PageSize) - 1) / int64(filter.PageSize),
	})
}

func CreateIPEntry(c *gin.Context) {
	var req models.CreateIPEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	// Check if IP already exists in the same list
	var existing models.IPEntry
	if err := database.DB.Where("ip_address = ? AND list_type = ?", req.IPAddress, req.ListType).
		First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "IP already exists in this list"})
		return
	}

	entry := models.IPEntry{
		IPAddress: req.IPAddress,
		ListType:  req.ListType,
		Reason:    req.Reason,
		IsActive:  true,
		ExpireAt:  req.ExpireAt,
		CreatedBy: userID.(uint),
	}

	if err := database.DB.Create(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create IP entry"})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

func UpdateIPEntry(c *gin.Context) {
	id := c.Param("id")
	var entry models.IPEntry
	if err := database.DB.First(&entry, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "IP entry not found"})
		return
	}

	var req models.UpdateIPEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Reason != "" {
		entry.Reason = req.Reason
	}
	if req.IsActive != nil {
		entry.IsActive = *req.IsActive
	}
	if req.ExpireAt != nil {
		entry.ExpireAt = req.ExpireAt
	}

	if err := database.DB.Save(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update IP entry"})
		return
	}

	c.JSON(http.StatusOK, entry)
}

func DeleteIPEntry(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&models.IPEntry{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete IP entry"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "IP entry deleted"})
}

// BulkImportIPs allows importing multiple IPs at once
func BulkImportIPs(c *gin.Context) {
	var req struct {
		IPs      []string           `json:"ips" binding:"required"`
		ListType models.IPListType `json:"list_type" binding:"required,oneof=blacklist whitelist"`
		Reason   string            `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	created := 0
	skipped := 0

	for _, ip := range req.IPs {
		// Skip if already exists
		var existing models.IPEntry
		if err := database.DB.Where("ip_address = ? AND list_type = ?", ip, req.ListType).
			First(&existing).Error; err == nil {
			skipped++
			continue
		}

		entry := models.IPEntry{
			IPAddress: ip,
			ListType:  req.ListType,
			Reason:    req.Reason,
			IsActive:  true,
			CreatedBy: userID.(uint),
		}

		if err := database.DB.Create(&entry).Error; err != nil {
			skipped++
			continue
		}
		created++
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Bulk import completed",
		"created": created,
		"skipped": skipped,
	})
}

// AutoExpireIPs checks and deactivates expired IP entries
func AutoExpireIPs() {
	now := time.Now()
	database.DB.Model(&models.IPEntry{}).
		Where("is_active = true AND expire_at IS NOT NULL AND expire_at < ?", now).
		Update("is_active", false)
}
