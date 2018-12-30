RELEASE_DIR ?= release
VENDOR_DIR ?= vendor

GO = go
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# Build the executable
$(RELEASE_DIR)/%:
	@GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -o "$@" -v $(GO_FLAGS) $(LD_FLAGS) $(MAIN)

# aggregator
aggregator: MAIN=./cmd/aggregator/main.go
aggregator: $(RELEASE_DIR)/aggregator-$(GOOS)-$(GOARCH)

build: aggregator

clean:
	@rm -rf $(RELEASE_DIR)

TEST_PKGS := $(shell go list ./... | grep -v /$(VENDOR_DIR))

.PHONY: test
test:
	@$(GO) test $(TEST_PKGS)
