# HTTPS Deployment Guide for WAF

## Overview

This guide shows how to deploy the WAF with HTTPS support using Nginx reverse proxy and Let's Encrypt SSL certificates.

## Architecture

```
Internet
   ↓
HTTPS:443 / HTTP:80
   ↓
[Nginx Reverse Proxy]
   ├─→ secure.vinahost.cloud → waf-frontend (Web UI)
   ├─→ /api/* → waf-api (Backend API)
   └─→ app.vinahost.cloud → waf-proxy (WAF protection)
```

## Deployment Steps

### Step 1: Deploy WAF Services

```bash
# Pull latest code
git pull

# Deploy all services
docker compose up -d --build

# Verify services are running
docker compose ps
```

### Step 2: Obtain SSL Certificate

**Option A: Using Certbot (Recommended)**

```bash
# Get certificate for your domain
docker run -it --rm \
  -v nginx-ssl:/etc/letsencrypt \
  -v certbot-www:/var/www/certbot \
  certbot/certbot certonly \
  --webroot \
  --webroot-path=/var/www/certbot \
  -d secure.vinahost.cloud \
  --email your-email@example.com \
  --agree-tos \
  --non-interactive
```

**Option B: Manual Certificate Upload**

If you already have SSL certificates:

```bash
# Copy certificates to volume
docker cp /path/to/fullchain.pem nginx-proxy:/etc/nginx/ssl/fullchain.pem
docker cp /path/to/privkey.pem nginx-proxy:/etc/nginx/ssl/privkey.pem

# Restart nginx
docker compose restart nginx-proxy
```

### Step 3: Configure DNS

Make sure your domain points to your server IP:

```
A record: secure.vinahost.cloud → YOUR_SERVER_IP
A record: app.vinahost.cloud → YOUR_SERVER_IP  (optional)
```

### Step 4: Verify HTTPS

```bash
# Test HTTPS connection
curl -I https://secure.vinahost.cloud

# Should return:
# HTTP/2 200
# strict-transport-security: max-age=31536000; includeSubDomains
```

### Step 5: Auto-Renewal (Automatic)

Certbot container automatically renews certificates every 12 hours.

```bash
# Check certificate expiry
docker exec certbot certbot certificates

# Manual renewal if needed
docker exec certbot certbot renew
```

## Configuration

### Domain Configuration

Edit `nginx/nginx.conf` and change:

```nginx
# Change these to your actual domains
server_name secure.vinahost.cloud;
server_name app.vinahost.cloud;
```

### SSL Certificate Paths

Certificates are stored in:
```
/etc/nginx/ssl/fullchain.pem  ← Public certificate
/etc/nginx/ssl/privkey.pem    ← Private key
```

Mapped to Docker volume: `nginx-ssl`

### HTTP to HTTPS Redirect

All HTTP traffic (port 80) is automatically redirected to HTTPS (port 443).

## Testing

### Test SSL Configuration

```bash
# Test with curl
curl -v https://secure.vinahost.cloud

# Test SSL Labs (online)
# https://www.ssllabs.com/ssltest/analyze.html?d=secure.vinahost.cloud
```

### Test Certificate Renewal

```bash
# Dry run renewal
docker exec certbot certbot renew --dry-run
```

## Troubleshooting

### Issue: Certificate not found

```bash
# Check certificate files
docker exec nginx-proxy ls -la /etc/nginx/ssl/

# Should show:
# fullchain.pem
# privkey.pem
```

### Issue: Nginx won't start

```bash
# Check nginx configuration
docker exec nginx-proxy nginx -t

# Check logs
docker compose logs nginx-proxy
```

### Issue: HTTP not redirecting to HTTPS

```bash
# Test redirect
curl -I http://secure.vinahost.cloud

# Should return: 301 Moved Permanently
# Location: https://secure.vinahost.cloud/
```

### Issue: Certificate renewal failed

```bash
# Check certbot logs
docker compose logs certbot

# Manual renewal with debug
docker exec -it certbot certbot renew --force-renewal
```

## Security Headers

The following security headers are automatically added:

- **Strict-Transport-Security**: HSTS enabled (1 year)
- **X-Frame-Options**: SAMEORIGIN (prevent clickjacking)
- **X-Content-Type-Options**: nosniff
- **X-XSS-Protection**: 1; mode=block
- **Referrer-Policy**: strict-origin-when-cross-origin

## Custom Domains

To add more domains:

1. Edit `nginx/nginx.conf`
2. Add new server block:

```nginx
server {
    listen 443 ssl http2;
    server_name your-new-domain.com;
    
    ssl_certificate /etc/nginx/ssl/fullchain.pem;
    ssl_certificate_key /etc/nginx/ssl/privkey.pem;
    
    location / {
        proxy_pass http://waf_frontend;
        # ... proxy settings
    }
}
```

3. Get certificate for new domain:

```bash
docker run -it --rm \
  -v nginx-ssl:/etc/letsencrypt \
  certbot/certbot certonly \
  --webroot \
  --webroot-path=/var/www/certbot \
  -d your-new-domain.com
```

4. Restart nginx:

```bash
docker compose restart nginx-proxy
```

## Production Checklist

- [ ] DNS configured for all domains
- [ ] SSL certificates obtained and valid
- [ ] HTTP → HTTPS redirect working
- [ ] Security headers present
- [ ] Auto-renewal configured and tested
- [ ] Firewall ports 80 and 443 open
- [ ] WAF services running and healthy
- [ ] HTTPS access verified

## Port Configuration

| Service | Port | Protocol |
|---------|------|----------|
| HTTP | 80 | TCP |
| HTTPS | 443 | TCP |
| WAF API (internal) | 3001 | TCP |
| WAF Proxy (internal) | 8080 | TCP |
| Frontend (internal) | 3000 | TCP |

Only ports 80 and 443 need to be exposed to the internet.

## Docker Volumes

| Volume | Purpose |
|--------|---------|
| `nginx-ssl` | SSL certificates |
| `certbot-www` | Let's Encrypt challenge files |
| `waf-data` | WAF database |
| `waf-logs` | WAF logs |

## Support

For issues:
1. Check logs: `docker compose logs nginx-proxy`
2. Check certificate: `docker exec certbot certbot certificates`
3. Test config: `docker exec nginx-proxy nginx -t`

---

**Updated:** 2026-07-02  
**Status:** Ready for production deployment
