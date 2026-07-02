package models

import (
	"time"

	"gorm.io/gorm"
)

type ProxyTarget struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:128;not null" json:"name"`
	UpstreamURL string         `gorm:"size:512;not null" json:"upstream_url"`
	IsEnabled   bool           `gorm:"default:true" json:"is_enabled"`
	CreatedBy   uint           `json:"created_by"`
	Creator     *User          `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type CreateTargetRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=128"`
	UpstreamURL string `json:"upstream_url" binding:"required,url"`
	IsEnabled   *bool  `json:"is_enabled"`
}

type UpdateTargetRequest struct {
	Name        string `json:"name" binding:"omitempty,min=1,max=128"`
	UpstreamURL string `json:"upstream_url" binding:"omitempty,url"`
	IsEnabled   *bool  `json:"is_enabled"`
}

type WAFSettings struct {
	Mode              string `json:"mode"`
	ChallengeEnabled  bool   `json:"challenge_enabled"`
	ChallengeThreshold int   `json:"challenge_threshold"`
	ChallengeDifficulty int  `json:"challenge_difficulty"`
	RequestBodyLimit  int    `json:"request_body_limit"`
	ResponseBodyLimit int    `json:"response_body_limit"`
}
