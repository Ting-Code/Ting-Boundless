# TLS settings for production. Mount certs to /etc/nginx/certs/.
listen 443 ssl;
http2 on;
ssl_certificate     /etc/nginx/certs/fullchain.pem;
ssl_certificate_key /etc/nginx/certs/privkey.pem;
ssl_protocols       TLSv1.2 TLSv1.3;
ssl_prefer_server_ciphers off;
ssl_session_cache     shared:SSL:10m;
ssl_session_timeout   1d;

add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
