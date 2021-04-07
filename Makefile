BINARY          := auth-proxy
ROOT_DIR        := $(if $(ROOT_DIR),$(ROOT_DIR),$(shell git rev-parse --show-toplevel))
BUILD_DIR       := $(ROOT_DIR)/dist
VERSION         := $(shell cat ./VERSION)
GITSHA          := $(shell git rev-parse --short HEAD)

.PHONY: build clean start test fmt release

prepare:
	mkdir -p $(BUILD_DIR)

test: prepare
	go test -v -coverprofile=$(BUILD_DIR)/cover.out ./...

build: prepare
	goreleaser build --snapshot --rm-dist

start:
	go run $(ROOT_DIR)/cmd/$(BINARY)/main.go

profile:
	go tool pprof -http=:7777 cpuprofile

clean:
	rm -rf $(BUILD_DIR)

fmt:
	go fmt ./...

release: export GITHUB_SHA=$(GITSHA)
release:
	goreleaser release --skip-publish --snapshot --rm-dist && git tag -a $(VERSION) -m "Release" && git push origin $(VERSION)
