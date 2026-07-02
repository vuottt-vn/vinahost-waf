package models

import (
	"time"

	"gorm.io/gorm"
)

type IPListType string

const (
	IPListBlacklist IPListType = "blacklist"
	IPListWhitelist IPListType = "whitelist"
)

type IPEntry struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	IPAddress   string         `gorm:"size:45;uniqueIndex;not null" json:"ip_address"`
	ListType    IPListType     `gorm:"size:16;not null;index" json:"list_type"`
	Reason      string         `gorm:"size:512" json:"reason"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	ExpireAt    *time.Time     `json:"expire_at,omitempty"` // nil = never expires
	CreatedBy   uint           `json:"created_by"`
	Creator     *User          `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type CreateIPEntryRequest struct {
	IPAddress string     `json:"ip_address" binding:"required,ip"`
	ListType  IPListType `json:"list_type" binding:"required,oneof=blacklist whitelist"`
	Reason    string     `json:"reason" binding:"max=512"`
	ExpireAt  *time.Time `json:"expire_at"`
}

type UpdateIPEntryRequest struct {
	Reason   string     `json:"reason" binding:"max=512"`
	IsActive *bool      `json:"is_active"`
	ExpireAt *time.Time `json:"expire_at"`
}

type IPFilter struct {
	ListType string `form:"list_type"`
	IsActive *bool  `form:"is_active"`
	Search   string `form:"search"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=50"`
}
