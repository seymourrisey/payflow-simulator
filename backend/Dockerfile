# Stage 1: Builder
FROM golang:1.25.4-alpine AS build

RUN apk add --no-cache git ca-certificates tzdata
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-w -s" \
    -o /app/bin/payflow \
    ./cmd/api/main.go


# Stage 2: Runtime (minimal image)
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app


COPY --from=build /app/bin/payflow .

COPY --from=build /app/db/migrations ./db/migrations

EXPOSE 8080

ENTRYPOINT ["./payflow"]
