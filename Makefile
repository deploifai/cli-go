BINARY_DIR=bin
BINARY_NAME=deploifai

.phony: vet
vet:
	go vet ./...

.phony: generate
generate:
	go generate ./...
	gqlgenc

.phony: build
build: vet generate
	@echo "Building for current platform..."
	go build -o $(BINARY_DIR)/$(BINARY_NAME) main.go

.phony: build-all
build-all: vet generate
	@echo "Building for all platforms..."
	goreleaser build --snapshot --clean
