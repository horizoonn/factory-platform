FROM golang:1.26.2-alpine AS builder

WORKDIR /app

COPY platform/go.mod platform/go.sum ./platform/
COPY shared/go.mod shared/go.sum ./shared/
COPY order/go.mod order/go.sum ./order/

WORKDIR /app/order
RUN GOWORK=off go mod download

WORKDIR /app
COPY platform ./platform
COPY shared ./shared
COPY order ./order

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /app/order
RUN GOWORK=off CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
	go build -trimpath -ldflags="-s -w" -o /out/app-order ./cmd

FROM alpine:3.21.3

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder --chown=appuser:appgroup /out/app-order ./app-order
COPY --from=builder --chown=appuser:appgroup /app/order/migrations ./migrations

ENV ORDER_HTTP_HOST=0.0.0.0 \
    ORDER_HTTP_PORT=8080 \
    ORDER_MIGRATIONS_DIR=/app/migrations

EXPOSE 8080

USER appuser

HEALTHCHECK --interval=10s --timeout=3s --retries=5 \
	CMD wget -qO- http://127.0.0.1:8080/health >/dev/null || exit 1

ENTRYPOINT ["./app-order"]
