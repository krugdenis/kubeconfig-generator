BINARY_NAME := kubeconfig-generator
OS := $(shell go env GOOS)

.PHONY: build release

build:
	@echo "Building $(BINARY_NAME)..."
	@if [ "$(OS)" = "windows" ]; then \
		GOOS=windows go build -o $(BINARY_NAME).exe; \
	else \
		go build -o $(BINARY_NAME); \
	fi

release:
	@echo "Creating release ..."
	@goreleaser release --clean
