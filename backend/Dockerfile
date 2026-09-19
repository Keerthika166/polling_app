# Build Stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git and ca-certificates
RUN apk update && apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/server

# Production Stage
FROM alpine:3.20

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/server /app/server

EXPOSE 8080

CMD ["/app/server"]
