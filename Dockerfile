# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /librendarium ./cmd/librendarium

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /

# Install certificates
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /librendarium /librendarium

# Run the utility
ENTRYPOINT ["/librendarium"]

