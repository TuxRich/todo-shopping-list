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

RUN mkdir -p /etc/s6-overlay/s6-rc.d/todo-app/dependencies.d \
             /etc/s6-overlay/s6-rc.d/user/contents.d

COPY run.sh /etc/s6-overlay/s6-rc.d/todo-app/run
RUN chmod a+x /etc/s6-overlay/s6-rc.d/todo-app/run \
    && echo "longrun" > /etc/s6-overlay/s6-rc.d/todo-app/type \
    && touch /etc/s6-overlay/s6-rc.d/user/contents.d/todo-app \
    && touch /etc/s6-overlay/s6-rc.d/todo-app/dependencies.d/base
