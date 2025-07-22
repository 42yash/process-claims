# Use the official Go image as the base image
FROM golang:1.24.5-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download Go module dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o process-claims

# Use a minimal alpine image for the final stage
FROM alpine:3.19

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/process-claims .
# Copy templates directory
COPY --from=builder /app/templates ./templates

# Expose port 8080
EXPOSE 8080

# Command to run the application
CMD ["./process-claims"]
