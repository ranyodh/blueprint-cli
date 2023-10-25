.PHONY: default
default:  build

.PHONY: build
build:  ## build charts
	@docker build --output=./build/bundle --target=charts-bundle .
	@docker build --output=./build/helm   --target=charts-helm .

