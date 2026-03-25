ARG BUILD_FROM=golang:1.22-alpine
FROM ${BUILD_FROM} AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /todo-app .

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /todo-app /usr/local/bin/todo-app

COPY run.sh /
RUN chmod a+x /run.sh

CMD ["/run.sh"]
