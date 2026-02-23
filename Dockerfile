# -------------------------------------------------------------------
# STAGE 1: Build client side assets (JavaScript & CSS)
# -------------------------------------------------------------------
FROM node:20-alpine AS build-js

WORKDIR /build
RUN npm install -g gulp-cli

# Cache dependencies first
COPY package*.json ./
RUN npm install --only=dev

# Build assets
COPY . .
RUN gulp

# -------------------------------------------------------------------
# STAGE 2: Build Golang binary
# -------------------------------------------------------------------
FROM golang:1.22-bookworm AS build-golang

WORKDIR /app

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a stripped binary
COPY . .
RUN go build -ldflags="-s -w" -trimpath -o gophish

# -------------------------------------------------------------------
# STAGE 3: Secure Runtime Container
# -------------------------------------------------------------------
FROM debian:bookworm-slim

# Create a non-root user
RUN useradd -m -d /opt/gophish -s /bin/bash app

# Install necessary runtime dependencies
RUN apt-get update && \
    apt-get install --no-install-recommends -y jq libcap2-bin ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

WORKDIR /opt/gophish

# Copy ONLY the compiled binary and necessary static files (No source code!)
COPY --from=build-golang /app/gophish ./
COPY --from=build-golang /app/config.json ./
COPY --from=build-golang /app/templates/ ./templates/
COPY --from=build-golang /app/static/ ./static/
COPY --from=build-golang /app/docker/run.sh ./docker/run.sh

# Overwrite static assets with Gulp build from Stage 1
COPY --from=build-js /build/static/js/dist/ ./static/js/dist/
COPY --from=build-js /build/static/css/dist/ ./static/css/dist/

# Set permissions and network capabilities
RUN chown -R app:app /opt/gophish && \
    setcap 'cap_net_bind_service=+ep' /opt/gophish/gophish && \
    chmod +x ./docker/run.sh

# Switch to non-root user
USER app

# Adjust configuration to bind to all interfaces
RUN sed -i 's/127.0.0.1/0.0.0.0/g' config.json && \
    touch config.json.tmp

EXPOSE 3333 8080 8443 80

CMD ["./docker/run.sh"]