BINARY_DIR=bin
BINARY_NAME=deploifai

.phony: build
build:
	@echo "Building for current platform..."
	@go build -o $(BINARY_DIR)/$(BINARY_NAME) main.go
	@echo "Done"

.phony: generate
generate:
	@echo "Generating go code..."
	@go generate ./...
	@echo "Generating graphql api client..."
	@gqlgenc
	@echo "Done"

.phony: build-all
build-all:
	@echo "Building for all platforms..."
	@goreleaser build --snapshot --clean
