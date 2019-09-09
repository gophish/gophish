# Minify client side assets (JavaScript)
FROM node:latest AS build-js

WORKDIR /build
COPY . .

RUN npm install gulp gulp-cli -g
RUN npm install --only=dev
RUN gulp


# Build Golang binary
FROM golang:1.11 AS build-golang

WORKDIR /go/src/github.com/gophish/gophish
COPY . .

RUN go get -v && go build -v


# Runtime container
FROM debian:stable-slim

WORKDIR /opt/gophish
RUN useradd -d /opt/gophish -s /bin/bash app

RUN apt-get update && \
	apt-get install --no-install-recommends -y jq && \
	apt-get clean && \
	rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=build-golang /go/src/github.com/gophish/gophish/ ./
COPY --from=build-js /build/static/js/dist/ ./static/js/dist/
COPY --from=build-js /build/static/css/dist/ ./static/css/dist/
RUN touch config.json.tmp && chown app. config.json config.json.tmp
RUN sed -i 's/127.0.0.1/0.0.0.0/g' config.json

USER app
EXPOSE 3333 8080 8443

CMD ["./docker/run.sh"]
