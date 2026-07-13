VERSION=0.0.3
GITCOMMIT?=$(shell git describe --dirty --always)
LDFLAGS=-ldflags "-w -s -X main.version=${VERSION} -X main.commit=${GITCOMMIT}"
all: mackerel-plugin-resolver-synthetic

.PHONY: mackerel-plugin-resolver-synthetic

mackerel-plugin-resolver-synthetic: cmd/mackerel-plugin-resolver-synthetic/main.go
	go build $(LDFLAGS) -o mackerel-plugin-resolver-synthetic cmd/mackerel-plugin-resolver-synthetic/main.go

linux: cmd/mackerel-plugin-resolver-synthetic/main.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o mackerel-plugin-resolver-synthetic cmd/mackerel-plugin-resolver-synthetic/main.go

check:
	go test -v ./...

