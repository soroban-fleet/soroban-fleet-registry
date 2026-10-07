.PHONY: all build test vet fmt fmt-check clean migrations-up

BIN_DIR := bin
BINARY := $(BIN_DIR)/sfr

all: fmt-check vet test build

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BINARY) ./cmd/sfr

test:
	go test -v -race -p 1 ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "Files not formatted:"; gofmt -l .; exit 1)

clean:
	rm -rf $(BIN_DIR)
