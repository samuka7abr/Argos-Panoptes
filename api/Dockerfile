FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY go.mod go.sum* ./
RUN go mod download

COPY ../shared /build/shared
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/api .
COPY schema.sql .

EXPOSE 8081

CMD ["./api"]

