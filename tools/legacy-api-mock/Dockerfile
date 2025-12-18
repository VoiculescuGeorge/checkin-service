# Use a simple base image
FROM golang:1.25.5-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum from the root to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the mock server source code
COPY ./tools/legacy-api-mock/ .

# Build the mock server as a static binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/legacy-api-mock .

# Stage 2: Create the final, minimal image
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/legacy-api-mock .
CMD ["./legacy-api-mock"]