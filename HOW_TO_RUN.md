# 🚀 How to Run the WAF Application

## ⚠️ Important: Go Not Installed

The current system does **not have Go installed**, which is required to build and run the WAF application.

## Option 1: Install Go (Recommended)

### Step 1: Install Go
1. Download Go from: https://go.dev/dl/
2. Run the Windows installer
3. Restart your terminal/PowerShell

### Step 2: Run the Application
```powershell
# Run the startup script
.\run-waf.ps1
```

This script will:
- ✅ Build the WAF Proxy with new features
- ✅ Build the WAF API Server
- ✅ Start both services
- ✅ Configure SQLite database
- ✅ Enable rate limiting
- ✅ Set up default admin account

### Step 3: Access the Application
- **WAF Proxy**: http://localhost:8080
- **WAF API**: http://localhost:3001
- **Frontend Dashboard**: http://localhost:5173 (if running)

**Default Login:**
- Username: `admin`
- Password: `admin123`

---

## Option 2: Manual Build (After Installing Go)

If you prefer to build manually:

```powershell
# Set environment variables
$env:DB_PATH = "waf.db"
$env:WAF_CRS_ENABLED = "false"
$env:RATE_LIMIT_ENABLED = "true"
$env:RATE_LIMIT_MAX_REQUESTS = "100"
$env:RATE_LIMIT_WINDOW_SECONDS = "60"

# Build WAF Proxy
go build -o waf.exe ./cmd/waf

# Build WAF API
go build -o waf-api.exe ./cmd/api

# Run WAF Proxy (in one terminal)
.\waf.exe

# Run WAF API (in another terminal)
.\waf-api.exe
```

---

## Option 3: Use Existing Executables (Limited)

The existing `.exe` files were built with PostgreSQL configuration and **won't work** without a PostgreSQL database. You must rebuild with Go to use SQLite.

---

## 📋 Environment Variables

### Database Configuration
```powershell
$env:DB_PATH = "waf.db"  # SQLite database file
```

### Rate Limiting
```powershell
$env:RATE_LIMIT_ENABLED = "true"
$env:RATE_LIMIT_MAX_REQUESTS = "100"       # Max requests per window
$env:RATE_LIMIT_WINDOW_SECONDS = "60"      # Time window in seconds
```

### OWASP CRS
```powershell
$env:WAF_CRS_ENABLED = "false"  # Set to "true" after running scripts\setup-crs.bat
$env:WAF_CRS_PATH = ".\configs\crs"
```

### Other Settings
```powershell
$env:API_PORT = "3001"
$env:PROXY_PORT = "8080"
$env:JWT_SECRET = "your-secret-key-here"
```

---

## 🔧 Setup OWASP CRS (Optional but Recommended)

To enable advanced attack detection:

```powershell
# Run the CRS setup script
.\scripts\setup-crs.bat

# Then rebuild and run with CRS enabled
$env:WAF_CRS_ENABLED = "true"
.\run-waf.ps1
```

---

## 🎯 What's New in This Version

### ✅ Rate Limiting
- Protects against abuse and DoS attacks
- 100 requests per minute by default
- Custom 429 error page with retry information
- Admin can reset limits via API

### ✅ IP Management
- **Whitelist**: Trusted IPs bypass all checks
- **Blacklist**: Block malicious IPs immediately
- Temporary blocks with auto-expiration
- Bulk import support

### ✅ OWASP CRS Integration
- SQL injection protection
- XSS attack prevention
- Path traversal detection
- Scanner detection
- Protocol enforcement
- And many more attack types

---

## 🐛 Troubleshooting

### Error: "Go is not recognized"
**Solution**: Install Go from https://go.dev/dl/ and restart terminal

### Error: "failed to connect to database"
**Solution**: Set `$env:DB_PATH = "waf.db"` to use SQLite instead of PostgreSQL

### Error: "port already in use"
**Solution**: Change the port:
```powershell
$env:API_PORT = "3002"
$env:PROXY_PORT = "8081"
```

### Frontend not loading
**Solution**: Start the frontend separately:
```powershell
cd web
npm install
npm run dev
```

---

## 📊 Testing the New Features

### Test Rate Limiting
```powershell
# Send 110 requests rapidly
for ($i=1; $i -le 110; $i++) {
    Invoke-WebRequest -Uri "http://localhost:8080/" -Method GET
}
# Should get 429 error after 100 requests
```

### Test IP Blacklist
```powershell
# Login first
$login = Invoke-RestMethod -Uri "http://localhost:3001/api/auth/login" `
  -Method POST -Body '{"username":"admin","password":"admin123"}' `
  -ContentType "application/json"
$token = $login.access_token
$headers = @{ "Authorization" = "Bearer $token" }

# Add IP to blacklist
Invoke-RestMethod -Uri "http://localhost:3001/api/ips" -Method POST `
  -Headers $headers -ContentType "application/json" `
  -Body '{"ip_address":"192.168.1.100","list_type":"blacklist","reason":"Testing"}'
```

### Check Rate Limit Stats
```powershell
Invoke-RestMethod -Uri "http://localhost:3001/api/rate-limit/stats" -Headers $headers
```

---

## 📝 Quick Start Checklist

- [ ] Install Go from https://go.dev/dl/
- [ ] Restart terminal
- [ ] Run `.\run-waf.ps1`
- [ ] Wait for both services to start
- [ ] Open http://localhost:8080 in browser
- [ ] Login with admin/admin123
- [ ] (Optional) Run `.\scripts\setup-crs.bat` for OWASP CRS
- [ ] (Optional) Start frontend: `cd web && npm run dev`

---

## 🎉 You're Ready!

Once Go is installed and you run the startup script, you'll have a fully functional WAF with:
- ✅ Rate limiting protection
- ✅ IP whitelist/blacklist management
- ✅ OWASP CRS attack detection
- ✅ Modern web dashboard
- ✅ RESTful API for management

**Enjoy your secure WAF!** 🛡️
