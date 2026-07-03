# Setup Guide for 0dong.io.vn

## Overview

This guide will help you protect your website `0dong.io.vn` with the WAF, with origin server at `https://103.9.76.10/`.

## Architecture

```
User → https://0dong.io.vn → Nginx (HTTPS) → WAF Proxy (HTTP:8080) → Origin (https://103.9.76.10/)
```

---

## Step 1: Configure DNS

In your domain registrar's DNS management (where you bought 0dong.io.vn):

**Add A record:**
```
Type: A
Name: @ (or leave blank)
Value: YOUR_SERVER_IP  (IP address where WAF is deployed)
TTL: 300
```

**Add A record for www:**
```
Type: A
Name: www
Value: YOUR_SERVER_IP
TTL: 300
```

**Example:**
```
0dong.io.vn.    A    YOUR_SERVER_IP
www.0dong.io.vn. A    YOUR_SERVER_IP
```

⏱️ **Wait 5-10 minutes** for DNS to propagate.

---

## Step 2: Add Target in WAF Dashboard

1. **Login to WAF Dashboard:**
   ```
   https://secure.vinahost.cloud
   Username: admin
   Password: admin123
   ```

2. **Navigate to Targets:**
   - Click on **Targets** in the left menu

3. **Add New Target:**
   - Click **Add Target** button
   - Fill in the form:
     ```
     Name: 0dong-io-vn
     Upstream URL: https://103.9.76.10/
     Is Enabled: ✅ (checked)
     ```
   - Click **Create**

4. **Verify Target:**
   - You should see your target in the list
   - Status should be **Active**

---

## Step 3: Deploy Updated Configuration

SSH into your server and run:

```bash
cd /path/to/vinahost-waf

# Pull latest code
git pull

# Restart services with new config
docker compose down
docker compose up -d --build

# Check status
docker compose ps
```

---

## Step 4: Get SSL Certificate for 0dong.io.vn

Run certbot to get free SSL certificate:

```bash
# Get certificate
docker run -it --rm \
  -v vinahost-waf_nginx-ssl:/etc/letsencrypt \
  -v vinahost-waf_certbot-www:/var/www/certbot \
  certbot/certbot certonly \
  --webroot \
  --webroot-path=/var/www/certbot \
  -d 0dong.io.vn \
  -d www.0dong.io.vn \
  --email your-email@example.com \
  --agree-tos \
  --non-interactive

# Restart nginx
docker compose restart nginx-proxy
```

**Replace `your-email@example.com` with your actual email.**

---

## Step 5: Verify Everything Works

### Test DNS:
```bash
# Should return YOUR_SERVER_IP
ping 0dong.io.vn
nslookup 0dong.io.vn
```

### Test HTTP Redirect:
```bash
# Should redirect to HTTPS
curl -I http://0dong.io.vn

# Expected:
# HTTP/1.1 301 Moved Permanently
# Location: https://0dong.io.vn/
```

### Test HTTPS:
```bash
# Should return your origin website
curl -I https://0dong.io.vn

# Expected:
# HTTP/2 200
# (content from https://103.9.76.10/)
```

### Test in Browser:
Open: **https://0dong.io.vn**

You should see your website (content from origin 103.9.76.10) protected by WAF!

---

## Step 6: Test WAF Protection

### Test SQL Injection:
```
https://0dong.io.vn/?id=1'+OR+'1'='1
```
**Expected:** WAF blocks the request (403 Forbidden)

### Test XSS:
```
https://0dong.io.vn/?q=<script>alert('xss')</script>
```
**Expected:** WAF blocks the request (403 Forbidden)

### Test Rate Limiting:
Make 100+ rapid requests to:
```
https://0dong.io.vn/
```
**Expected:** After 100 requests, get 429 Too Many Requests

---

## Configuration Summary

### Nginx Configuration:

```nginx
server {
    listen 443 ssl http2;
    server_name 0dong.io.vn www.0dong.io.vn;
    
    location / {
        proxy_pass http://waf_proxy;  # → Port 8080
        # WAF inspects and forwards to origin
    }
}
```

### WAF Target:

```
Name: 0dong-io-vn
Upstream URL: https://103.9.76.10/
```

### Request Flow:

1. User accesses `https://0dong.io.vn`
2. Nginx receives request (HTTPS)
3. Nginx proxies to WAF Proxy (HTTP:8080)
4. WAF Proxy inspects request:
   - ✅ Rate limiting check
   - ✅ IP whitelist/blacklist check
   - ✅ OWASP CRS rules check
   - ✅ Custom rules check
5. If request passes all checks:
   - WAF Proxy forwards to `https://103.9.76.10/`
   - Returns response to user
6. If request is blocked:
   - WAF returns 403 Forbidden page

---

## Troubleshooting

### Issue: DNS not resolving

```bash
# Check DNS propagation
dig 0dong.io.vn

# Should show YOUR_SERVER_IP
```

**Solution:** Wait for DNS to propagate (up to 24 hours)

### Issue: SSL certificate error

```bash
# Check certificate
docker exec certbot certbot certificates

# Renew if expired
docker exec certbot certbot renew
```

### Issue: 502 Bad Gateway

```bash
# Check WAF Proxy logs
docker compose logs waf-proxy

# Check if origin is accessible
curl -I https://103.9.76.10/

# Verify target is enabled in WAF Dashboard
```

### Issue: WAF not blocking attacks

```bash
# Check WAF rules are enabled
# Login to Dashboard → Rules → Verify rules are active

# Check logs for matched rules
docker compose logs waf-proxy | grep "Rule matched"
```

---

## Security Checklist

- [x] DNS configured for 0dong.io.vn
- [x] Target added in WAF Dashboard
- [x] HTTPS certificate obtained
- [x] HTTP → HTTPS redirect working
- [x] WAF protection active
- [x] Rate limiting enabled
- [x] OWASP CRS rules loaded
- [x] Tested SQL injection protection
- [x] Tested XSS protection
- [x] Tested rate limiting

---

## Quick Commands Reference

```bash
# View WAF Proxy logs (real-time)
docker compose logs -f waf-proxy

# View Nginx logs
docker compose logs -f nginx-proxy

# Check if origin is accessible
curl -I https://103.9.76.10/

# Test WAF blocking
curl https://0dong.io.vn/?id=1'+OR+'1'='1

# Restart all services
docker compose down && docker compose up -d

# Check disk space
df -h

# Check memory usage
free -h
```

---

## Support

If you encounter issues:

1. Check logs: `docker compose logs`
2. Verify DNS: `nslookup 0dong.io.vn`
3. Test origin: `curl -I https://103.9.76.10/`
4. Check WAF Dashboard → Targets → Verify enabled

---

**Updated:** 2026-07-02  
**Domain:** 0dong.io.vn  
**Origin:** https://103.9.76.10/  
**Status:** Ready for deployment
