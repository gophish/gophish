# setup build image
FROM golang:1.11 AS build

# build Gophish binary
WORKDIR /build/gophish
COPY . .
RUN go get -d -v ./...
RUN go build


# setup run image
FROM debian:stable-slim

RUN apt-get update && \
    apt-get install --no-install-recommends -y \
    jq && \
    apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# copy Gophish assets from the build image
WORKDIR /gophish
COPY --from=build /build/gophish/ /gophish/
RUN chmod +x gophish

# expose the admin port to the host
RUN sed -i 's/127.0.0.1/0.0.0.0/g' config.json

# expose default ports
EXPOSE 80 443 3333

CMD ["./docker/run.sh"]
