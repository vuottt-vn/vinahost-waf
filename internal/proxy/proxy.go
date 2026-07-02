package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/webappfirewall/waf/internal/database"
	"github.com/webappfirewall/waf/internal/engine"
	"github.com/webappfirewall/waf/internal/models"
	"github.com/webappfirewall/waf/internal/ratelimit"
)

// Proxy is the WAF-aware HTTP proxy
type Proxy struct {
	wafEngine        *engine.Engine
	challengeManager *engine.ChallengeManager
	target           map[string]*models.ProxyTarget
	rateLimiter      *ratelimit.RateLimiter
	mu               sync.RWMutex
}

// NewProxy creates a new WAF proxy
func NewProxy(wafEngine *engine.Engine, challengeMgr *engine.ChallengeManager, rateLimiter *ratelimit.RateLimiter) *Proxy {
	return &Proxy{
		wafEngine:        wafEngine,
		challengeManager: challengeMgr,
		target:           make(map[string]*models.ProxyTarget),
		rateLimiter:      rateLimiter,
	}
}

// LoadTargets loads proxy targets from the database
func (p *Proxy) LoadTargets() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var targets []models.ProxyTarget
	if err := database.DB.Where("is_enabled = ?", true).Find(&targets).Error; err != nil {
		return err
	}

	p.target = make(map[string]*models.ProxyTarget)
	for i := range targets {
		p.target[targets[i].Name] = &targets[i]
		log.Printf("[Proxy] Loaded target: %s -> %s", targets[i].Name, targets[i].UpstreamURL)
	}

	return nil
}

// ServeHTTP handles incoming proxy requests
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Handle challenge verification endpoint
	if r.URL.Path == "/__waf_challenge/verify" {
		p.handleChallengeVerify(w, r)
		return
	}

	clientIP := getClientIP(r)

	// Check IP whitelist/blacklist
	if ipAction := p.checkIPFilter(clientIP, w); ipAction != "allow" {
		return
	}

	// Check rate limit
	if p.rateLimiter != nil && !p.rateLimiter.Allow(clientIP) {
		log.Printf("[Proxy] Rate limit exceeded for IP: %s", clientIP)
		p.writeRateLimitExceeded(w, clientIP)
		return
	}

	// Check WAF token bypass
	clientPort := 0
	serverIP := r.Host
	serverPort := 80

	wafToken := getWAFToken(r)
	if wafToken != "" && p.challengeManager.VerifyToken(wafToken, clientIP) {
		p.forwardRequest(w, r, startTime, models.ActionAllow, 200, 0, "")
		return
	}

	var body []byte
	if r.Body != nil {
		var err error
		body, err = io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		r.Body.Close()
	}

	result, err := p.wafEngine.ProcessTransaction(
		clientIP, clientPort,
		serverIP, serverPort,
		r.Method, r.RequestURI, r.Proto,
		r.Header, body,
	)
	if err != nil {
		log.Printf("[Proxy] WAF error: %v", err)
		p.forwardRequest(w, r, startTime, models.ActionAllow, 200, 0, "")
		return
	}

	if body != nil {
		r.Body = io.NopCloser(strings.NewReader(string(body)))
	}

	switch result.Action {
	case models.ActionBlock:
		p.logAudit(result, r, clientIP)
		writeBlockPage(w, result)
		return
	case models.ActionChallenge:
		if result.AnomalyScore >= 5 {
			p.logAudit(result, r, clientIP)
			challengeData, sessionID := p.challengeManager.GenerateChallenge(clientIP)
			engine.WriteChallengePage(w, challengeData, sessionID)
			return
		}
		p.forwardRequest(w, r, startTime, models.ActionAllow, 200, result.AnomalyScore, result.MatchedRules)
	default:
		p.forwardRequest(w, r, startTime, models.ActionAllow, 200, result.AnomalyScore, result.MatchedRules)
	}
}

func (p *Proxy) forwardRequest(w http.ResponseWriter, r *http.Request, startTime time.Time,
	action models.ActionType, statusCode int, anomalyScore int, matchedRules string) {

	upstream := p.getUpstream(r)
	if upstream == "" {
		upstream = "http://localhost:8081"
	}

	target, err := url.Parse(upstream)
	if err != nil {
		http.Error(w, "Invalid upstream URL", http.StatusBadGateway)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("[Proxy] Upstream error: %v", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
	}

	proxy.ModifyResponse = func(resp *http.Response) error {
		statusCode = resp.StatusCode
		return nil
	}

	r.Host = target.Host
	proxy.ServeHTTP(w, r)

	clientIP := getClientIP(r)
	logEntry := &models.AuditLog{
		TransactionID: fmt.Sprintf("%d", startTime.UnixNano()),
		ClientIP:      clientIP,
		ServerIP:      r.Host,
		RequestURI:    r.RequestURI,
		Method:        r.Method,
		StatusCode:    statusCode,
		Action:        action,
		MatchedRules:  matchedRules,
		AnomalyScore:  anomalyScore,
		UserAgent:     r.UserAgent(),
	}
	if err := database.DB.Create(logEntry).Error; err != nil {
		log.Printf("[Proxy] Failed to log: %v", err)
	}
}

func (p *Proxy) getUpstream(r *http.Request) string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	host := r.Host
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	for _, target := range p.target {
		if target.Name == host || target.Name == r.Host {
			return target.UpstreamURL
		}
	}

	for _, target := range p.target {
		return target.UpstreamURL
	}

	return ""
}

func (p *Proxy) handleChallengeVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, _ := io.ReadAll(r.Body)
	r.Body.Close()

	var req struct {
		SessionID string `json:"session_id"`
		Nonce     string `json:"nonce"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false, "error": "Invalid request",
		})
		return
	}

	clientIP := getClientIP(r)
	if p.challengeManager.VerifySolution(req.SessionID, req.Nonce, clientIP) {
		token := p.challengeManager.IssueToken(clientIP)

		http.SetCookie(w, &http.Cookie{
			Name:     "__waf_token",
			Value:    token,
			Path:     "/",
			MaxAge:   1800,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": true, "token": token,
		})
	} else {
		writeJSON(w, http.StatusForbidden, map[string]interface{}{
			"success": false, "error": "Challenge failed",
		})
	}
}

func (p *Proxy) logAudit(result *engine.TransactionResult, r *http.Request, clientIP string) {
	logEntry := &models.AuditLog{
		TransactionID: result.TransactionID,
		ClientIP:      clientIP,
		ServerIP:      r.Host,
		RequestURI:    r.RequestURI,
		Method:        r.Method,
		StatusCode:    result.Status,
		Action:        result.Action,
		MatchedRules:  result.MatchedRules,
		AnomalyScore:  result.AnomalyScore,
		UserAgent:     r.UserAgent(),
	}
	database.DB.Create(logEntry)
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0]
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx > 0 {
		return addr[:idx]
	}
	return addr
}

func getWAFToken(r *http.Request) string {
	if cookie, err := r.Cookie("__waf_token"); err == nil {
		return cookie.Value
	}
	if token := r.Header.Get("X-WAF-Token"); token != "" {
		return token
	}
	return ""
}

func writeBlockPage(w http.ResponseWriter, result *engine.TransactionResult) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	statusCode := result.Status
	if statusCode == 0 {
		statusCode = http.StatusForbidden
	}
	w.WriteHeader(statusCode)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Access Denied</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #0f172a; color: #e2e8f0;
            display: flex; align-items: center; justify-content: center; min-height: 100vh;
        }
        .container {
            background: #1e293b; border: 1px solid #334155;
            border-radius: 16px; padding: 48px; max-width: 480px;
            width: 90%%; text-align: center;
        }
        .icon { font-size: 64px; margin-bottom: 24px; }
        h1 { font-size: 24px; margin-bottom: 12px; color: #f1f5f9; }
        p { color: #94a3b8; margin-bottom: 16px; line-height: 1.6; }
        .detail { font-size: 12px; color: #64748b; margin-top: 24px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">&#128683;</div>
        <h1>Access Denied</h1>
        <p>Your request has been blocked by the Web Application Firewall.</p>
        <p>If you believe this is an error, please contact the administrator.</p>
        <div class="detail">Transaction ID: %s</div>
    </div>
</body>
</html>`, result.TransactionID)
	w.Write([]byte(html))
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// checkIPFilter checks if the IP is whitelisted or blacklisted
func (p *Proxy) checkIPFilter(clientIP string, w http.ResponseWriter) string {
	// Check whitelist first (if enabled)
	var whitelistCount int64
	database.DB.Model(&models.IPEntry{}).
		Where("ip_address = ? AND list_type = ? AND is_active = ?", 
			clientIP, models.IPListWhitelist, true).
		Count(&whitelistCount)
	
	if whitelistCount > 0 {
		// IP is whitelisted - bypass all checks
		return "allow"
	}

	// Check blacklist
	var blacklistEntry models.IPEntry
	if err := database.DB.Where("ip_address = ? AND list_type = ? AND is_active = ?",
		clientIP, models.IPListBlacklist, true).
		First(&blacklistEntry).Error; err == nil {
		// Check if expired
		if blacklistEntry.ExpireAt != nil && time.Now().After(*blacklistEntry.ExpireAt) {
			// Expired - deactivate
			database.DB.Model(&models.IPEntry{}).
				Where("id = ?", blacklistEntry.ID).
				Update("is_active", false)
			return "allow"
		}

		// IP is blacklisted - block request
		log.Printf("[Proxy] Blacklisted IP blocked: %s (reason: %s)", clientIP, blacklistEntry.Reason)
		p.writeBlockedPage(w, "IP Address Blacklisted", fmt.Sprintf("Your IP (%s) has been blocked. Reason: %s", clientIP, blacklistEntry.Reason))
		return "block"
	}

	return "allow"
}

// writeRateLimitExceeded sends a rate limit exceeded response
func (p *Proxy) writeRateLimitExceeded(w http.ResponseWriter, clientIP string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Retry-After", "60")
	w.WriteHeader(http.StatusTooManyRequests)

	retryAfter := p.rateLimiter.GetRetryAfter(clientIP)
	retrySeconds := int(retryAfter.Seconds())
	if retrySeconds == 0 {
		retrySeconds = 60
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Rate Limit Exceeded</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #0f172a; color: #e2e8f0;
            display: flex; align-items: center; justify-content: center; min-height: 100vh;
        }
        .container {
            background: #1e293b; border: 1px solid #334155;
            border-radius: 16px; padding: 48px; max-width: 480px;
            width: 90%%; text-align: center;
        }
        .icon { font-size: 64px; margin-bottom: 24px; }
        h1 { font-size: 24px; margin-bottom: 12px; color: #f1f5f9; }
        p { color: #94a3b8; margin-bottom: 16px; line-height: 1.6; }
        .detail { font-size: 12px; color: #64748b; margin-top: 24px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">⏱️</div>
        <h1>Rate Limit Exceeded</h1>
        <p>You have made too many requests. Please wait before trying again.</p>
        <p>Retry after %d seconds.</p>
        <div class="detail">IP: %s</div>
    </div>
</body>
</html>`, retrySeconds, clientIP)
	w.Write([]byte(html))
}

// writeBlockedPage sends a generic blocked page
func (p *Proxy) writeBlockedPage(w http.ResponseWriter, title, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #0f172a; color: #e2e8f0;
            display: flex; align-items: center; justify-content: center; min-height: 100vh;
        }
        .container {
            background: #1e293b; border: 1px solid #334155;
            border-radius: 16px; padding: 48px; max-width: 480px;
            width: 90%%; text-align: center;
        }
        .icon { font-size: 64px; margin-bottom: 24px; }
        h1 { font-size: 24px; margin-bottom: 12px; color: #f1f5f9; }
        p { color: #94a3b8; margin-bottom: 16px; line-height: 1.6; }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">🚫</div>
        <h1>%s</h1>
        <p>%s</p>
    </div>
</body>
</html>`, title, title, message)
	w.Write([]byte(html))
}
