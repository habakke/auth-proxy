VERSION=$(shell (git describe --abbrev=0 --tags || echo '0.1.0') 2>/dev/null)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
HEAD=$(shell (git rev-list -1 HEAD || echo 'git') | tr -d '\n')
SEMVER=${VERSION}+${HEAD}-${BRANCH}

REPO := ghcr.io/habakke
RELDIR   := $(subst $(TOPD)/,,$(shell pwd))

ifneq ($(C), 0)
C_Y=$(shell tput setaf 3)
C_R=$(shell tput setaf 1)
C_BLD=$(shell tput bold)
C_RST=$(shell tput sgr0)
endif

DOCKERFILE         := Dockerfile
GO_ENV_VARS        := $(if $(GO_ENV_VARS),$(GO_ENV_VARS),)
GO_RUN_ENV_VARS    := $(if $(GO_ENV_VARS),$(GO_ENV_VARS),CGO_ENABLED=0)
GO_RUN             := $(GO_RUN_ENV_VARS) go
GO_BUILD_ENV_VARS  := $(if $(GO_ENV_VARS),$(GO_ENV_VARS),CGO_ENABLED=0 GOOS=linux GOARCH=amd64)
GO_BUILD           := $(GO_BUILD_ENV_VARS) go
GO_TEST            := go
GO_BUILD_ARGS      += --ldflags "-X internal/config.appVersion=${SEMVER} -X internal/config.appBuildTime=${BUILD_TIME} -X internal/config.appBuildUser=${USER}"
GO_BUILD_ARGS      += $(if $(GO_GCFLAGS),-gcflags "$(GO_GCFLAGS)",)
GO_BUILD_ARGS      += -o $(RELDIR)

define buildstep
	@printf '$(C_BLD)$(C_Y)%-20s$(C_RST) $(C_BLD)%16s$(C_RST) %s\n' "$@" "$(1) $(2)" "$(RELDIR)/cmd/$(2)"
endef