BIN ?= $(HOME)/bin

install:
	go build -o $(BIN)/kustomize ./cmd/kustomize