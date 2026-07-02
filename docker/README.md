# 🐳 Docker Deployment Components for WAF

All necessary files for deploying the Web Application Firewall on Dokploy or any Docker-based platform.

## 📁 Files Created

### **1. Dockerfiles**

#### `docker/Dockerfile.waf`
- Multi-stage build for WAF Proxy
- Alpine Linux base (minimal image)
- Health check included
- Exposes port 8080

#### `docker/Dockerfile.api`
- Multi-stage build for WAF API Server
- Alpine Linux base (minimal image)
- Health check included
- Exposes port 3001

### **2. Docker Compose**

#### `docker-compose.yml`
- Complete stack orchestration
- WAF Proxy service
- WAF API service
- Shared volumes for data persistence
- Custom network isolation
- Environment variable configuration

### **3. Configuration Files**

#### `.env.example`
- Complete environment variable template
- Documented all configuration options
- Production-ready defaults
- Security best practices

#### `docker/.dockerignore`
- Excludes unnecessary files from Docker context
- Reduces build time and image size
- Excludes sensitive files

### **4. Documentation**

#### `docker/DOKPLOY_DEPLOYMENT.md`
- Step-by-step deployment guide
- Dokploy-specific instructions
- Domain & SSL setup
- Monitoring & maintenance
- Troubleshooting guide
- Security best practices

## 🚀 Quick Start for Dokploy

### **Step 1: Copy Files to Server**

```bash
# Clone or upload project to server
scp -r webappfirewall/ user@your-server:/opt/waf-deployment/
```

### **Step 2: Configure Environment**

```bash
cd /opt/waf-deployment
cp .env.example .env
nano .env  # Edit with your settings
```

**Minimum required changes:**
```env
JWT_SECRET=generate-a-secure-random-key-here
WAF_CRS_ENABLED=true  # Enable for production
```

### **Step 3: Deploy**

```bash
docker compose up -d --build
```

### **Step 4: Verify**

```bash
# Check status
docker compose ps

# View logs
docker compose logs -f

# Test health
curl http://localhost:3001/health
```

## 📊 Service Ports

| Service | Port | Purpose |
|---------|------|---------|
| WAF Proxy | 8080 | Traffic protection (Coraza WAF) |
| WAF API | 3001 | Management backend |

## 🔧 Environment Variables

### **Critical (Must Configure)**

```env
JWT_SECRET=<generate-secure-random-key>
```

### **Important for Production**

```env
WAF_CRS_ENABLED=true
RATE_LIMIT_ENABLED=true
RATE_LIMIT_MAX_REQUESTS=100
```

### **Optional Tuning**

```env
WAF_MODE=on                    # on, detection_only, off
WAF_REQ_BODY_LIMIT=131072      # 128KB
WAF_RES_BODY_LIMIT=131072      # 128KB
CHALLENGE_DIFFICULTY=3         # 1-5
```

## 🛡️ Security Checklist

- [ ] Generated secure JWT_SECRET
- [ ] Changed default admin password
- [ ] Enabled OWASP CRS (`WAF_CRS_ENABLED=true`)
- [ ] Enabled rate limiting
- [ ] Configured HTTPS/SSL
- [ ] Set up firewall rules
- [ ] Configured regular backups

## 📈 Monitoring

### **View Logs**

```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f waf-proxy
```

### **Check Health**

```bash
curl http://localhost:3001/health
curl http://localhost:8080/
```

### **Resource Usage**

```bash
docker stats waf-proxy waf-api
```

## 🔄 Updates

```bash
# Pull latest changes
git pull

# Rebuild and restart
docker compose up -d --build

# Clean up old images
docker image prune -f
```

## 💾 Backup

```bash
# Backup database
docker cp waf-proxy:/app/data/waf.db ./waf-backup-$(date +%Y%m%d).db

# Backup logs
docker cp waf-proxy:/app/logs ./waf-logs-$(date +%Y%m%d)
```

## 🆘 Troubleshooting

### **Port Already in Use**

```bash
# Check what's using the port
sudo lsof -i :8080
sudo lsof -i :3001

# Change ports in .env
PROXY_PORT=8081
API_PORT=3002
```

### **Services Not Starting**

```bash
# Check logs
docker compose logs waf-proxy
docker compose logs waf-api

# Rebuild
docker compose down
docker compose up -d --build
```

### **Database Issues**

```bash
# Check volume
docker volume inspect waf-deployment_waf-data

# Reset database (WARNING: deletes all data)
docker compose down -v
docker compose up -d
```

## 📚 Additional Resources

- **Full Deployment Guide**: `docker/DOKPLOY_DEPLOYMENT.md`
- **Environment Config**: `.env.example`
- **WAF Documentation**: `IMPLEMENTATION_SUMMARY.md`
- **Testing Guide**: `TESTING_GUIDE.md`

## ✅ Deployment Verification

After deployment, verify:

1. ✅ Services running: `docker compose ps`
2. ✅ Health checks passing
3. ✅ Can access API: `curl http://localhost:3001/health`
4. ✅ Can access proxy: `curl http://localhost:8080/`
5. ✅ Login to dashboard works
6. ✅ WAF blocking test attacks
7. ✅ Rate limiting working
8. ✅ OWASP CRS loaded (if enabled)

---

**Ready for Production!** 🎉

All Docker components are prepared for deployment on Dokploy or any Docker-based platform.
