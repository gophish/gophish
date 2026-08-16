# Minify client side assets (JavaScript)
FROM node:latest AS build-js

RUN npm install gulp gulp-cli -g

WORKDIR /build
COPY . .
RUN npm install --only=dev
RUN gulp


# Build Golang binary
FROM golang:1.15.2 AS build-golang

WORKDIR /go/src/github.com/gophish/gophish
COPY . .
# Run the models tests (incl. the CVE-2024-55196 MarshalJSON checks) before
# building. A test failure fails the build, so a broken patch never produces
# an image. Tests run only in this build stage, never in the runtime image.
RUN go get -v && go test ./models/... && go build -v


# Runtime container
FROM debian:stable-slim

RUN useradd -m -d /opt/gophish -s /bin/bash app

RUN apt-get update && \
	apt-get install --no-install-recommends -y jq libcap2-bin ca-certificates gosu && \
	apt-get clean && \
	rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

WORKDIR /opt/gophish
COPY --from=build-golang /go/src/github.com/gophish/gophish/ ./
COPY --from=build-js /build/static/js/dist/ ./static/js/dist/
COPY --from=build-js /build/static/css/dist/ ./static/css/dist/
COPY --from=build-golang /go/src/github.com/gophish/gophish/config.json ./
RUN chown app. config.json

RUN setcap 'cap_net_bind_service=+ep' /opt/gophish/gophish

# Build-time config prep as the app user so the files it needs to rewrite at
# runtime (config.json, config.json.tmp) are owned by app.
USER app
RUN sed -i 's/127.0.0.1/0.0.0.0/g' config.json
RUN touch config.json.tmp

# The Railway entrypoint needs root to fix Volume ownership before dropping
# back to app via gosu, so the runtime user is switched inside the entrypoint,
# not here.
USER root
RUN chmod +x docker/run.sh docker/railway-entrypoint.sh

EXPOSE 3333 8080 8443 80

ENTRYPOINT ["./docker/railway-entrypoint.sh"]
