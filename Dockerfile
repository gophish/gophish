# create image from the official Go image
FROM golang:alpine
RUN apk add --update tzdata \
    bash wget curl git;
# Create binary directory, install glide and fresh
RUN mkdir -p $$GOPATH/bin && \
    curl https://glide.sh/get | sh && \
    go get github.com/pilu/fresh
# define work directory
ADD . /Go/src/github.com/gophish/gophish
WORKDIR /Go/src/github.com/gophish/gophish
# serve the app
CMD glide update && fresh -c runner.conf main.go