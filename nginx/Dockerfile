FROM nginx:alpine

# Install certbot for Let's Encrypt
RUN apk add --no-cache certbot

# Copy nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Create directories
RUN mkdir -p /var/www/certbot \
    && mkdir -p /etc/nginx/ssl

# Expose ports
EXPOSE 80 443

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
