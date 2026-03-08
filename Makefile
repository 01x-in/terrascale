BINARY_NAME=terrascale
BUILD_DIR=.

.PHONY: build test clean install lint

build:
	go build -o $(BINARY_NAME) ./cmd/terrascale/

test:
	go test ./... -v

clean:
	rm -f $(BINARY_NAME)
	go clean

install:
	go install ./cmd/terrascale/

lint:
	golangci-lint run ./...
