.PHONY: all test bench clean proto proto-gen proto-lint proto-breaking buf-push help

# Variables
GO_PACKAGES := ./...
BUF_AGAINST := buf.build/tsdkv/pathmatch
BUF_TEMPLATE := ./buf.gen.yaml

all: test bench 

help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*?## "}; /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Go targets
test: ## Run all go tests
	go test -v $(GO_PACKAGES)

bench: ## Run all go benchmarks
	go test -bench=. -benchtime=5s $(GO_PACKAGES)

clean: ## Clean go test cache
	go clean -testcache

# Protobuf targets
proto: proto-gen proto-lint proto-breaking ## Run all proto checks and generate code

proto-gen: ## Generate code from proto files
	@buf generate --template $(BUF_TEMPLATE)

proto-lint: ## Lint proto files
	@buf lint

proto-breaking: ## Check for breaking changes in proto files
	@buf breaking --against=$(BUF_AGAINST)

buf-push: ## Push proto files to buf.build
	@buf push

