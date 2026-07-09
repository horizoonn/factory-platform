FROM golang:1.26.2-alpine AS builder

WORKDIR /app

COPY platform/go.mod platform/go.sum ./platform/
COPY shared/go.mod shared/go.sum ./shared/
COPY payment/go.mod payment/go.sum ./payment/

WORKDIR /app/payment
RUN GOWORK=off go mod download

WORKDIR /app
COPY platform ./platform
COPY shared ./shared
COPY payment ./payment

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG GRPC_HEALTH_PROBE_VERSION=v0.4.53

WORKDIR /app/payment
RUN GOWORK=off CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
	go build -trimpath -ldflags="-s -w" -o /out/app-payment ./cmd

RUN GOWORK=off CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
	GOBIN=/out go install github.com/grpc-ecosystem/grpc-health-probe@$GRPC_HEALTH_PROBE_VERSION

FROM alpine:3.21.3

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder --chown=appuser:appgroup /out/app-payment ./app-payment
COPY --from=builder /out/grpc-health-probe /bin/grpc-health-probe

ENV PAYMENT_GRPC_HOST=0.0.0.0 \
    PAYMENT_GRPC_PORT=50052

EXPOSE 50052

USER appuser

HEALTHCHECK --interval=10s --timeout=3s --retries=5 \
	CMD ["/bin/grpc-health-probe", "-addr=:50052"]

ENTRYPOINT ["./app-payment"]
