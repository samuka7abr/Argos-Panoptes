FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY go.mod go.sum* ./
RUN go mod download

COPY ../shared /build/shared
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o agent .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/agent .
COPY config.example.yaml /app/config.yaml

ENV CONFIG_PATH=/app/config.yaml

CMD ["./agent"]

