SHELL = /bin/bash
OS = $(shell uname -s)

# Project variables
PACKAGE = github.com/martonsereg/scheduler
BINARY_NAME = my-custom-scheduler
IMAGE = 4yxy/my-custom-scheduler
TAG = v0.3

# Build variables
BUILD_DIR ?= bin
BUILD_PACKAGE = .
BUILD_DATE ?= $(shell date +%FT%T%z)
export CGO_ENABLED ?= 0

.PHONY: clean
clean: ## Clean the working area and the project
	rm -rf ${BUILD_DIR}/

.PHONY: build
build: clean ## build binary
	CGO_ENABLED=0 go build -o ${BUILD_DIR}/${BINARY_NAME} cmd/main.go

.PHONY: docker-image
docker-build: build ## Builds docker image for the scheduler
	docker build -t $(IMAGE):$(TAG) .

.PHONY: test
test:
	set -o pipefail; go list ./... | xargs -n1 go test ${GOARGS} -v -parallel 1 2>&1 | tee test.txt

.PHONY: list
list: ## List all make targets
	@$(MAKE) -pRrn : -f $(MAKEFILE_LIST) 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | sort

.PHONY: help
.DEFAULT_GOAL := help
help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: run
run:  ## Run a controller from your host.
	 go run cmd/main.go --config=config/kube-scheduler-example-config.yaml

