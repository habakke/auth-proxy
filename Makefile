SVC=auth-proxy
TOPD := $(if $(TOPD),$(TOPD),$(shell git rev-parse --show-toplevel))
include ${TOPD}/base.mk

prepare::
	@:
build:: prepare
	$(call buildstep,go build,$(SVC))
	@$(GO_BUILD) build $(GO_BUILD_ARGS) ./cmd/$(SVC)
test:: prepare
	@$(GO_BUILD) test ./...
run:: build
	@$(GO_RUN) run ./cmd/$(SVC)
docker-build:: build $(DOCKERFILE)
	@docker build \
		--file "${DOCKERFILE}" \
		-t "${REPO}/${SVC}:$(HEAD)" \
		-t "${REPO}/${SVC}" .
docker-upload::
	@docker image tag "${REPO}/${SVC}" latest
	@docker push --all-tags "${REPO}/${SVC}"
docker-run:: docker-build
	@docker run \
    -e TARGET=$(TARGET) \
    -e COOKIE_SEED=$(COOKIE_SEED) \
    -e COOKIE_KEY=$(COOKIE_KEY) \
    -e TOKEN=$(TOKEN) \
    "${REPO}/${SVC}"

clean:
	@rm $(RELDIR)/$(SVC)