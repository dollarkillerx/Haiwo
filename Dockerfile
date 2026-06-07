# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/haiwo-server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/haiwo-agent ./cmd/agent

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates git openssh-client tzdata

WORKDIR /app

COPY --from=builder /app/bin/haiwo-server .
COPY --from=builder /app/bin/haiwo-agent .
COPY configs/config.toml ./configs/

EXPOSE 8080

CMD ["./haiwo-server", "-c", "config", "-cPath", "./,./configs/"]
