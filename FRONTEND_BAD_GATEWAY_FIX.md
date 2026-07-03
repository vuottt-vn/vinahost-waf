# 🔧 Troubleshooting: Frontend Bad Gateway

## Vấn đề
Frontend deployed thành công nhưng khi truy cập bị **Bad Gateway**.

## Nguyên nhân
Frontend nginx không thể kết nối đến WAF API backend.

---

## ✅ Giải Pháp

### **Cách 1: Đảm bảo cả 3 services cùng Docker Network**

#### **Kiểm tra network hiện tại:**

SSH vào server và chạy:
```bash
# List all networks
docker network ls

# Check which network each container is in
docker inspect waf-frontend | grep NetworkMode
docker inspect waf-api | grep NetworkMode
docker inspect waf-proxy | grep NetworkMode
```

#### **Tạo và gán network (nếu chưa có):**

```bash
# Tạo network mới
docker network create waf-network

# Gán các containers vào network
docker network connect waf-network waf-frontend
docker network connect waf-network waf-api
docker network connect waf-network waf-proxy
```

#### **Trong Dokploy Dashboard:**

1. Vào **Networks** tab
2. Tạo network: `waf-network`
3. Edit từng application:
   - WAF Frontend → Networks → Add `waf-network`
   - WAF API → Networks → Add `waf-network`
   - WAF Proxy → Networks → Add `waf-network`
4. Redeploy tất cả

---

### **Cách 2: Sử dụng localhost (Quick Fix)**

Nếu cả 3 services chạy trên **cùng một server**:

#### **Bước 1: Edit nginx.conf trong frontend container**

SSH vào server:
```bash
# Enter frontend container
docker exec -it waf-frontend sh

# Edit nginx config
vi /etc/nginx/conf.d/default.conf
```

Thay đổi:
```nginx
# FROM:
proxy_pass http://waf-api:3001/api/;

# TO:
proxy_pass http://localhost:3001/api/;
```

```bash
# Reload nginx
nginx -s reload
```

#### **Bước 2: Cập nhật source code và rebuild**

Edit `web/nginx.conf` trong source code:
```nginx
location /api/ {
    # Use localhost instead of container name
    proxy_pass http://localhost:3001/api/;
    ...
}
```

Rebuild và push:
```bash
git add web/nginx.conf
git commit -m "Use localhost for API proxy in frontend"
git push
```

Redeploy trong Dokploy.

---

### **Cách 3: Kiểm tra API có đang chạy không**

```bash
# Test API trực tiếp
curl http://localhost:3001/health

# Nếu trả về {"status":"ok"} → API đang chạy tốt
# Nếu timeout/refused → API chưa chạy hoặc sai port
```

```bash
# Check API container logs
docker logs waf-api

# Expected:
# API Server starting on :3001
```

---

### **Cách 4: Kiểm tra Firewall/Ports**

```bash
# Check if port 3001 is listening
netstat -tlnp | grep 3001

# Or
ss -tlnp | grep 3001

# Should show something like:
# tcp  LISTEN  0  0  0.0.0.0:3001  0.0.0.0:*  users:(("waf-api",pid=xxx,fd=x))
```

---

## 🔍 Debug Steps

### **Step 1: Verify API Service**

```bash
# Check if API container is running
docker ps | grep waf-api

# Check API logs
docker logs waf-api --tail 50

# Test API endpoint
curl -v http://localhost:3001/health
```

**Expected output:**
```
* Connected to localhost (127.0.0.1) port 3001
< HTTP/1.1 200 OK
< Content-Type: application/json
{"status":"ok"}
```

### **Step 2: Verify Frontend Container**

```bash
# Check frontend logs
docker logs waf-frontend --tail 50

# Enter container to debug
docker exec -it waf-frontend sh

# Test connectivity to API from inside container
wget -O- http://waf-api:3001/health
# OR
wget -O- http://localhost:3001/health

# Check nginx config
cat /etc/nginx/conf.d/default.conf
```

### **Step 3: Test Network Connectivity**

```bash
# From frontend container, test API
docker exec waf-frontend wget -O- http://waf-api:3001/health

# If fails, try with IP
docker inspect waf-api | grep IPAddress
# Suppose IP is 172.17.0.2
docker exec waf-frontend wget -O- http://172.17.0.2:3001/health
```

### **Step 4: Check nginx Error Logs**

```bash
# View nginx error logs
docker logs waf-frontend 2>&1 | grep -i error

# Or enter container
docker exec waf-frontend cat /var/log/nginx/error.log
```

**Common errors:**
- `connect() failed (111: Connection refused)` → API not reachable
- `no resolver defined` → DNS resolution failed
- `host not found` → Container name not in DNS

---

## 🛠️ Quick Fix Commands

### **Fix 1: Restart all services**

```bash
docker restart waf-api waf-frontend waf-proxy
```

### **Fix 2: Recreate frontend with correct config**

```bash
# Stop and remove frontend
docker stop waf-frontend
docker rm waf-frontend

# Rebuild and start
cd /path/to/vinahost-waf
docker compose up -d --build waf-frontend
```

### **Fix 3: Update docker-compose.yml**

Add `depends_on` to ensure API starts before frontend:

```yaml
waf-frontend:
  depends_on:
    - waf-api
  networks:
    - waf-network
```

---

## ✅ Verification Checklist

Sau khi fix, kiểm tra:

- [ ] `docker ps` shows all 3 containers as "Up"
- [ ] `curl http://localhost:3001/health` returns `{"status":"ok"}`
- [ ] `docker exec waf-frontend wget -O- http://localhost:3001/health` succeeds
- [ ] Access `https://secure.vinahost.cloud` → Login page appears
- [ ] Login with admin/admin123 → Dashboard loads
- [ ] No "Bad Gateway" errors in browser

---

## 📊 Expected Architecture

```
User Browser
    ↓
https://secure.vinahost.cloud:3000
    ↓
[waf-frontend container]
    nginx on port 80
    ↓
Proxy /api/* requests to:
    ↓
[waf-api container] ← MUST be reachable!
    API on port 3001
    ↓
Returns JSON responses
```

---

## 🆘 Still Not Working?

### **Collect Debug Info:**

```bash
# Save to file for support
docker ps > debug.txt
docker logs waf-api >> debug.txt
docker logs waf-frontend >> debug.txt
docker network ls >> debug.txt
docker inspect waf-frontend >> debug.txt
docker inspect waf-api >> debug.txt

# Share debug.txt for help
```

### **Common Issues:**

1. **API not in same network as Frontend**
   - Solution: Add both to `waf-network`

2. **Wrong API URL in nginx config**
   - Solution: Change `waf-api` to `localhost` or correct IP

3. **API not started before Frontend**
   - Solution: Add `depends_on: waf-api` in docker-compose

4. **Port 3001 not exposed**
   - Solution: Check API environment: `API_PORT=3001`

---

## 📝 Notes

- **Container name resolution** only works if containers are in the **same Docker network**
- **localhost** works if all containers run on the **same host**
- **IP address** always works but may change on container restart

---

**Updated:** 2026-07-02  
**Related to:** Frontend Bad Gateway on Dokploy deployment
