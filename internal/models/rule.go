package models

import (
	"time"

	"gorm.io/gorm"
)

type Rule struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:128;not null" json:"name"`
	Description string         `gorm:"size:512" json:"description"`
	SecRule     string         `gorm:"type:text;not null" json:"sec_rule"`
	IsEnabled   bool           `gorm:"default:true" json:"is_enabled"`
	Priority    int            `gorm:"default:100" json:"priority"`
	CreatedBy   uint           `json:"created_by"`
	Creator     *User          `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type CreateRuleRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=128"`
	Description string `json:"description" binding:"max=512"`
	SecRule     string `json:"sec_rule" binding:"required"`
	IsEnabled   *bool  `json:"is_enabled"`
	Priority    int    `json:"priority"`
}

type UpdateRuleRequest struct {
	Name        string `json:"name" binding:"omitempty,min=1,max=128"`
	Description string `json:"description" binding:"max=512"`
	SecRule     string `json:"sec_rule" binding:"omitempty"`
	IsEnabled   *bool  `json:"is_enabled"`
	Priority    int    `json:"priority"`
}
