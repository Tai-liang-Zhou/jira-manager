BINARY_NAME=jira-mcp-server
GO=go

.PHONY: build test run fmt vet

## build: Build the server binary
build:
	$(GO) build -o bin/$(BINARY_NAME) ./cmd/server

## test: Run all tests with race detection
test:
	$(GO) test -race ./...

## run: Build and run the server
run: build
	./bin/$(BINARY_NAME)

## fmt: Format code
fmt:
	$(GO) fmt ./...

## vet: Run go vet
vet:
	$(GO) vet ./...
