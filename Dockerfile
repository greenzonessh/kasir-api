# ---------- Build stage ----------
FROM golang:1.26-alpine AS builder
WORKDIR /app

# Kalau tidak punya go.sum tidak apa-apa; cukup copy go.mod
COPY go.mod ./
RUN go mod download || true

COPY . .
# Build semua, asumsikan hanya 1 package main
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -ldflags="-s -w" -o kasir-api ./...

# ---------- Runtime stage ----------
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/kasir-api .
EXPOSE 8080
CMD ["./kasir-api"]