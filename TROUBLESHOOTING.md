# 🔧 Troubleshooting: Bad Gateway Error

## Issue
Getting "Bad Gateway" error when accessing the WAF proxy on port 8080.

## Root Cause
The WAF Proxy container was missing required environment variables and had a broken health check.

## ✅ Fixed In Latest Commit

The following fixes have been pushed to the repository:

1. ✅ Added missing `CHALLENGE_*` environment variables
2. ✅ Added `WAF_REQ_BODY_LIMIT` and `WAF_RES_BODY_LIMIT`
3. ✅ Removed broken health checks (wget not available in alpine)
4. ✅ Improved logging configuration

## 🚀 How to Fix on Dokploy

### **Step 1: Pull Latest Changes**

SSH into your Dokploy server:

```bash
cd /path/to/vinahost-waf
git pull
```

### **Step 2: Rebuild and Restart**

```bash
# Stop existing containers
docker compose down

# Rebuild with fixes
docker compose up -d --build

# Check logs
docker compose logs -f
```

### **Step 3: Verify Both Services**

```bash
# List containers
docker compose ps
```

**Expected output:**
```
NAME          STATUS         PORTS
waf-proxy     Up             0.0.0.0:8080->8080/tcp
waf-api       Up             0.0.0.0:3001->3001/tcp
```

### **Step 4: Check Logs**

```bash
# Check WAF Proxy logs (should show startup messages)
docker compose logs waf-proxy

# Check WAF API logs
docker compose logs waf-api
```

**Expected WAF Proxy logs:**
```
Database connected successfully (SQLite: waf.db)
Database migrations completed
Loaded 0 custom rules from database
[WAF] Engine loaded with 0 custom rules
[WAF] Rate limiter enabled: 100 requests per 60 seconds
WAF Proxy starting on :8080
```

### **Step 5: Test Services**

```bash
# Test WAF API (should return {"status":"ok"})
curl http://localhost:3001/health

# Test WAF Proxy (should return block page or forward request)
curl http://localhost:8080/
```

## 🔍 Common Issues

### **Issue 1: WAF Proxy Not Starting**

**Symptoms:**
- No logs from waf-proxy container
- Container status is "Exited" or "Restarting"

**Solution:**
```bash
# Check container logs
docker logs waf-proxy

# Check if port is already in use
sudo lsof -i :8080

# Kill process using the port
sudo kill -9 <PID>

# Restart
docker compose up -d
```

### **Issue 2: Database Permission Error**

**Symptoms:**
```
failed to initialize database: permission denied
```

**Solution:**
```bash
# Fix volume permissions
docker compose down
sudo chown -R 1000:1000 ./data ./logs
docker compose up -d
```

### **Issue 3: Missing Environment Variables**

**Symptoms:**
```
panic: environment variable not set
```

**Solution:**

Make sure your `.env` file has these variables:

```env
# Required
JWT_SECRET=your-secret-key-here

# WAF Configuration
WAF_MODE=on
WAF_CRS_ENABLED=false

# Challenge Configuration
CHALLENGE_ENABLED=true
CHALLENGE_TOKEN_TTL=30
CHALLENGE_DIFFICULTY=3

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_MAX_REQUESTS=100
RATE_LIMIT_WINDOW_SECONDS=60

# Body Limits
WAF_REQ_BODY_LIMIT=131072
WAF_RES_BODY_LIMIT=131072
```

### **Issue 4: Container Crashes Immediately**

**Symptoms:**
- Container starts then stops within seconds
- No useful logs

**Solution:**

Run container interactively to see errors:

```bash
# Run WAF proxy in foreground
docker run --rm -it \
  -e DB_PATH=/app/data/waf.db \
  -e JWT_SECRET=test \
  -p 8080:8080 \
  vinahost-waf-waf-proxy
```

This will show the actual error message.

## 📊 Deployment Checklist

After deploying, verify:

- [ ] `docker compose ps` shows both containers as "Up"
- [ ] `docker compose logs waf-proxy` shows "WAF Proxy starting on :8080"
- [ ] `docker compose logs waf-api` shows "API Server starting on :3001"
- [ ] `curl http://localhost:3001/health` returns `{"status":"ok"}`
- [ ] `curl http://localhost:8080/` returns HTML (block page or proxied content)
- [ ] Can access dashboard at `http://your-server:3001`
- [ ] Can login with admin/admin123

## 🆘 Still Having Issues?

### **Collect Debug Information**

```bash
# Docker version
docker --version
docker compose version

# Container status
docker compose ps

# Full logs
docker compose logs > waf-logs.txt

# Environment
docker compose exec waf-proxy env
docker compose exec waf-api env

# Disk space
df -h

# Memory
free -h
```

### **Share Debug Info**

Create an issue on GitHub with:
1. `waf-logs.txt` output
2. Docker version
3. Server OS
4. `.env` file (remove secrets)

## ✅ Success Indicators

You'll know it's working when:

1. **Both containers running:**
   ```
   docker compose ps
   waf-proxy   Up
   waf-api     Up
   ```

2. **API responds:**
   ```bash
   $ curl http://localhost:3001/health
   {"status":"ok"}
   ```

3. **WAF Proxy responds:**
   ```bash
   $ curl http://localhost:8080/
   <!DOCTYPE html>
   <html>
   ... (WAF block page or proxied content)
   ```

4. **Dashboard accessible:**
   - Open browser: `http://your-server:3001`
   - Login page appears
   - Can login with admin/admin123

---

**Fixed in commit:** `855fa96`
**Pushed to:** https://github.com/vuottt-vn/vinahost-waf
