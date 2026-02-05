# ---------- Build stage ----------
FROM golang:1.23-alpine AS builder
WORKDIR /app

# Aktifkan toolchain forwarding agar bisa otomatis upgrade ke 1.25.6
ENV GOTOOLCHAIN=auto

# Jika repo-mu tidak punya go.sum, tidak apa-apa; cukup copy go.mod
COPY go.mod ./
COPY go.sum ./
RUN go mod download || true

COPY . .
# Build semua paket; asumsi hanya ada satu package main
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -ldflags="-s -w" -o kasir-api .

# ---------- Runtime stage ----------
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/kasir-api .
EXPOSE 8080
CMD ["./kasir-api"]