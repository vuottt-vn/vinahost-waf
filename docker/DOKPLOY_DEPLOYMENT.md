# 🚀 Dokploy Deployment Guide for WAF

This guide will help you deploy the Web Application Firewall on Dokploy.

## 📋 Prerequisites

- ✅ Dokploy installed and running
- ✅ Docker & Docker Compose installed on server
- ✅ Domain name (optional, for production)
- ✅ SSL certificates (optional, for HTTPS)

## 🏗️ Architecture

```
Internet
   ↓
[Dokploy Reverse Proxy]
   ↓
┌─────────────────────────────────────┐
│  waf-proxy (port 8080)              │
│  - Coraza WAF Engine                │
│  - Rate Limiting                    │
│  - IP Filter                        │
└──────────────┬──────────────────────┘
               ↓
┌─────────────────────────────────────┐
│  Your Backend Application           │
│  (protected by WAF)                 │
└─────────────────────────────────────┘

Separate Container:
┌─────────────────────────────────────┐
│  waf-api (port 3001)                │
│  - Management API                   │
│  - Dashboard Backend                │
└─────────────────────────────────────┘
```

## 📦 Deployment Steps

### **Step 1: Prepare Your Server**

SSH into your Dokploy server:

```bash
ssh user@your-server.com
```

Create a deployment directory:

```bash
mkdir -p /opt/waf-deployment
cd /opt/waf-deployment
```

### **Step 2: Upload Project Files**

You have two options:

#### **Option A: Clone from Git Repository**

```bash
git clone https://github.com/your-username/webappfirewall.git .
```

#### **Option B: Upload via SCP/SFTP**

From your local machine:

```bash
scp -r webappfirewall/* user@your-server.com:/opt/waf-deployment/
```

### **Step 3: Configure Environment**

Copy the environment template:

```bash
cp .env.example .env
```

Edit the `.env` file with your production settings:

```bash
nano .env
```

**Required changes for production:**

```env
# CRITICAL: Change this!
JWT_SECRET=your-very-long-random-secret-key-here

# Enable OWASP CRS for production
WAF_CRS_ENABLED=true

# Configure rate limiting for your needs
RATE_LIMIT_MAX_REQUESTS=100
RATE_LIMIT_WINDOW_SECONDS=60
```

**Generate a secure JWT secret:**

```bash
# Linux/Mac
openssl rand -base64 64

# Or use Python
python3 -c "import secrets; print(secrets.token_urlsafe(64))"
```

### **Step 4: Deploy with Docker Compose**

Build and start the services:

```bash
docker compose up -d --build
```

Check the logs:

```bash
docker compose logs -f
```

### **Step 5: Verify Deployment**

Check if services are running:

```bash
docker compose ps
```

Expected output:
```
NAME          STATUS          PORTS
waf-proxy     Up (healthy)    0.0.0.0:8080->8080/tcp
waf-api       Up (healthy)    0.0.0.0:3001->3001/tcp
```

Test the health endpoints:

```bash
# Test WAF API
curl http://localhost:3001/health

# Test WAF Proxy
curl http://localhost:8080/
```

### **Step 6: Configure Firewall**

Open required ports:

```bash
# UFW (Ubuntu/Debian)
sudo ufw allow 8080/tcp  # WAF Proxy
sudo ufw allow 3001/tcp  # WAF API
sudo ufw reload

# Or with iptables
sudo iptables -A INPUT -p tcp --dport 8080 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 3001 -j ACCEPT
```

### **Step 7: Access the Dashboard**

Open your browser:

```
http://your-server-ip:3001
```

**Default Login:**
- Username: `admin`
- Password: `admin123`

⚠️ **Change the default password immediately!**

## 🔧 Dokploy-Specific Configuration

### **Option 1: Deploy as Docker Compose Application**

In Dokploy dashboard:

1. Go to **Applications** → **Add Application**
2. Select **Docker Compose**
3. Upload `docker-compose.yml`
4. Set environment variables from `.env`
5. Deploy

### **Option 2: Deploy as Separate Services**

#### **Service 1: WAF Proxy**

```yaml
Service Name: waf-proxy
Dockerfile: docker/Dockerfile.waf
Ports: 8080:8080
Environment:
  DB_PATH: /app/data/waf.db
  JWT_SECRET: ${JWT_SECRET}
  WAF_MODE: on
  RATE_LIMIT_ENABLED: "true"
Volumes:
  waf-data:/app/data
  waf-logs:/app/logs
```

#### **Service 2: WAF API**

```yaml
Service Name: waf-api
Dockerfile: docker/Dockerfile.api
Ports: 3001:3001
Environment:
  DB_PATH: /app/data/waf.db
  JWT_SECRET: ${JWT_SECRET}
Volumes:
  waf-data:/app/data
  waf-logs:/app/logs
```

### **Option 3: Deploy via Dokploy CLI**

If Dokploy has CLI support:

```bash
dokploy deploy --name waf --compose docker-compose.yml --env .env
```

## 🌐 Setting Up Domain & SSL

### **With Dokploy Reverse Proxy**

1. In Dokploy dashboard, go to **Reverse Proxy**
2. Add new proxy rule:
   - **Domain**: `waf.yourdomain.com`
   - **Target**: `http://localhost:8080`
3. Enable SSL/Let's Encrypt
4. Save

### **Manual Nginx Configuration**

Create `/etc/nginx/sites-available/waf`:

```nginx
server {
    listen 80;
    server_name waf.yourdomain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

server {
    listen 80;
    server_name waf-api.yourdomain.com;

    location / {
        proxy_pass http://localhost:3001;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

Enable the site:

```bash
sudo ln -s /etc/nginx/sites-available/waf /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

Get SSL certificate:

```bash
sudo certbot --nginx -d waf.yourdomain.com -d waf-api.yourdomain.com
```

## 📊 Monitoring & Maintenance

### **View Logs**

```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f waf-proxy
docker compose logs -f waf-api
```

### **Restart Services**

```bash
docker compose restart
```

### **Update WAF**

```bash
# Pull latest code
git pull

# Rebuild and restart
docker compose up -d --build
```

### **Backup Database**

```bash
# Backup SQLite database
docker cp waf-proxy:/app/data/waf.db ./waf-backup-$(date +%Y%m%d).db
```

### **Restore Database**

```bash
# Stop services
docker compose down

# Restore database
docker cp ./waf-backup-20240101.db waf-proxy:/app/data/waf.db

# Start services
docker compose up -d
```

## 🔍 Troubleshooting

### **Services Won't Start**

Check logs for errors:

```bash
docker compose logs waf-proxy
docker compose logs waf-api
```

Common issues:
- **Port already in use**: Change ports in `.env`
- **Permission denied**: Check volume permissions
- **Out of memory**: Increase server RAM

### **Can't Access Dashboard**

1. Check if API is running:
   ```bash
   curl http://localhost:3001/health
   ```

2. Check firewall:
   ```bash
   sudo ufw status
   ```

3. Check Docker ports:
   ```bash
   docker compose ps
   ```

### **WAF Not Blocking Attacks**

1. Verify WAF mode:
   ```bash
   docker compose exec waf-proxy env | grep WAF_MODE
   ```

2. Check WAF logs:
   ```bash
   docker compose logs waf-proxy | grep -i "rule matched"
   ```

3. Enable CRS:
   ```env
   WAF_CRS_ENABLED=true
   ```

## 🛡️ Security Best Practices

### **1. Change Default Credentials**

Immediately after first login:
- Go to **Users** → Select `admin` → Change password

### **2. Use Strong JWT Secret**

```bash
# Generate secure secret
openssl rand -base64 64
```

### **3. Enable OWASP CRS**

```env
WAF_CRS_ENABLED=true
```

### **4. Configure Rate Limiting**

```env
RATE_LIMIT_ENABLED=true
RATE_LIMIT_MAX_REQUESTS=100
RATE_LIMIT_WINDOW_SECONDS=60
```

### **5. Enable HTTPS**

- Use Let's Encrypt for free SSL
- Force HTTPS redirect
- Use HSTS headers

### **6. Regular Updates**

```bash
# Weekly update check
git pull
docker compose up -d --build
```

### **7. Backup Regularly**

```bash
# Daily backup cron
0 2 * * * docker cp waf-proxy:/app/data/waf.db /backup/waf-$(date +\%Y\%m\%d).db
```

## 📈 Performance Tuning

### **Increase Body Limits** (if needed)

```env
WAF_REQ_BODY_LIMIT=262144  # 256KB
WAF_RES_BODY_LIMIT=262144
```

### **Adjust Rate Limits**

```env
# For high-traffic sites
RATE_LIMIT_MAX_REQUESTS=500
RATE_LIMIT_WINDOW_SECONDS=60
```

### **Detection Mode** (for testing)

```env
WAF_MODE=detection_only  # Log only, don't block
```

## 🎯 Post-Deployment Checklist

- [ ] Changed default admin password
- [ ] Set strong JWT_SECRET
- [ ] Enabled OWASP CRS
- [ ] Configured rate limiting
- [ ] Set up SSL/HTTPS
- [ ] Configured domain name
- [ ] Tested WAF blocking (SQL injection, XSS)
- [ ] Set up backups
- [ ] Configured monitoring
- [ ] Documented access credentials (securely)

## 📞 Support

If you encounter issues:

1. Check logs: `docker compose logs -f`
2. Review this guide
3. Check WAF documentation
4. Open an issue on GitHub

---

**Deployment Complete!** 🎉

Your WAF is now protecting your applications with:
- ✅ Coraza WAF Engine
- ✅ OWASP CRS Rules
- ✅ Rate Limiting
- ✅ IP Management
- ✅ JS Challenge
- ✅ Modern Dashboard

Access your dashboard at: `http://your-server:3001`
