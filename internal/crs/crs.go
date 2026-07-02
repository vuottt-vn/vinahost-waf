package crs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Config holds OWASP CRS configuration
type Config struct {
	Enabled    bool
	CRSPath    string // Path to CRS rules directory
	IncludeCRS bool   // Whether to include CRS rules in WAF
}

// DefaultConfig returns a default CRS configuration
func DefaultConfig() *Config {
	// Look for CRS in common locations
	crsPath := findCRSPath()
	
	return &Config{
		Enabled:    crsPath != "",
		CRSPath:    crsPath,
		IncludeCRS: true,
	}
}

// findCRSPath searches for OWASP CRS in common locations
func findCRSPath() string {
	// Check common locations
	locations := []string{
		"/etc/coraza/crs",
		"/usr/local/lib/coraza/crs",
		"./configs/crs",
		"../configs/crs",
	}

	for _, loc := range locations {
		if _, err := os.Stat(filepath.Join(loc, "crs-setup.conf.example")); err == nil {
			log.Printf("[CRS] Found OWASP CRS at: %s", loc)
			return loc
		}
	}

	log.Printf("[CRS] OWASP CRS not found in default locations")
	return ""
}

// LoadCRSRules loads all CRS rule files and returns them as a string
func LoadCRSRules(crsPath string) (string, error) {
	if crsPath == "" {
		return "", fmt.Errorf("CRS path is empty")
	}

	// Check if CRS directory exists
	if _, err := os.Stat(crsPath); os.IsNotExist(err) {
		return "", fmt.Errorf("CRS directory not found: %s", crsPath)
	}

	var rules string

	// Load CRS setup configuration
	setupFile := filepath.Join(crsPath, "crs-setup.conf.example")
	if content, err := loadFile(setupFile); err == nil {
		rules += content + "\n"
		log.Printf("[CRS] Loaded CRS setup configuration")
	} else {
		// Try without .example extension
		setupFile = filepath.Join(crsPath, "crs-setup.conf")
		if content, err := loadFile(setupFile); err == nil {
			rules += content + "\n"
			log.Printf("[CRS] Loaded CRS setup configuration")
		}
	}

	// Load REQUEST-9XX rules
	ruleFiles := []string{
		"REQUEST-901-INITIALIZATION.conf",
		"REQUEST-903.9001-DRUPAL-EXCLUSION-RULES.conf",
		"REQUEST-903.9002-WORDPRESS-EXCLUSION-RULES.conf",
		"REQUEST-903.9003-NEXTCLOUD-EXCLUSION-RULES.conf",
		"REQUEST-903.9004-DOKUWIKI-EXCLUSION-RULES.conf",
		"REQUEST-903.9005-CPANEL-EXCLUSION-RULES.conf",
		"REQUEST-903.9006-XENFORO-EXCLUSION-RULES.conf",
		"REQUEST-905-COMMON-EXCEPTIONS.conf",
		"REQUEST-910-IP-REPUTATION.conf",
		"REQUEST-911-METHOD-ENFORCEMENT.conf",
		"REQUEST-912-DOS-PROTECTION.conf",
		"REQUEST-913-SCANNER-DETECTION.conf",
		"REQUEST-920-PROTOCOL-ENFORCEMENT.conf",
		"REQUEST-921-PROTOCOL-ATTACK.conf",
		"REQUEST-930-APPLICATION-ATTACK-LFI.conf",
		"REQUEST-931-APPLICATION-ATTACK-RFI.conf",
		"REQUEST-932-APPLICATION-ATTACK-RCE.conf",
		"REQUEST-933-APPLICATION-ATTACK-PHP.conf",
		"REQUEST-934-APPLICATION-ATTACK-NODEJS.conf",
		"REQUEST-941-APPLICATION-ATTACK-XSS.conf",
		"REQUEST-942-APPLICATION-ATTACK-SQLI.conf",
		"REQUEST-943-APPLICATION-ATTACK-SESSION-FIXATION.conf",
		"REQUEST-944-APPLICATION-ATTACK-JAVA.conf",
		"REQUEST-949-BLOCKING-EVALUATION.conf",
	}

	loadedCount := 0
	for _, file := range ruleFiles {
		filePath := filepath.Join(crsPath, file)
		if content, err := loadFile(filePath); err == nil {
			rules += content + "\n"
			loadedCount++
		}
	}

	log.Printf("[CRS] Loaded %d CRS rule files", loadedCount)
	return rules, nil
}

// loadFile reads a file and returns its content as a string
func loadFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// DownloadCRS provides instructions for downloading OWASP CRS
func DownloadCRS(targetDir string) error {
	log.Printf("[CRS] To download OWASP CRS, run:")
	log.Printf("[CRS]   mkdir -p %s", targetDir)
	log.Printf("[CRS]   cd %s", targetDir)
	log.Printf("[CRS]   git clone https://github.com/coreruleset/coreruleset.git .")
	log.Printf("[CRS]   cp crs-setup.conf.example crs-setup.conf")
	log.Printf("[CRS]")
	log.Printf("[CRS] Or manually download from: https://github.com/coreruleset/coreruleset/releases")
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create CRS directory: %w", err)
	}

	return nil
}
