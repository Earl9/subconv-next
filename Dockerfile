FROM golang:1.22 AS build

WORKDIR /src

ARG TARGETOS=linux
ARG TARGETARCH=amd64

COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
COPY static ./static

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w" -o /out/subconv-next ./cmd/subconv-next

FROM alpine:3.20

RUN mkdir -p /config /data

COPY --from=build /out/subconv-next /usr/local/bin/subconv-next

WORKDIR /app

EXPOSE 9876

ENTRYPOINT ["subconv-next"]
CMD ["serve", "--config", "/config/config.json"]
