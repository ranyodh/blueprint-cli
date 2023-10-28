
.PHONY: default
default:  build

BIN_DIR := $(shell pwd)/bin

.PHONY: build
build:  ## build locally
	@go mod download
	@CGO_ENABLED=1 go build -o ${BIN_DIR}/bocli ./

.PHONY: install
install:  ## install locally
	@go mod download
	@CGO_ENABLED=1 go build -o ${GOPATH}/bin/bocli ./

.PHONY: init
init:
	@${BIN_DIR}/bocli init

.PHONY: apply
apply:
	@${BIN_DIR}/bocli apply --config blueprint.yaml

.PHONY: update
update:
	@${BIN_DIR}/bocli update --config blueprint.yaml

.PHONY: reset
reset:
	@${BIN_DIR}/bocli reset --config blueprint.yaml

.PHONY: build-charts
build-charts:
	@cd ./charts && make build
