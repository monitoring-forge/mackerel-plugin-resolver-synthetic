VERSION=0.0.6
LDFLAGS=-ldflags "-w -s -X main.version=${VERSION}"
all: mackerel-plugin-resolver-synthetic

.PHONY: mackerel-plugin-resolver-synthetic linux check lint

mackerel-plugin-resolver-synthetic: cmd/mackerel-plugin-resolver-synthetic/*.go
	go build $(LDFLAGS) -o mackerel-plugin-resolver-synthetic ./cmd/mackerel-plugin-resolver-synthetic

linux: cmd/mackerel-plugin-resolver-synthetic/*.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o mackerel-plugin-resolver-synthetic ./cmd/mackerel-plugin-resolver-synthetic

check:
	go test -v ./...
	go test -race ./...

lint:
	golangci-lint run --timeout 5m ./...

