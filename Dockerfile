# Build stage
FROM golang:1-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Install build dependencies for CGO and build the Go app
RUN apk add --no-cache gcc musl-dev \
 && CGO_ENABLED=1 GOOS=linux go build -o server ./cmd/server

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the built binary
COPY --from=builder /app/server .

# Copy .env and db directory
COPY --from=builder /app/.env ./
COPY --from=builder /app/db ./db

# Copy templates and static assets
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

# Expose the internal app port
EXPOSE 8080

# Command to run the app
CMD ["./server"]