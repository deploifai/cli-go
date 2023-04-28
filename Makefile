BINARY_NAME=bin/deploifai

.phony: build
build:
	@echo "Building..."
	@go build -o $(BINARY_NAME) main.go
	@echo "Done."

.phony: generate
generate:
	@echo "Generating graphql api client..."
	@gqlgenc
	@echo "Done."
