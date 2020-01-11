SHELL = bash
GO = go
DOCKER = docker
GOBIN = $(PWD)/.bin
export GOBIN

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

RELEASE_DIR = dist
DOCKER_TAG_PREFIX = emando/vantage-events-

.PHONY: deps.dev
deps.dev:
	@cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % $(GO) install %

.PHONY: deps.tidy
deps.tidy:
	@$(GO) mod tidy

.PHONY: fmt
fmt:
	@$(GOBIN)/gofumports -w .

.PHONY: quality
quality:
	@$(GOBIN)/golint -set_exit_status ./... && \
		$(GO) vet ./...

.PHONY: test
test:
	@$(GO) test ./...

.PHONY: test.race
test.race:
	@$(GO) test -race -covermode=atomic ./...

.PHONY: test.cover
test.cover:
	@$(GO) test -cover ./...

.PHONY: git.nodiff
git.nodiff:
	@if [[ ! -z "`git diff`" ]]; then \
		git diff; \
		exit 1; \
	fi

$(RELEASE_DIR)/%:
	@GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -o "$@" $(GO_FLAGS) $(LD_FLAGS) $(MAIN)

.PHONY: aggregator
aggregator: MAIN=./cmd/aggregator/main.go
aggregator: $(RELEASE_DIR)/aggregator-$(GOOS)-$(GOARCH)

.PHONY: build
build: aggregator

.PHONY: docker
docker: GOOS=linux
docker: GOARCH=amd64
docker: MAIN=./cmd/aggregator/main.go
docker: $(RELEASE_DIR)/aggregator-linux-amd64
	@$(DOCKER) build -t $(addsuffix "aggregator", $(DOCKER_TAG_PREFIX)) -f build/aggregator.Dockerfile .

.PHONY: clean
clean:
	@rm -rf $(RELEASE_DIR)

.PHONY: certs.dev
certs.dev:
	@mkdir -p .dev
	@CAROOT=.dev $(GOBIN)/mkcert -cert-file .dev/cert.pem -key-file .dev/key.pem localhost

# vim: ft=make
