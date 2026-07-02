package models

import "time"

type ActionType string

const (
	ActionAllow     ActionType = "allow"
	ActionBlock     ActionType = "block"
	ActionChallenge ActionType = "challenge"
)

type AuditLog struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	TransactionID string     `gorm:"size:64;index" json:"transaction_id"`
	ClientIP      string     `gorm:"size:45;index" json:"client_ip"`
	ServerIP      string     `gorm:"size:45" json:"server_ip"`
	RequestURI    string     `gorm:"size:2048" json:"request_uri"`
	Method        string     `gorm:"size:10" json:"method"`
	StatusCode    int        `json:"status_code"`
	Action        ActionType `gorm:"size:16;index" json:"action"`
	MatchedRules  string     `gorm:"type:text" json:"matched_rules"`
	AnomalyScore  int        `json:"anomaly_score"`
	RequestBody   string     `gorm:"type:text" json:"request_body,omitempty"`
	UserAgent     string     `gorm:"size:512" json:"user_agent"`
	CreatedAt     time.Time  `gorm:"index" json:"created_at"`
}

type AuditLogFilter struct {
	ClientIP   string `form:"client_ip"`
	Action     string `form:"action"`
	Method     string `form:"method"`
	StartDate  string `form:"start_date"`
	EndDate    string `form:"end_date"`
	StatusCode int    `form:"status_code"`
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"page_size,default=50"`
}

type DashboardStats struct {
	TotalRequests   int64 `json:"total_requests"`
	BlockedRequests int64 `json:"blocked_requests"`
	ChallengedReqs  int64 `json:"challenged_requests"`
	AllowedRequests int64 `json:"allowed_requests"`
	TopAttackTypes  []AttackTypeStat `json:"top_attack_types"`
	TopBlockedIPs   []BlockedIPStat  `json:"top_blocked_ips"`
	TrafficTimeline []TrafficPoint   `json:"traffic_timeline"`
}

type AttackTypeStat struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
}

type BlockedIPStat struct {
	IP    string `json:"ip"`
	Count int64  `json:"count"`
}

type TrafficPoint struct {
	Time    string `json:"time"`
	Total   int64  `json:"total"`
	Blocked int64  `json:"blocked"`
}
