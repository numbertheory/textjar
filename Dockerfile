# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from the build stage
COPY --from=builder /app/main .

# Copy the static assets and templates
COPY content ./content
COPY templates ./templates
COPY static ./static

# Expose the port the app runs on
EXPOSE 8080

# Run the application
CMD ["./main"]
