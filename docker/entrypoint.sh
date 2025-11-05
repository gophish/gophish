#!/usr/bin/env sh
set -e

# --- Required hostnames (set in App Settings) ---
: "${ADMIN_HOST:=admin.gophish.fairplay-digital.com}"
: "${PHISH_HOST:=gophish.fairplay-digital.com}"

# --- Ports inside the container ---
: "${FRONT_PORT:=8080}"   # Nginx external (exposed to Azure)
: "${ADMIN_PORT:=3333}"   # Gophish admin (internal)
: "${PHISH_PORT:=8081}"   # Gophish phish (internal)

# --- TLS terminates at Azure frontends; keep false in Gophish ---
: "${ADMIN_USE_TLS:=false}"
: "${PHISH_USE_TLS:=false}"

# --- MySQL (required) ---
: "${DB_HOST:?Set DB_HOST (e.g., sos-kinderdoerfer-gophish-server.mysql.database.azure.com)}"
: "${DB_PORT:=3306}"
: "${DB_USERNAME:?Set DB_USERNAME}"
: "${DB_PASSWORD:?Set DB_PASSWORD}"
: "${DB_DATABASE:?Set DB_DATABASE}"

# Optional first-login password for > v0.10.1: GOPHISH_INITIAL_ADMIN_PASSWORD (env)

# --- Write Gophish config.json ---
ADMIN_ORIGIN="https://${ADMIN_HOST}"
PHISH_ORIGIN="https://${PHISH_HOST}"

cat > /opt/gophish/config.json <<EOF_CONFIG
{
  "admin_server": {
    "listen_url": "0.0.0.0:${ADMIN_PORT}",
    "use_tls": ${ADMIN_USE_TLS},
    "trusted_origins": ["${ADMIN_ORIGIN}"]
  },
  "phish_server": {
    "listen_url": "0.0.0.0:${PHISH_PORT}",
    "use_tls": ${PHISH_USE_TLS},
    "trusted_origins": ["${PHISH_ORIGIN}"]
  },
  "db_name": "mysql",
  "db_path": "${DB_USERNAME}:${DB_PASSWORD}@(${DB_HOST}:${DB_PORT})/${DB_DATABASE}?charset=utf8mb4&parseTime=True&loc=UTC&tls=true"
}
EOF_CONFIG

# --- Nginx vhosts (host-based routing to the two internal ports) ---
mkdir -p /etc/nginx/conf.d
cat > /etc/nginx/conf.d/gophish.conf <<EOF_NGINX
server {
  listen ${FRONT_PORT};
  server_name ${PHISH_HOST};
  location / {
    proxy_pass http://127.0.0.1:${PHISH_PORT};
    proxy_set_header Host \$host;
    proxy_set_header X-Forwarded-Host \$host;
    proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto https;
    proxy_set_header X-Forwarded-Port 443;
    proxy_set_header X-Real-IP \$remote_addr;
  }
}
server {
  listen ${FRONT_PORT};
  server_name ${ADMIN_HOST};
  location / {
    proxy_pass http://127.0.0.1:${ADMIN_PORT};
    proxy_set_header Host \$host;
    proxy_set_header X-Forwarded-Host \$host;
    proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto https;
    proxy_set_header X-Forwarded-Port 443;
    proxy_set_header X-Real-IP \$remote_addr;
  }
}
EOF_NGINX

exec /usr/bin/supervisord -c /etc/supervisor/conf.d/supervisord.conf
