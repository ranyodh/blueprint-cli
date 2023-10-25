PACKAGES := $(shell go list ./... | grep -v /vendor/)

ARGS :=
.PHONY: default
default:  build run

.PHONY: build
build:  ## build the redpanda worker locally
	CGO_ENABLED=1 go build -o bin/redpanda-driver .

.PHONY: run
run: ## run the redpanda driver binary
	@./run.sh $(ARGS)

.PHONY: lint
lint: ## run linter
	@golangci-lint run

.PHONY: clean
clean: ## remove generated binary
	@rm -rf bin/
