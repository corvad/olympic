# Build stage
FROM golang:1.25.6-alpine AS builder

# Install make for building via Makefile
RUN apk add --no-cache make

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application using Makefile
RUN CGO_ENABLED=0 make build && mv cascade.bin cascade

# Final stage
FROM alpine:latest

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/cascade .

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./cascade"]
