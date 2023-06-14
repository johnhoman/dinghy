BIN ?= $(HOME)/bin

install:
	go build -o $(BIN)/dinghy ./cmd/dinghy

test:
	go test ./...
