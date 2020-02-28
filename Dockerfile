# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Multistage builder tutorial: https://www.callicoder.com/docker-golang-image-container-example/

# Start from the latest golang base image
FROM golang:latest as builder

# Add Maintainer Info
LABEL maintainer="Karl Santa"

# Set the Current Working Directory inside the container
WORKDIR /go/src/app

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

######## Start a new stage from scratch #######
FROM alpine:latest

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /go/src/app/app /usr/local/bin

# Default command if none is given. Override to choose different colours.
CMD ["app"]