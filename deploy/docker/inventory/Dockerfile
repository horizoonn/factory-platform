# syntax=docker/dockerfile:1

FROM golang:1.26.2-alpine AS builder

WORKDIR /app

COPY platform/go.mod platform/go.sum ./platform/
COPY shared/go.mod shared/go.sum ./shared/
COPY inventory/go.mod inventory/go.sum ./inventory/

WORKDIR /app/inventory
RUN GOWORK=off go mod download

WORKDIR /app
COPY platform ./platform
COPY shared ./shared
COPY inventory ./inventory

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG GRPC_HEALTH_PROBE_VERSION=v0.4.53

WORKDIR /app/inventory
RUN GOWORK=off CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
	go build -trimpath -ldflags="-s -w" -o /out/app-inventory ./cmd

RUN GOWORK=off CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
	GOBIN=/out go install github.com/grpc-ecosystem/grpc-health-probe@$GRPC_HEALTH_PROBE_VERSION

FROM alpine:3.21.3
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder --chown=appuser:appgroup /out/app-inventory ./app-inventory
COPY --from=builder --chown=appuser:appgroup /app/inventory/migrations ./migrations
COPY --from=builder /out/grpc-health-probe /bin/grpc-health-probe

ENV INVENTORY_GRPC_HOST=0.0.0.0 \
    INVENTORY_GRPC_PORT=50051 \
    INVENTORY_HTTP_HOST=0.0.0.0 \
    INVENTORY_HTTP_PORT=8082 \
    INVENTORY_MIGRATIONS_DIR=/app/migrations

EXPOSE 50051 8082

USER appuser

HEALTHCHECK --interval=10s --timeout=3s --retries=5 \
	CMD ["/bin/grpc-health-probe", "-addr=:50051"]

ENTRYPOINT ["./app-inventory"]
