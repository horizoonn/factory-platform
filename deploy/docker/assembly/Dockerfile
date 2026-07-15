FROM golang:1.26.2-alpine AS builder

WORKDIR /app

COPY platform/go.mod platform/go.sum ./platform/
COPY shared/go.mod shared/go.sum ./shared/
COPY assembly/go.mod assembly/go.sum ./assembly/

WORKDIR /app/assembly
RUN GOWORK=off go mod download

WORKDIR /app
COPY platform ./platform
COPY shared ./shared
COPY assembly ./assembly

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /app/assembly
RUN GOWORK=off CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
	go build -trimpath -ldflags="-s -w" -o /out/app-assembly ./cmd

FROM alpine:3.21.3

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder --chown=appuser:appgroup /out/app-assembly ./app-assembly
COPY --from=builder --chown=appuser:appgroup /app/assembly/migrations ./migrations

ENV ASSEMBLY_MIGRATIONS_DIR=/app/migrations

USER appuser

ENTRYPOINT ["./app-assembly"]
