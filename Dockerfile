FROM --platform=$BUILDPLATFORM golang:1.27.3-alpine AS build

WORKDIR /src

ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev

COPY go.mod go.sum ./
COPY cmd ./cmd
COPY internal ./internal
COPY static ./static
COPY testdata ./testdata

RUN GOMAXPROCS=2 go test -p 2 ./...
RUN GOMAXPROCS=2 CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -p 2 -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /out/subconv-next ./cmd/subconv-next

FROM --platform=$BUILDPLATFORM alpine:3.23 AS runtime-deps

RUN apk add --no-cache ca-certificates tzdata

FROM alpine:3.23

ARG VERSION=dev
ARG SOURCE_REPOSITORY=""

LABEL org.opencontainers.image.title="SubConv Next" \
	org.opencontainers.image.description="Modern subscription converter for Mihomo / Clash Meta" \
	org.opencontainers.image.source="${SOURCE_REPOSITORY}" \
	org.opencontainers.image.version="${VERSION}" \
	org.opencontainers.image.licenses="MIT"

RUN addgroup -S -g 10001 subconv \
	&& adduser -S -D -H -u 10001 -G subconv subconv \
	&& install -d -o subconv -g subconv -m 0700 /data \
	&& install -d -o root -g subconv -m 0550 /config

COPY --from=runtime-deps /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=runtime-deps /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /out/subconv-next /usr/bin/subconv-next

WORKDIR /app

EXPOSE 9876
VOLUME ["/data"]

ENV SUBCONV_HOST=0.0.0.0 \
	SUBCONV_PORT=9876 \
	SUBCONV_DATA_DIR=/data \
	SUBCONV_PUBLIC_BASE_URL= \
	SUBCONV_LOG_LEVEL=info

USER subconv:subconv

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
	CMD wget -qO- "http://127.0.0.1:${SUBCONV_PORT}/healthz" >/dev/null || exit 1

ENTRYPOINT ["/usr/bin/subconv-next"]
CMD ["serve", "--config", "/config/config.json"]
