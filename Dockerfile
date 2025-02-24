# Stage 1: Build the Go application
FROM golang:1.22 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first (for caching dependencies)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o rsvbackend main.go

# Stage 2: Create a minimal runtime image
FROM alpine:latest

# Install ca-certificates for HTTPS (if needed)
RUN apk --no-cache add ca-certificates

# Set the working directory
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/rsvbackend .

# Copy the templates directory
COPY --from=builder /app/templates ./templates

# Expose port 8080
EXPOSE 8080

# Set environment variables (optional, can be overridden in docker-compose or runtime)
ENV PORT=8080

# Run the application
CMD ["./rsvbackend"]