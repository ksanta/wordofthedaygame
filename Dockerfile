# Dockerfile References: https://docs.docker.com/engine/reference/builder/
# Multistage builder tutorial: https://www.callicoder.com/docker-golang-image-container-example/

# Start from the latest golang base image
FROM golang:latest as builder

LABEL maintainer="Karl Santa"

WORKDIR /go/src/github.com/ksanta/wordofthedaygame

COPY . .

# Build the "server-app" binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server-app server/main.go

# Start a new stage from scratch
FROM alpine:latest

# Copy the pre-built binary file from the previous stage
COPY --from=builder /go/src/github.com/ksanta/wordofthedaygame/server-app /usr/local/bin

ENTRYPOINT ["server-app"]