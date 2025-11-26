FROM golang:1.22-alpine AS builder

WORKDIR /build/agent

COPY shared ../shared
COPY agent/go.mod agent/go.sum ./
RUN go mod download

COPY agent/ .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o agent .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/agent/agent .
COPY agent/config.docker.yaml /app/config.yaml

ENV CONFIG_PATH=/app/config.yaml

CMD ["./agent"]

