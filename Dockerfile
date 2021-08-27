# Use base golang image from Docker Hub
FROM golang:1.16 AS builder


# package update
RUN apk update &&\
    apk add --no-cache git mercurial

# app copy
WORKDIR /build
COPY . /build/


# Compile
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go mod vendor
RUN go mod download
WORKDIR /build
RUN go build -a -o goapp


# Now create separate deployment image
FROM gcr.io/distroless/base

# Definition of this variable is used by 'skaffold debug' to identify a golang binary.
# Default behavior - a failure prints a stack trace for the current goroutine.
# See https://golang.org/pkg/runtime/
ENV GOTRACEBACK=single

# Copy template & assets
WORKDIR /hello-world
COPY --from=builder /build/goapp /goapp
COPY index.html index.html
COPY assets assets/
COPY templates templates/

ENTRYPOINT ["./goapp"]
