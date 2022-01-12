# https://suva.sh/posts/well-documented-makefiles/#simple-makefile
.DEFAULT_GOAL:=help
SHELL:=/bin/bash
PKG_NAME:=github.com/productboard/puma-exporter
BUILD_DIR:=bin
EXPORTER_BINARY:=$(BUILD_DIR)/puma_exporter
IMAGE := productboard/puma-exporter
VERSION=1.0.0
LDFLAGS=-s -w -X main.Version=$(VERSION) -X main.GITCOMMIT=`git rev-parse --short HEAD`
CGO_ENABLED=0
GOARCH=amd64
GOOS=linux

.PHONY: help deps clean build

help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

deps:  ## Check dependencies
	go mod tidy

clean: ## Cleanup the project folders
	@rm -rf bin/

build: clean deps ## Build the project
	@mkdir -p $(BUILD_DIR)
	go build -o $(EXPORTER_BINARY) -ldflags="$(LDFLAGS)" $(PKG_NAME)

docker: build ## Create docker image
	docker build -t $(IMAGE):$(VERSION) .

