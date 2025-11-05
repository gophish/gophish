# syntax=docker/dockerfile:1

### Build stage
FROM golang:1.22-bullseye AS build
WORKDIR /src

# Install gcc + sqlite3 dev libs for CGO support
RUN apt-get update && \
    apt-get install -y --no-install-recommends gcc libsqlite3-dev && \
    rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build for linux/amd64 (App Service) with CGO enabled
ENV CGO_ENABLED=1
RUN GOOS=linux GOARCH=amd64 go build -o gophish

### Runtime stage
FROM debian:stable-slim
WORKDIR /opt/gophish

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates tzdata nginx supervisor libsqlite3-0 && \
    rm -rf /var/lib/apt/lists/* && \
    rm -f /etc/nginx/sites-enabled/default

# App + supervisor config + entrypoint
COPY --from=build /src/ /opt/gophish/
COPY docker/supervisord.conf /etc/supervisor/conf.d/supervisord.conf
COPY docker/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# App Service exposes ONE HTTP port; we use 8080
EXPOSE 8080
ENV FRONT_PORT=8080 ADMIN_PORT=3333 PHISH_PORT=8081 ADMIN_USE_TLS=false PHISH_USE_TLS=false

ENTRYPOINT ["/entrypoint.sh"]
