BINARY          := auth-proxy
ROOT_DIR        := $(if $(ROOT_DIR),$(ROOT_DIR),$(shell git rev-parse --show-toplevel))
BUILD_DIR       := $(ROOT_DIR)/dist
VERSION         := $(shell cat VERSION)
GITSHA          := $(shell git rev-parse --short HEAD)

BUILD_TIME         := $(shell date +'%Y-%m-%d_%T')
GO_OS              := $(if $(GOHOSTOS),$(GOHOSTOS),$(shell go env GOHOSTOS))
GO_ARCH            := $(if $(GOHOSTARCH),$(GOHOSTARCH),$(shell go env GOHOSTARCH))
OS_ARCH            := $(GO_OS)_$(GO_ARCH)
GIT_BRANCH         :=$(shell git rev-parse --abbrev-ref HEAD)
GIT_REVISION       :=$(shell git rev-list -1 HEAD)
GIT_REVISION_DIRTY :=$(shell (git diff-index --quiet HEAD -- . && git diff --staged --quiet -- .) || echo "-dirty")
GO_LINT_CHECKS     := govet ineffassign staticcheck deadcode unused

.PHONY: build clean start lint staticcheck test fmt release-test release

prepare:
	mkdir -p $(BUILD_DIR)

lint:
	$(GO_LINT_HEAD) $(GO_ENV_VARS) golangci-lint run --disable-all $(foreach check,$(GO_LINT_CHECKS), -E $(check)) $(foreach issue,$(GO_LINT_EXCLUDE_ISSUES), -e $(issue)) $(GO_LINT_TRAIL)

check:
	go get -u honnef.co/go/tools/cmd/staticcheck
	$(shell go list -f {{.Target}} honnef.co/go/tools/cmd/staticcheck) ./...

sec:
	go get -u github.com/securego/gosec/v2/cmd/gosec
	$(shell go list -f {{.Target}} github.com/securego/gosec/v2/cmd/gosec) -fmt=golint ./...

test: prepare
	go test -short -race -coverprofile=$(BUILD_DIR)/cover.out ./...

build: prepare
	goreleaser build --snapshot --rm-dist --single-target

start:
	go run $(ROOT_DIR)/cmd/$(BINARY)/main.go

profile:
	go tool pprof -http=:7777 cpuprofile

clean:
	rm -rf $(BUILD_DIR)

release-test: export GITHUB_SHA=$(GITSHA)
release-test:
	goreleaser release --skip-publish --snapshot --rm-dist

release: export GITHUB_SHA=$(GITSHA)
release: release-test
	git tag -a $(VERSION) -m "Release" && git push origin $(VERSION)
