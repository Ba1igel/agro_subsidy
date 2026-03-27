FROM golang:1.21-alpine AS builder

WORKDIR /app

# Download deps first (cached layer)
COPY go.mod ./
RUN go mod download || true

COPY . .
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -o go-service ./cmd/main.go

# ---

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/go-service .

CMD ["./go-service"]
