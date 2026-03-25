ARG BUILD_FROM=ghcr.io/home-assistant/amd64-base:3.21
FROM golang:1.25-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /todo-app .

FROM ${BUILD_FROM}

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /todo-app /usr/local/bin/todo-app

COPY run.sh /
RUN chmod a+x /run.sh

CMD ["/run.sh"]
