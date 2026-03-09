
# Stage 1: Builder

FROM golang:1.25.4-alpine AS build

# Install deps yang dibutuhkan untuk build (git, ca-certificates)
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy dependency files dulu (layer cache lebih efisien)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy seluruh source code
COPY . .

# Build binary — CGO disabled untuk static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-w -s" \
    -o /app/bin/payflow \
    ./cmd/api/main.go


# Stage 2: Runtime (minimal image)

FROM alpine:3.21

# ca-certificates → HTTPS ke PostgreSQL/Supabase
# tzdata          → timezone support
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary dari builder stage
COPY --from=builder /app/bin/payflow .

# Copy migration SQL (opsional, kalau dibutuhkan runtime)
COPY --from=builder /app/db/migrations ./db/migrations

# Expose port default (bisa di-override via APP_PORT env)
EXPOSE 8080

# Jalankan binary
ENTRYPOINT ["./payflow"]
