BINARY          := auth-proxy
ROOT_DIR        := $(if $(ROOT_DIR),$(ROOT_DIR),$(shell git rev-parse --show-toplevel))
BUILD_DIR       := $(ROOT_DIR)/dist
VERSION         := $(shell cat ./VERSION)
GIT_SHA         := $(shell git rev-parse --short HEAD)

.PHONY: build clean start test fmt release

prepare:
	mkdir -p $(BUILD_DIR)

test: prepare
	go test -v -coverprofile=$(BUILD_DIR)/cover.out ./...

build: prepare
	goreleaser build --snapshot --rm-dist

start:
	go run $(ROOT_DIR)/cmd/$(BINARY)/main.go

clean:
	rm -rf $(BUILD_DIR)

fmt:
	go fmt ./...

release: export GITHUB_SHA=$(GIT_SHA)
release:
	goreleaser release --skip-publish --snapshot --rm-dist && git tag -a $(VERSION) -m "Release" && git push origin $(VERSION)
