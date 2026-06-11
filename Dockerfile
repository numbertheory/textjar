# Node build stage
FROM node:20-alpine AS node-builder

WORKDIR /app

# Copy package files and install dependencies
COPY package*.json ./
RUN npm install

# Copy source and build
COPY static/js ./static/js
RUN npm run build

# Go build stage
FROM golang:1.26-alpine AS builder

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
COPY templates ./templates
COPY static ./static
# Overwrite static/dist with the built bundle
COPY --from=node-builder /app/static/dist ./static/dist

# Expose the port the app runs on
EXPOSE 3636

# Run the application
CMD ["./main"]
