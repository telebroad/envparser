# syntax=docker/dockerfile:1

# Comments are provided throughout this file to help you get started.
# If you need more help, visit the Dockerfile reference guide at
# https://docs.docker.com/go/dockerfile-reference/

# Want to help us make this template better? Share your feedback here: https://forms.gle/ybq9Krt8jtBL3iCk7

################################################################################
# Create a stage for building the application.
ARG GO_VERSION=latest
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS build
WORKDIR /src

copy . .

# Build the application.
RUN go build -o /out/app ./cmd/example/main.go

ENTRYPOINT ["/out/app","--log-db.uri=some_uri"]


