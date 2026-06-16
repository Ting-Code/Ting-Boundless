# Shared location blocks for HTTP (:80) and HTTPS (:443) server blocks.

add_header X-Content-Type-Options nosniff always;
add_header X-Frame-Options DENY always;
add_header Referrer-Policy no-referrer-when-downgrade always;

location /oidc/   { proxy_pass http://logto:3001; include /etc/nginx/proxy.conf; }
location /logto/  { proxy_pass http://logto:3001; include /etc/nginx/proxy.conf; }

location = /admin {
    return 302 /admin/;
}

location /admin/ {
    root /usr/share/nginx/html;
    try_files $uri $uri/ /admin/index.html;
}

location /v1/auth/ {
    limit_req zone=auth burst=10 nodelay;
    proxy_pass http://gateway;
    include /etc/nginx/proxy.conf;
}

location = /internal/webhooks/logto {
    limit_req zone=auth burst=10 nodelay;
    proxy_pass http://gateway;
    include /etc/nginx/proxy.conf;
}

location / {
    limit_req zone=general burst=100 nodelay;
    proxy_pass http://gateway;
    include /etc/nginx/proxy.conf;
}
