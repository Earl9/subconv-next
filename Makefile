BINARY := subconv-next

.PHONY: build test

build:
	go build -o $(BINARY) ./cmd/subconv-next

test:
	go test ./...
