BINARY_DIR=bin
BINARY_NAME=deploifai

.phony: vet
vet:
	go vet ./...

.phony: generate
generate:
	go install github.com/Yamashou/gqlgenc@latest
	gqlgenc
	go generate ./...

.phony: build
build: vet generate
	@echo "Building for current platform..."
	go build -o $(BINARY_DIR)/$(BINARY_NAME) main.go

.phony: build-all
build-all: vet generate
	@echo "Building for all platforms in dist/ ..."
	goreleaser build --snapshot --clean
