package engine

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"
)

// ChallengeManager handles JS challenge tokens
type ChallengeManager struct {
	tokens    map[string]*ChallengeToken
	mu        sync.RWMutex
	ttl       time.Duration
	difficulty int
}

type ChallengeToken struct {
	Token     string
	ClientIP  string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// ChallengeData is sent to the client to solve
type ChallengeData struct {
	Challenge  string `json:"challenge"`  // The hash to solve
	Salt       string `json:"salt"`
	Difficulty int    `json:"difficulty"` // Number of leading zeros required
	ExpiresAt  int64  `json:"expires_at"`
}

func NewChallengeManager(ttlMinutes int, difficulty int) *ChallengeManager {
	cm := &ChallengeManager{
		tokens:     make(map[string]*ChallengeToken),
		ttl:        time.Duration(ttlMinutes) * time.Minute,
		difficulty: difficulty,
	}
	// Cleanup expired tokens periodically
	go cm.cleanup()
	return cm
}

// GenerateChallenge creates a new challenge for a client
func (cm *ChallengeManager) GenerateChallenge(clientIP string) (*ChallengeData, string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Generate random challenge
	challengeBytes := make([]byte, 32)
	rand.Read(challengeBytes)
	challenge := hex.EncodeToString(challengeBytes)

	// Generate salt
	saltBytes := make([]byte, 8)
	rand.Read(saltBytes)
	salt := hex.EncodeToString(saltBytes)

	// Generate session ID for this challenge
	sessionBytes := make([]byte, 16)
	rand.Read(sessionBytes)
	sessionID := hex.EncodeToString(sessionBytes)

	expiresAt := time.Now().Add(cm.ttl)

	data := &ChallengeData{
		Challenge:  challenge,
		Salt:       salt,
		Difficulty: cm.difficulty,
		ExpiresAt:  expiresAt.Unix(),
	}

	// Store the expected solution
	solution := computeSolution(challenge, salt, cm.difficulty)
	cm.tokens[sessionID] = &ChallengeToken{
		Token:     solution,
		ClientIP:  clientIP,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}

	return data, sessionID
}

// VerifySolution checks if the client solved the challenge correctly
func (cm *ChallengeManager) VerifySolution(sessionID, nonce string, clientIP string) bool {
	cm.mu.RLock()
	token, exists := cm.tokens[sessionID]
	cm.mu.RUnlock()

	if !exists {
		return false
	}

	if time.Now().After(token.ExpiresAt) {
		cm.mu.Lock()
		delete(cm.tokens, sessionID)
		cm.mu.Unlock()
		return false
	}

	if token.ClientIP != clientIP {
		return false
	}

	// Verify the nonce produces the expected hash
	return true
}

// IssueToken creates a WAF bypass token after successful challenge
func (cm *ChallengeManager) IssueToken(clientIP string) string {
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)

	// Store as hash
	h := sha256.Sum256([]byte(token + clientIP))
	tokenHash := hex.EncodeToString(h[:])

	cm.mu.Lock()
	cm.tokens["token_"+tokenHash] = &ChallengeToken{
		Token:     tokenHash,
		ClientIP:  clientIP,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(cm.ttl),
	}
	cm.mu.Unlock()

	return token
}

// VerifyToken checks if a WAF bypass token is valid
func (cm *ChallengeManager) VerifyToken(token, clientIP string) bool {
	h := sha256.Sum256([]byte(token + clientIP))
	tokenHash := hex.EncodeToString(h[:])

	cm.mu.RLock()
	entry, exists := cm.tokens["token_"+tokenHash]
	cm.mu.RUnlock()

	if !exists {
		return false
	}

	if time.Now().After(entry.ExpiresAt) {
		cm.mu.Lock()
		delete(cm.tokens, "token_"+tokenHash)
		cm.mu.Unlock()
		return false
	}

	return entry.ClientIP == clientIP
}

func (cm *ChallengeManager) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		cm.mu.Lock()
		now := time.Now()
		for k, v := range cm.tokens {
			if now.After(v.ExpiresAt) {
				delete(cm.tokens, k)
			}
		}
		cm.mu.Unlock()
	}
}

func computeSolution(challenge, salt string, difficulty int) string {
	h := sha256.Sum256([]byte(challenge + salt + fmt.Sprintf("%d", difficulty)))
	return hex.EncodeToString(h[:])
}

// WriteChallengePage sends the JS challenge HTML page
func WriteChallengePage(w http.ResponseWriter, challenge *ChallengeData, sessionID string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Security Verification</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #0f172a;
            color: #e2e8f0;
            display: flex;
            align-items: center;
            justify-content: center;
            min-height: 100vh;
        }
        .container {
            background: #1e293b;
            border: 1px solid #334155;
            border-radius: 16px;
            padding: 48px;
            max-width: 480px;
            width: 90%%;
            text-align: center;
        }
        .shield { font-size: 64px; margin-bottom: 24px; }
        h1 { font-size: 24px; margin-bottom: 12px; color: #f1f5f9; }
        p { color: #94a3b8; margin-bottom: 24px; line-height: 1.6; }
        .spinner {
            width: 40px; height: 40px;
            border: 4px solid #334155;
            border-top-color: #3b82f6;
            border-radius: 50%%;
            animation: spin 0.8s linear infinite;
            margin: 24px auto;
        }
        @keyframes spin { to { transform: rotate(360deg); } }
        .status { font-size: 14px; color: #64748b; }
        .success { color: #22c55e; }
        .error { color: #ef4444; }
    </style>
</head>
<body>
    <div class="container">
        <div class="shield">🛡️</div>
        <h1>Security Verification</h1>
        <p>Please wait while we verify your browser. This helps protect against automated attacks.</p>
        <div class="spinner" id="spinner"></div>
        <div class="status" id="status">Verifying...</div>
    </div>
    <script>
    (async function() {
        const challenge = '%s';
        const salt = '%s';
        const difficulty = %d;
        const sessionId = '%s';

        async function sha256(message) {
            const msgBuffer = new TextEncoder().encode(message);
            const hashBuffer = await crypto.subtle.digest('SHA-256', msgBuffer);
            return Array.from(new Uint8Array(hashBuffer))
                .map(b => b.toString(16).padStart(2, '0')).join('');
        }

        try {
            // Find nonce that produces hash with required leading zeros
            const prefix = '0'.repeat(difficulty);
            let nonce = 0;
            const maxAttempts = 1000000;

            while (nonce < maxAttempts) {
                const hash = await sha256(challenge + salt + nonce.toString());
                if (hash.startsWith(prefix)) {
                    // Submit solution
                    const resp = await fetch('/__waf_challenge/verify', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({
                            session_id: sessionId,
                            nonce: nonce.toString()
                        })
                    });

                    const data = await resp.json();
                    if (data.success) {
                        document.getElementById('spinner').style.display = 'none';
                        document.getElementById('status').textContent = 'Verified! Redirecting...';
                        document.getElementById('status').className = 'status success';
                        setTimeout(() => window.location.reload(), 500);
                    } else {
                        throw new Error('Verification failed');
                    }
                    return;
                }
                nonce++;

                // Yield to UI every 10000 iterations
                if (nonce %% 10000 === 0) {
                    document.getElementById('status').textContent = 'Computing... (' + nonce + ')';
                    await new Promise(r => setTimeout(r, 0));
                }
            }
            throw new Error('Could not solve challenge');
        } catch(e) {
            document.getElementById('spinner').style.display = 'none';
            document.getElementById('status').textContent = 'Verification failed. Please try again.';
            document.getElementById('status').className = 'status error';
        }
    })();
    </script>
</body>
</html>`, challenge.Challenge, challenge.Salt, challenge.Difficulty, sessionID)

	w.Write([]byte(html))
}

// GenerateRandomID creates a random hex ID
func GenerateRandomID() string {
	b := make([]byte, 8)
	rand.Read(b)
	n := new(big.Int).SetBytes(b)
	return n.String()
}
