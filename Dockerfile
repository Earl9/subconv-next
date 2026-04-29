FROM --platform=$BUILDPLATFORM golang:1.22 AS build

WORKDIR /src

ARG BUILDPLATFORM
ARG TARGETOS=linux
ARG TARGETARCH=amd64

COPY go.mod go.sum ./
COPY cmd ./cmd
COPY internal ./internal
COPY static ./static

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w" -o /out/subconv-next ./cmd/subconv-next

FROM alpine:3.20

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out/subconv-next /usr/local/bin/subconv-next

WORKDIR /app

EXPOSE 9876
VOLUME ["/data"]

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
	CMD wget -qO- http://127.0.0.1:9876/healthz >/dev/null || exit 1

ENTRYPOINT ["subconv-next"]
CMD ["serve", "--config", "/config/config.json"]
