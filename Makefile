CMD=auth-proxy
BINARY=auth-proxy
IMAGE=auth-proxy
ROOT_DIR        := $(if $(ROOT_DIR),$(ROOT_DIR),$(shell git rev-parse --show-toplevel))
BUILD_DIR       := $(ROOT_DIR)/build
VERSION         := $(shell cat ./VERSION)

.PHONY: build clean start test fmt release

prepare:
	mkdir -p $(BUILD_DIR)

test: export CGO_ENABLED=0
test: prepare
	 go test -v -coverprofile=$(BUILD_DIR)/cover.out ./...

build: export CGO_ENABLED=0
build: prepare
	go build -o $(BUILD_DIR)/$(BINARY) -a -ldflags '-extldflags "-static"' .

start: build
	go run $(ROOT_DIR)/main.go

clean:
	rm -rf $(BUILD_DIR)

fmt:
	go fmt ./...

release:
	git tag -a $(VERSION) -m "Release" || true
	git push origin $(VERSION)
