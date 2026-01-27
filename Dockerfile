# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app

# Tidak perlu go.sum/go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o kasir-api ./cmd/main.go

# Runtime
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/kasir-api .
EXPOSE 8080
CMD ["./kasir-api"]