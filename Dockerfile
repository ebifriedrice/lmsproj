# Build stage
FROM golang:1.24.6-alpine AS builder

# Enable CGO
ENV CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64

# Install required build tools for CGO
RUN apk add --no-cache build-base

# Set working directory
WORKDIR /app

# Download dependencies early to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN go build -o main cmd/lms/main.go

# Run the binary
CMD ["./main"]
