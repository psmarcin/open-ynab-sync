# Build stage
FROM golang:1.24.4-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/open-ynab-sync

# Final stage
FROM alpine:latest

WORKDIR /app

# Add non root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy the binary from builder
COPY --from=builder /app/open-ynab-sync .

# Set ownership
RUN chown -R appuser:appgroup /app

# Use non root user
USER appuser

# Set environment variables
ENV GC_SECRET_ID=""
ENV GC_SECRET_KEY=""
ENV GC_ACCOUNT_ID=""
ENV YNAB_ACCOUNT_ID=""
ENV YNAB_TOKEN=""

# Run the application
CMD ["./open-ynab-sync"]
