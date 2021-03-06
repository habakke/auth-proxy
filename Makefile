BINARY          := auth-proxy
ROOT_DIR        := $(if $(ROOT_DIR),$(ROOT_DIR),$(shell git rev-parse --show-toplevel))
BUILD_DIR       := $(ROOT_DIR)/dist
VERSION         := $(shell cat ./VERSION)
GITSHA          := $(shell git rev-parse --short HEAD)

.PHONY: build clean start lint staticcheck test fmt release-test release

prepare:
	mkdir -p $(BUILD_DIR)

check:
	golangci-lint run --fast

sec:
	go get -u github.com/securego/gosec/v2/cmd/gosec
	$(shell go list -f {{.Target}} github.com/securego/gosec/v2/cmd/gosec) -fmt=golint ./...

test: prepare
	go test -short -race -coverprofile=$(BUILD_DIR)/cover.out ./...

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

release-test: export GITHUB_SHA=$(GITSHA)
release-test:
	goreleaser release --skip-publish --snapshot --rm-dist

release: export GITHUB_SHA=$(GITSHA)
release: test release-test
	git tag -a $(VERSION) -m "Release" && git push origin $(VERSION)
