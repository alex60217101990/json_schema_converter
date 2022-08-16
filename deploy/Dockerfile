############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/github.com/alex60217101990/json_schema_converter
COPY . .
# Fetch dependencies.
# Using go get.
RUN go get -d -v
# Build the binary.
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -v -a -o /go/bin/app ./main.go
############################
# STEP 2 install python genson package
############################
FROM python:3.9-alpine as builder-py
RUN pip3 install genson
############################
# STEP 3 build a small image
############################
FROM builder-py
# Copy our static executable.
COPY --from=builder /go/bin/app /go/bin/app
COPY ./schemas /schemas
# Run the hello binary.
ENTRYPOINT ["/go/bin/app"]