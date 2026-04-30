FROM --platform=$BUILDPLATFORM golang:1.22 AS build

WORKDIR /src

ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

COPY go.mod go.sum ./
COPY cmd ./cmd
COPY internal ./internal
COPY static ./static
COPY testdata ./testdata

RUN go test ./...
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -trimpath -ldflags="-s -w" -o /out/subconv-next ./cmd/subconv-next

FROM --platform=$BUILDPLATFORM alpine:3.20 AS runtime-deps

RUN apk add --no-cache ca-certificates tzdata

FROM alpine:3.20

COPY --from=runtime-deps /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=runtime-deps /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /out/subconv-next /usr/bin/subconv-next
COPY config /config

WORKDIR /app

EXPOSE 9876
VOLUME ["/data"]

ENV SUBCONV_HOST=0.0.0.0 \
	SUBCONV_PORT=9876 \
	SUBCONV_DATA_DIR=/data \
	SUBCONV_PUBLIC_BASE_URL= \
	SUBCONV_LOG_LEVEL=info

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
	CMD wget -qO- "http://127.0.0.1:${SUBCONV_PORT}/healthz" >/dev/null || exit 1

ENTRYPOINT ["/usr/bin/subconv-next"]
CMD ["serve", "--config", "/config/config.json"]
