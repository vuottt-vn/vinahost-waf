package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/webappfirewall/waf/internal/auth"
	"github.com/webappfirewall/waf/internal/config"
	"github.com/webappfirewall/waf/internal/database"
	"github.com/webappfirewall/waf/internal/engine"
	"github.com/webappfirewall/waf/internal/models"
	"github.com/webappfirewall/waf/internal/proxy"
	"github.com/webappfirewall/waf/internal/ratelimit"
)

func main() {
	cfg := config.Load()

	// Initialize auth
	auth.Init(cfg.JWT.Secret)

	// Initialize database
	if err := database.Init(&cfg.Database); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Load custom rules from database
	var rules []models.Rule
	database.DB.Where("is_enabled = ?", true).Find(&rules)
	log.Printf("Loaded %d custom rules from database", len(rules))

	// Initialize WAF engine
	wafEngine, err := engine.NewEngine(&cfg.WAF)
	if err != nil {
		log.Fatalf("Failed to initialize WAF engine: %v", err)
	}

	// Load rules into engine
	if cfg.WAF.CRSEnabled {
		log.Printf("[WAF] Attempting to load with OWASP CRS...")
		if err := wafEngine.ReloadWithCRS(rules); err != nil {
			log.Printf("[WAF] Warning: Failed to load with CRS, falling back to basic mode: %v", err)
			if err := wafEngine.Reload(rules); err != nil {
				log.Fatalf("Failed to load rules into WAF engine: %v", err)
			}
		}
	} else {
		if err := wafEngine.Reload(rules); err != nil {
			log.Fatalf("Failed to load rules into WAF engine: %v", err)
		}
	}

	// Initialize challenge manager
	challengeMgr := engine.NewChallengeManager(cfg.Challenge.TokenTTL, cfg.Challenge.DifficultyLevel)

	// Initialize rate limiter
	var rateLimiter *ratelimit.RateLimiter
	if cfg.RateLimit.Enabled {
		rateLimiter = ratelimit.NewRateLimiter(
			cfg.RateLimit.MaxRequests,
			time.Duration(cfg.RateLimit.WindowSeconds)*time.Second,
		)
		log.Printf("[WAF] Rate limiter enabled: %d requests per %d seconds", 
			cfg.RateLimit.MaxRequests, cfg.RateLimit.WindowSeconds)
	} else {
		log.Printf("[WAF] Rate limiter disabled")
	}

	// Initialize proxy
	p := proxy.NewProxy(wafEngine, challengeMgr, rateLimiter)

	// Load proxy targets
	if err := p.LoadTargets(); err != nil {
		log.Printf("Warning: Failed to load proxy targets: %v", err)
	}

	addr := fmt.Sprintf(":%d", cfg.Server.ProxyPort)
	log.Printf("WAF Proxy starting on %s", addr)
	if err := http.ListenAndServe(addr, p); err != nil {
		log.Fatalf("Failed to start WAF proxy: %v", err)
	}
}
