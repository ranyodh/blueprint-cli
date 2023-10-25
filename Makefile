
.PHONY: default
default:  build

BIN_DIR := $(shell pwd)/bin

.PHONY: build
build:  ## build locally
	@go mod download
	@CGO_ENABLED=1 go build -o ${BIN_DIR}/boctl ./

.PHONY: install
install:  ## install locally
	@go mod download
	@CGO_ENABLED=1 go build -o ${GOPATH}/bin/boctl ./

.PHONY: init
init:
	@${BIN_DIR}/boctl init

.PHONY: apply
apply:
	@${BIN_DIR}/boctl apply --config blueprint.yaml

.PHONY: update
update:
	@${BIN_DIR}/boctl update --config blueprint.yaml

.PHONY: reset
reset:
	@${BIN_DIR}/boctl reset --config blueprint.yaml

.PHONY: build-charts
build-charts:
	@cd ./charts && make build
