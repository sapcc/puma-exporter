PKG_NAME:=github.com/sapcc/puma-exporter
BUILD_DIR:=bin
EXPORTER_BINARY:=$(BUILD_DIR)/puma_exporter
IMAGE := keppel.eu-de-1.cloud.sap/ccloud/puma-exporter
VERSION=1.0.7
LDFLAGS=-s -w -X main.Version=$(VERSION) -X main.GITCOMMIT=`git rev-parse --short HEAD`
CGO_ENABLED=0
GOARCH=amd64
.PHONY: help
help:
	@echo
	@echo "Available targets:"
	@echo "  * build             - build the binary, output to $(ARC_BINARY)"
	@echo "  * linux             - build the binary, output to $(ARC_BINARY)"
	@echo "  * docker            - build docker image"

.PHONY: build
build:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -o $(EXPORTER_BINARY) -ldflags="$(LDFLAGS)" $(PKG_NAME)

linux: export GOOS=linux
linux: export GOARCH=amd64
linux: build

docker: linux
	docker build -t $(IMAGE):$(VERSION) .

push:
	docker push $(IMAGE):$(VERSION)
