FROM golang:1.22-alpine AS builder

WORKDIR /build/api

COPY shared ../shared
COPY api/go.mod api/go.sum ./
RUN go mod download

COPY api/ .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/api/api .
COPY api/schema.sql .

EXPOSE 8081

CMD ["./api"]

