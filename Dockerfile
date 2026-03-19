# Stage 1: Build the Go binary
FROM golang:1.25-alpine AS builder
WORKDIR /app

# Install git for fetching dependencies
RUN apk add --no-cache git

# Download dependencies first (caching layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build a statically linked binary
# Pointing specifically to the cmd directory where main.go lives
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o alphacrawl ./cmd/alphacrawl

# Stage 2: Create a minimal production image
FROM alpine:latest
WORKDIR /root/

# Install root certificates for HTTPS scraping and tzdata for timezones
RUN apk --no-cache add ca-certificates tzdata

# Copy the binary from the builder stage
COPY --from=builder /app/alphacrawl .

EXPOSE 8080

# Run the binary
CMD ["./alphacrawl"]