package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/webappfirewall/waf/internal/database"
	"github.com/webappfirewall/waf/internal/models"
)

// WAFEngineReloader is a function to reload the WAF engine
var WAFEngineReloader func() error

func ListRules(c *gin.Context) {
	var rules []models.Rule
	q := database.DB.Order("priority DESC, id DESC")
	if search := c.Query("search"); search != "" {
		q = q.Where("name LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if err := q.Preload("Creator").Find(&rules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rules"})
		return
	}
	c.JSON(http.StatusOK, rules)
}

func CreateRule(c *gin.Context) {
	var req models.CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	isEnabled := true
	if req.IsEnabled != nil {
		isEnabled = *req.IsEnabled
	}

	rule := models.Rule{
		Name:        req.Name,
		Description: req.Description,
		SecRule:     req.SecRule,
		IsEnabled:   isEnabled,
		Priority:    req.Priority,
		CreatedBy:   userID.(uint),
	}

	if err := database.DB.Create(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create rule"})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

func UpdateRule(c *gin.Context) {
	id := c.Param("id")
	var rule models.Rule
	if err := database.DB.First(&rule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}

	var req models.UpdateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		rule.Name = req.Name
	}
	if req.Description != "" {
		rule.Description = req.Description
	}
	if req.SecRule != "" {
		rule.SecRule = req.SecRule
	}
	if req.IsEnabled != nil {
		rule.IsEnabled = *req.IsEnabled
	}
	if req.Priority != 0 {
		rule.Priority = req.Priority
	}

	if err := database.DB.Save(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update rule"})
		return
	}

	c.JSON(http.StatusOK, rule)
}

func DeleteRule(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&models.Rule{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete rule"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Rule deleted"})
}

func ToggleRule(c *gin.Context) {
	id := c.Param("id")
	var rule models.Rule
	if err := database.DB.First(&rule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}

	rule.IsEnabled = !rule.IsEnabled
	if err := database.DB.Save(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle rule"})
		return
	}

	c.JSON(http.StatusOK, rule)
}

func ReloadEngine(c *gin.Context) {
	if WAFEngineReloader != nil {
		if err := WAFEngineReloader(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload WAF engine: " + err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "WAF engine reloaded"})
}

func ListTargets(c *gin.Context) {
	var targets []models.ProxyTarget
	if err := database.DB.Preload("Creator").Find(&targets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch targets"})
		return
	}
	c.JSON(http.StatusOK, targets)
}

func CreateTarget(c *gin.Context) {
	var req models.CreateTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	isEnabled := true
	if req.IsEnabled != nil {
		isEnabled = *req.IsEnabled
	}

	target := models.ProxyTarget{
		Name:        req.Name,
		UpstreamURL: req.UpstreamURL,
		IsEnabled:   isEnabled,
		CreatedBy:   userID.(uint),
	}

	if err := database.DB.Create(&target).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create target"})
		return
	}

	c.JSON(http.StatusCreated, target)
}

func UpdateTarget(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	var target models.ProxyTarget
	if err := database.DB.First(&target, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Target not found"})
		return
	}

	var req models.UpdateTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		target.Name = req.Name
	}
	if req.UpstreamURL != "" {
		target.UpstreamURL = req.UpstreamURL
	}
	if req.IsEnabled != nil {
		target.IsEnabled = *req.IsEnabled
	}

	if err := database.DB.Save(&target).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update target"})
		return
	}

	c.JSON(http.StatusOK, target)
}

func DeleteTarget(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&models.ProxyTarget{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete target"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Target deleted"})
}
