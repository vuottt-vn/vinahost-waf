package engine

import (
	"fmt"
	"log"
	"sync"

	"github.com/corazawaf/coraza/v3"
	"github.com/corazawaf/coraza/v3/types"
	"github.com/webappfirewall/waf/internal/config"
	"github.com/webappfirewall/waf/internal/crs"
	"github.com/webappfirewall/waf/internal/models"
)

// Engine wraps the Coraza WAF instance
type Engine struct {
	waf      coraza.WAF
	mu       sync.RWMutex
	cfg      *config.WAFConfig
	matchedRules []types.MatchedRule
	rulesMu  sync.Mutex
}

// NewEngine creates and initializes a new WAF engine
func NewEngine(cfg *config.WAFConfig) (*Engine, error) {
	e := &Engine{cfg: cfg}
	if err := e.Reload(nil); err != nil {
		return nil, err
	}
	return e, nil
}

// Reload rebuilds the WAF with updated rules
func (e *Engine) Reload(rules []models.Rule) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	wafConfig := coraza.NewWAFConfig().
		WithRequestBodyAccess().
		WithRequestBodyLimit(e.cfg.RequestBodyLimit).
		WithResponseBodyAccess().
		WithResponseBodyLimit(e.cfg.ResponseBodyLimit).
		WithErrorCallback(func(mr types.MatchedRule) {
			e.rulesMu.Lock()
			e.matchedRules = append(e.matchedRules, mr)
			e.rulesMu.Unlock()
			log.Printf("[WAF] Rule matched: ID=%d msg=%q severity=%s ip=%s",
				mr.Rule().ID(), mr.Message(), mr.Rule().Severity().String(), mr.ClientIPAddress())
		})

	// Build directives string
	directives := buildDirectives(e.cfg, rules)

	wafConfig = wafConfig.WithDirectives(directives)

	waf, err := coraza.NewWAF(wafConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize WAF: %w", err)
	}

	e.waf = waf
	e.matchedRules = nil
	log.Printf("[WAF] Engine loaded with %d custom rules", len(rules))
	return nil
}

// ReloadWithCRS rebuilds the WAF with CRS rules
func (e *Engine) ReloadWithCRS(rules []models.Rule) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	wafConfig := coraza.NewWAFConfig().
		WithRequestBodyAccess().
		WithRequestBodyLimit(e.cfg.RequestBodyLimit).
		WithResponseBodyAccess().
		WithResponseBodyLimit(e.cfg.ResponseBodyLimit).
		WithErrorCallback(func(mr types.MatchedRule) {
			e.rulesMu.Lock()
			e.matchedRules = append(e.matchedRules, mr)
			e.rulesMu.Unlock()
			log.Printf("[WAF] Rule matched: ID=%d msg=%q severity=%s ip=%s",
				mr.Rule().ID(), mr.Message(), mr.Rule().Severity().String(), mr.ClientIPAddress())
		})

	// Build directives with CRS
	directives := buildDirectivesWithCRS(e.cfg, rules)

	wafConfig = wafConfig.WithDirectives(directives)

	waf, err := coraza.NewWAF(wafConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize WAF with CRS: %w", err)
	}

	e.waf = waf
	e.matchedRules = nil
	log.Printf("[WAF] Engine loaded with CRS and %d custom rules", len(rules))
	return nil
}

// DrainMatchedRules returns and clears collected matched rules
func (e *Engine) DrainMatchedRules() []types.MatchedRule {
	e.rulesMu.Lock()
	defer e.rulesMu.Unlock()
	rules := e.matchedRules
	e.matchedRules = nil
	return rules
}

// ProcessTransaction runs a request through the WAF and returns the action
func (e *Engine) ProcessTransaction(clientIP string, clientPort int, serverIP string, serverPort int,
	method, uri, httpVersion string, headers map[string][]string, body []byte) (*TransactionResult, error) {

	e.mu.RLock()
	waf := e.waf
	e.mu.RUnlock()

	if waf == nil {
		return &TransactionResult{Action: models.ActionAllow}, nil
	}

	tx := waf.NewTransaction()
	defer func() {
		tx.ProcessLogging()
		tx.Close()
	}()

	result := &TransactionResult{
		TransactionID: tx.ID(),
	}

	// Phase 1: Connection
	tx.ProcessConnection(clientIP, clientPort, serverIP, serverPort)

	// Phase 1: URI
	tx.ProcessURI(uri, method, httpVersion)

	// Phase 1: Request headers
	for key, values := range headers {
		for _, value := range values {
			tx.AddRequestHeader(key, value)
		}
	}

	if it := tx.ProcessRequestHeaders(); it != nil {
		result.Action = interruptionToAction(it)
		result.Status = it.Status
		result.MatchedRules = collectMatchedRules(tx)
		return result, nil
	}

	// Phase 2: Request body
	if len(body) > 0 {
		tx.WriteRequestBody(body)
	}
	if it, err := tx.ProcessRequestBody(); err == nil && it != nil {
		result.Action = interruptionToAction(it)
		result.Status = it.Status
		result.MatchedRules = collectMatchedRules(tx)
		return result, nil
	}

	// No interruption - allow
	result.Action = models.ActionAllow
	result.Status = 200
	result.MatchedRules = collectMatchedRules(tx)

	// Check anomaly score from matched rules
	score := 0
	for range tx.MatchedRules() {
		// OWASP CRS uses tx.anomaly_score_* variables
		score += 5 // default increment per matched rule
	}
	result.AnomalyScore = score

	return result, nil
}

// ProcessResponse processes the response phase
func (e *Engine) ProcessResponse(tx types.Transaction, statusCode int, headers map[string][]string, body []byte) *TransactionResult {
	result := &TransactionResult{Action: models.ActionAllow}

	for key, values := range headers {
		for _, value := range values {
			tx.AddResponseHeader(key, value)
		}
	}

	if it := tx.ProcessResponseHeaders(statusCode, "HTTP/1.1"); it != nil {
		result.Action = interruptionToAction(it)
		result.Status = it.Status
		return result
	}

	if len(body) > 0 {
		tx.WriteResponseBody(body)
	}
	if it, err := tx.ProcessResponseBody(); err == nil && it != nil {
		result.Action = interruptionToAction(it)
		result.Status = it.Status
		return result
	}

	return result
}

// TransactionResult holds the result of WAF processing
type TransactionResult struct {
	TransactionID string
	Action        models.ActionType
	Status        int
	AnomalyScore  int
	MatchedRules  string
}

func interruptionToAction(it *types.Interruption) models.ActionType {
	switch it.Action {
	case "deny":
		return models.ActionBlock
	case "drop":
		return models.ActionBlock
	case "redirect":
		return models.ActionBlock
	default:
		return models.ActionBlock
	}
}

func collectMatchedRules(tx types.Transaction) string {
	rules := tx.MatchedRules()
	if len(rules) == 0 {
		return ""
	}
	result := ""
	for _, mr := range rules {
		if result != "" {
			result += "\n"
		}
		result += fmt.Sprintf("Rule ID: %d, Message: %s, Severity: %s, Tags: %v",
			mr.Rule().ID(), mr.Message(), mr.Rule().Severity().String(), mr.Rule().Tags())
	}
	return result
}

func buildDirectives(cfg *config.WAFConfig, rules []models.Rule) string {
	base := `
SecRuleEngine On
SecRequestBodyAccess On
SecResponseBodyAccess On
SecRequestBodyLimit ` + fmt.Sprintf("%d", cfg.RequestBodyLimit) + `
SecResponseBodyLimit ` + fmt.Sprintf("%d", cfg.ResponseBodyLimit) + `
SecAuditEngine On

# Default deny rules for testing
SecRule REQUEST_HEADERS:User-Agent "@rx ^$" "id:10001,phase:1,deny,status:403,msg:'Empty User-Agent',severity:WARNING,tag:'custom/empty-ua'"
`

	if cfg.Mode == "detection_only" {
		base += "\nSecRuleEngine DetectionOnly\n"
	} else if cfg.Mode == "off" {
		base += "\nSecRuleEngine Off\n"
	}

	// Append custom rules
	for _, r := range rules {
		if r.IsEnabled {
			base += "\n" + r.SecRule + "\n"
		}
	}

	return base
}

func buildDirectivesWithCRS(cfg *config.WAFConfig, rules []models.Rule) string {
	var directives string

	// Load CRS rules if enabled
	if cfg.CRSEnabled && cfg.CRSPath != "" {
		crsRules, err := crs.LoadCRSRules(cfg.CRSPath)
		if err != nil {
			log.Printf("[WAF] Warning: Failed to load CRS rules: %v", err)
			log.Printf("[WAF] Falling back to basic rules without CRS")
			return buildDirectives(cfg, rules)
		}
		directives += crsRules + "\n"
		log.Printf("[WAF] CRS rules loaded successfully")
	} else {
		log.Printf("[WAF] CRS disabled or path not configured")
		// Add basic rules if CRS not enabled
		directives += buildDirectives(cfg, nil)
	}

	// Append custom rules
	for _, r := range rules {
		if r.IsEnabled {
			directives += "\n" + r.SecRule + "\n"
		}
	}

	return directives
}
