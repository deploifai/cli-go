BINARY_DIR=bin
BINARY_NAME=deploifai

.phony: fmt
fmt:
	go fmt ./...

.phony: vet
vet:
	go vet ./...

.phony: generate
generate:
	go generate ./...

.phony: build
build: fmt generate vet
	@echo "Building for current platform..."
	go build -o $(BINARY_DIR)/$(BINARY_NAME) main.go

.phony: build-all
build-all: fmt generate vet
	@echo "Building for all platforms in dist/ ..."
	goreleaser build --snapshot --clean
