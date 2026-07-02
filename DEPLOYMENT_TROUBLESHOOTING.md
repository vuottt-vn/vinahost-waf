# 🔧 Dokploy Deployment Troubleshooting

## Error: "go.sum: not found"

### Cause
Docker build context is not set to the root directory where `go.mod` and `go.sum` files exist.

### Solution

**In Dokploy Dashboard:**

When configuring the application, make sure:

#### **For WAF API:**
```
Build Context: . (root directory)
Dockerfile Path: docker/Dockerfile.api
```

#### **For WAF Proxy:**
```
Build Context: . (root directory)
Dockerfile Path: docker/Dockerfile.waf
```

#### **For WAF Frontend:**
```
Build Context: ./web
Dockerfile Path: Dockerfile
```

---

## Common Build Context Mistakes

### ❌ WRONG:
```
Build Context: docker/
Dockerfile Path: Dockerfile.api
```
This won't work because `go.mod` and `go.sum` are in the root directory, not in `docker/`.

### ✅ CORRECT:
```
Build Context: .
Dockerfile Path: docker/Dockerfile.api
```
This works because Docker will look for `go.mod` and `go.sum` in the build context (root).

---

## Directory Structure

```
vinahost-waf/
├── go.mod          ← Must be in build context
├── go.sum          ← Must be in build context
── docker/
│   ├── Dockerfile.api
│   └── Dockerfile.waf
├── cmd/
│   ├── api/main.go
│   └── waf/main.go
└── web/
    ├── Dockerfile
    └── src/
```

---

## Dokploy Configuration Checklist

### WAF API Service
- [ ] **Build Context:** `.` (dot, meaning root)
- [ ] **Dockerfile:** `docker/Dockerfile.api`
- [ ] Verify `go.mod` and `go.sum` exist in repository root

### WAF Proxy Service
- [ ] **Build Context:** `.` (dot, meaning root)
- [ ] **Dockerfile:** `docker/Dockerfile.waf`
- [ ] Verify `go.mod` and `go.sum` exist in repository root

### WAF Frontend Service
- [ ] **Build Context:** `./web`
- [ ] **Dockerfile:** `Dockerfile`
- [ ] Verify `web/package.json` exists

---

## Manual Build Test

Before deploying to Dokploy, test locally:

```bash
# Test API build
docker build -f docker/Dockerfile.api -t waf-api-test .

# Test Proxy build
docker build -f docker/Dockerfile.waf -t waf-proxy-test .

# Test Frontend build
cd web
docker build -t waf-frontend-test .
```

If these work locally, they will work on Dokploy.

---

## Git Repository Verification

Make sure files are committed:

```bash
# Check if go.mod and go.sum are tracked
git ls-files | grep go.

# Expected output:
# go.mod
# go.sum

# If missing, add them:
git add go.mod go.sum
git commit -m "Add go.mod and go.sum"
git push
```

---

## Docker Compose Local Test

Test with docker-compose before Dokploy:

```bash
# Build all services
docker compose build

# Should complete without errors
# If you see "go.sum: not found", check build context
```

---

## Quick Fix for Dokploy

If you're getting "go.sum: not found":

1. **Pull latest code:**
   ```bash
   git pull
   ```

2. **In Dokploy Dashboard:**
   - Go to Application settings
   - Check "Build Context" field
   - Set to `.` (single dot)
   - NOT `./docker` or any subdirectory

3. **Redeploy:**
   - Click "Redeploy" or "Update"
   - Watch build logs for errors

---

## Expected Build Log

Successful build should show:

```
Step 1/10 : FROM golang:1.25-alpine AS builder
Step 2/10 : WORKDIR /app
Step 3/10 : COPY go.mod go.sum ./
  ---> Using cache
Step 4/10 : RUN go mod download
  ---> Running in xxxxx
  go: downloading github.com/corazawaf/coraza/v3 v3.7.0
  ...
Step 5/10 : COPY . .
  ---> xxxxx
Step 6/10 : RUN CGO_ENABLED=0 GOOS=linux go build ...
  ---> xxxxx
Successfully built xxxxx
```

If you see error at Step 3, the build context is wrong.

---

## Support

If still having issues:

1. Check Docker build logs in Dokploy
2. Verify repository structure on server:
   ```bash
   ls -la /path/to/repo/
   # Should show go.mod and go.sum
   ```
3. Check if using correct branch (main)
4. Verify Dokploy has access to repository

---

**Updated:** 2026-07-02  
**Related Issue:** Docker build context configuration
