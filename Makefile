BINARY := subconv-next

.PHONY: build test test-race vet fmt-check package docker-buildx

build:
	go build -o $(BINARY) ./cmd/subconv-next

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

fmt-check:
	@test -z "$$(gofmt -l $$(find . -path './.git' -prune -o -path './.tmpinspect' -prune -o -path './vendor' -prune -o -name '*.go' -print))"

package:
	./scripts/package-openwrt-ipk-portable.sh

docker-buildx:
	docker buildx build --platform linux/amd64,linux/arm64 -t subconv-next:local .
