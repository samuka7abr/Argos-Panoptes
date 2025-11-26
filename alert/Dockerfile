FROM golang:1.22-alpine AS builder

WORKDIR /build/alert

COPY shared ../shared
COPY alert/go.mod alert/go.sum ./
RUN go mod download

COPY alert/ .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o alert .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/alert/alert .
COPY alert/rules.docker.yaml /app/rules.yaml

ENV RULES_PATH=/app/rules.yaml
ENV API_URL=http://api:8082

CMD ["./alert"]


