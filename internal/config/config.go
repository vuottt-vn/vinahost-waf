package config

import (
	"os"
	"strconv"
)

type Config struct {
	Database  DatabaseConfig
	Server    ServerConfig
	JWT       JWTConfig
	WAF       WAFConfig
	Challenge ChallengeConfig
	RateLimit RateLimitConfig
	IP        IPConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	DBPath   string // SQLite file path
}

type ServerConfig struct {
	APIPort  int
	ProxyPort int
}

type JWTConfig struct {
	Secret          string
	AccessExpiry    int // minutes
	RefreshExpiry   int // hours
}

type WAFConfig struct {
	Mode              string // "on", "detection_only", "off"
	RequestBodyLimit  int
	ResponseBodyLimit int
	CRSEnabled        bool
	CRSPath           string
}

type ChallengeConfig struct {
	Enabled          bool
	Threshold        int // anomaly score threshold for challenge
	TokenTTL         int // minutes
	DifficultyLevel  int // 1-5
}

type RateLimitConfig struct {
	Enabled       bool
	MaxRequests   int           // Maximum requests per window
	WindowSeconds int           // Time window in seconds
	BlockDuration int           // How long to block after exceeding limit (seconds)
}

type IPConfig struct {
	BlacklistEnabled bool
	WhitelistEnabled bool
}

func Load() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "wafadmin"),
			Password: getEnv("DB_PASSWORD", "wafpassword"),
			DBName:   getEnv("DB_NAME", "webappfirewall"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			DBPath:   getEnv("DB_PATH", "waf.db"),
		},
		Server: ServerConfig{
			APIPort:   getEnvInt("API_PORT", 3001),
			ProxyPort: getEnvInt("PROXY_PORT", 8080),
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", "super-secret-key-change-in-production-32chars!"),
			AccessExpiry:  getEnvInt("JWT_ACCESS_EXPIRY", 15),
			RefreshExpiry: getEnvInt("JWT_REFRESH_EXPIRY", 168), // 7 days
		},
		WAF: WAFConfig{
			Mode:              getEnv("WAF_MODE", "on"),
			RequestBodyLimit:  getEnvInt("WAF_REQ_BODY_LIMIT", 131072),
			ResponseBodyLimit: getEnvInt("WAF_RES_BODY_LIMIT", 131072),
			CRSEnabled:        getEnv("WAF_CRS_ENABLED", "true") == "true",
			CRSPath:           getEnv("WAF_CRS_PATH", "./configs/crs"),
		},
		Challenge: ChallengeConfig{
			Enabled:         getEnv("CHALLENGE_ENABLED", "true") == "true",
			Threshold:       getEnvInt("CHALLENGE_THRESHOLD", 5),
			TokenTTL:        getEnvInt("CHALLENGE_TOKEN_TTL", 30),
			DifficultyLevel: getEnvInt("CHALLENGE_DIFFICULTY", 3),
		},
		RateLimit: RateLimitConfig{
			Enabled:       getEnv("RATE_LIMIT_ENABLED", "true") == "true",
			MaxRequests:   getEnvInt("RATE_LIMIT_MAX_REQUESTS", 100),
			WindowSeconds: getEnvInt("RATE_LIMIT_WINDOW_SECONDS", 60),
			BlockDuration: getEnvInt("RATE_LIMIT_BLOCK_DURATION", 300),
		},
		IP: IPConfig{
			BlacklistEnabled: getEnv("IP_BLACKLIST_ENABLED", "true") == "true",
			WhitelistEnabled: getEnv("IP_WHITELIST_ENABLED", "false") == "true",
		},
	}
}

func (d *DatabaseConfig) DSN() string {
	return "host=" + d.Host +
		" port=" + strconv.Itoa(d.Port) +
		" user=" + d.User +
		" password=" + d.Password +
		" dbname=" + d.DBName +
		" sslmode=" + d.SSLMode
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}
