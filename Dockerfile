# Build stage
FROM golang:1.24-alpine AS builder

# Install git (needed for go modules)
RUN apk add --no-cache git

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with flags to reduce size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o leon-speedcams-go ./cmd

# Final stage - minimal image
FROM gcr.io/distroless/static:nonroot

WORKDIR /

COPY --from=builder /app/leon-speedcams-go .

# Default command
CMD ["./leon-speedcams-go"]
