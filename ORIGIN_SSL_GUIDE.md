# SSL Configuration Guide for Origins

## Overview

This guide explains how to handle different SSL scenarios when your origin/backend servers have various SSL configurations.

## Architecture

```
User (HTTPS)
    ↓
Nginx Reverse Proxy (HTTPS:443)
    ↓
WAF Proxy (HTTP:8080)
    ↓
Origin Server (HTTP or HTTPS)
```

---

## Scenario 1: Origin without SSL (HTTP) ✅

**Status:** Works automatically, no configuration needed.

### Configuration

In WAF Dashboard → Targets:
```
Upstream URL: http://your-origin:8080
```

### How it works:
- User → Nginx (HTTPS) ✅
- Nginx → WAF Proxy (HTTP:8080) ✅
- WAF Proxy → Origin (HTTP:8080) ✅

**No SSL errors because WAF Proxy communicates with origin via HTTP.**

---

## Scenario 2: Origin with Self-Signed SSL ⚠️

**Problem:** WAF Proxy may reject self-signed certificates.

### Solution A: Use HTTP instead of HTTPS

If possible, configure origin to accept HTTP:

```
Upstream URL: http://your-origin:8080  ← Use HTTP
```

### Solution B: Configure Nginx to skip SSL verification

Add to nginx.conf (already added):

```nginx
location /direct/ {
    rewrite ^/direct/(.*)$ /$1 break;
    
    proxy_pass https://your-origin-ip:443;
    
    # Skip SSL verification for self-signed certs
    proxy_ssl_verify off;
    proxy_ssl_server_name on;
    
    proxy_set_header Host $host;
    # ... other headers
}
```

**Access:**
```
https://app.vinahost.cloud/direct/ → Direct to origin (bypasses WAF)
```

### Solution C: Trust self-signed certificate

```bash
# Copy origin's certificate to nginx container
docker cp /path/to/origin.crt nginx-proxy:/etc/nginx/ssl/origin.crt

# Add to nginx.conf
proxy_ssl_trusted_certificate /etc/nginx/ssl/origin.crt;
proxy_ssl_verify on;
```

---

## Scenario 3: Origin with Expired/Invalid SSL ❌

**Problem:** SSL handshake fails.

### Solution 1: Use HTTP (if available)

```
Upstream URL: http://your-origin:8080
```

### Solution 2: Fix origin SSL certificate

```bash
# On origin server, renew certificate
certbot renew

# Or use Let's Encrypt
certbot --nginx -d your-origin.domain.com
```

### Solution 3: Skip SSL verification (temporary)

In nginx.conf:
```nginx
location / {
    proxy_pass https://your-origin:443;
    proxy_ssl_verify off;  # ← Skip verification
    proxy_ssl_server_name on;
}
```

️ **Warning:** This bypasses SSL security. Only use temporarily!

---

## Scenario 4: Origin with Valid SSL ✅

**Status:** Works perfectly, no issues.

### Configuration

In WAF Dashboard → Targets:
```
Upstream URL: https://your-origin:443
```

### How it works:
- User → Nginx (HTTPS) ✅
- Nginx → WAF Proxy (HTTP:8080) ✅
- WAF Proxy → Origin (HTTPS:443) ✅

**WAF Proxy can proxy to HTTPS origins without issues.**

---

## Recommended Best Practices

### 1. Use HTTPS for all origins (if possible)

```
✅ https://your-origin:443
❌ http://your-origin:8080
```

### 2. Use Let's Encrypt for free SSL

```bash
# On origin server
certbot --nginx -d your-origin.domain.com
```

### 3. If must use HTTP, ensure internal network is secure

```
Internal network only: http://origin:8080  ← OK
Public internet: http://origin:8080         ← NOT RECOMMENDED
```

### 4. Use WAF Proxy for protection

```
User → WAF Proxy (with security rules) → Origin
```

Instead of:
```
User → Origin directly (no protection)
```

---

## Troubleshooting

### Issue: "SSL certificate verify failed"

**Cause:** Origin has self-signed or invalid certificate.

**Solution:**
```nginx
# Option 1: Skip verification (temporary)
proxy_ssl_verify off;

# Option 2: Trust specific certificate
proxy_ssl_trusted_certificate /etc/nginx/ssl/origin.crt;
proxy_ssl_verify on;
```

### Issue: "SSL handshake failed"

**Cause:** Origin SSL/TLS protocol mismatch.

**Solution:**
```nginx
# Force specific SSL protocol
proxy_ssl_protocols TLSv1.2 TLSv1.3;
proxy_ssl_ciphers HIGH:!aNULL:!MD5;
```

### Issue: Connection works via HTTP but not HTTPS

**Cause:** Origin doesn't have SSL configured.

**Solution:**
```
Use HTTP: http://your-origin:8080
```

Or install SSL on origin:
```bash
certbot --nginx -d your-origin.domain.com
```

---

## Security Considerations

### Internal HTTP is OK if:

- ✅ WAF Proxy and Origin are on same internal network
- ✅ No public access to origin
- ✅ Firewall blocks external access to origin port

### Public HTTP is NOT OK if:

-  Origin is directly accessible from internet
- ❌ Sensitive data transmitted over HTTP
- ❌ Compliance requirements (PCI DSS, HIPAA, etc.)

---

## Quick Reference

| Origin SSL Status | WAF Proxy Config | Nginx Config | Status |
|-------------------|------------------|--------------|--------|
| No SSL (HTTP) | `http://origin:8080` | Default | ✅ Works |
| Valid SSL | `https://origin:443` | Default | ✅ Works |
| Self-signed SSL | `https://origin:443` | `proxy_ssl_verify off;` | ️ Works with config |
| Expired SSL | `https://origin:443` | `proxy_ssl_verify off;` | ⚠️ Works with config |
| Invalid cert | `https://origin:443` | Fix cert or use HTTP | ❌ Fix needed |

---

## Example Configurations

### Example 1: Internal API (HTTP)

```nginx
# WAF protects internal API
location /api/ {
    proxy_pass http://waf_proxy;
}
```

WAF Target:
```
Name: internal-api
Upstream URL: http://10.0.0.100:8080
```

### Example 2: External Service (HTTPS with valid cert)

```nginx
location /service/ {
    proxy_pass http://waf_proxy;
}
```

WAF Target:
```
Name: external-service
Upstream URL: https://api.example.com:443
```

### Example 3: Dev Server (Self-signed SSL)

```nginx
location /dev/ {
    proxy_pass https://dev-server:443;
    proxy_ssl_verify off;  # Self-signed cert
}
```

---

## Summary

1. **HTTP origins work fine** - WAF Proxy can proxy to HTTP backends
2. **Valid SSL origins work fine** - No configuration needed
3. **Self-signed SSL** - Use `proxy_ssl_verify off;` in nginx
4. **Best practice** - Use Let's Encrypt for free valid SSL
5. **Security** - Internal HTTP is OK, public HTTP is NOT OK

---

**Updated:** 2026-07-02  
**Related to:** WAF Proxy origin SSL configuration
