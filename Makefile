export PATH := $(PATH):$(GOPATH)/bin
NATIVEOS    := $(shell go version | awk -F '[ /]' '{print $$4}')
NATIVEARCH  := $(shell go version | awk -F '[ /]' '{print $$5}')
INTEGRATION     := postgresql
GOFLAGS          = -mod=readonly # ignore the vendor directory and to report an error if go.mod needs to be updated.
BINARY_NAME      = nri-$(INTEGRATION)
INTEGRATIONS_DIR = /var/db/newrelic-infra/newrelic-integrations/
CONFIG_DIR       = /etc/newrelic-infra/integrations.d
GO_FILES        := ./src/
GOLANGCI_LINT	 = github.com/golangci/golangci-lint/cmd/golangci-lint
GOCOV            = github.com/axw/gocov/gocov
GOCOV_XML		 = github.com/AlekSi/gocov-xml

all: build

build: clean validate test compile

build-container:
	docker build -t nri-postgresql .

clean:
	@echo "=== $(INTEGRATION) === [ clean ]: Removing binaries and coverage file..."
	@rm -rfv bin coverage.xml


format:
	sh scripts/format.sh

validate:
ifeq ($(strip $(GO_FILES)),)
	@echo "=== $(INTEGRATION) === [ validate ]: no Go files found. Skipping validation."
else
	@printf "=== $(INTEGRATION) === [ validate ]: running golangci-lint & semgrep... "
	go run  $(GOFLAGS) github.com/golangci/golangci-lint/cmd/golangci-lint run --verbose
	docker run --rm -v "${PWD}:/src:ro" --workdir /src returntocorp/semgrep -c .semgrep.yml
endif


bin/$(BINARY_NAME):
	@echo "=== $(INTEGRATION) === [ compile ]: building $(BINARY_NAME)..."
	@go build -v -o bin/$(BINARY_NAME) cmd/nri-postgresql/main.go


compile: bin/$(BINARY_NAME)

test:
	@echo "=== $(INTEGRATION) === [ test ]: running unit tests..."
	@go run $(GOFLAGS) $(GOCOV) test ./... | go run $(GOFLAGS) $(GOCOV_XML) > coverage.xml


integration-test:
	@echo "=== $(INTEGRATION) === [ test ]: running integration tests..."
	@docker-compose -f tests/docker-compose.yml pull
	@go test -v -tags=integration ./tests/. || (ret=$$?; docker-compose -f tests/docker-compose.yml down && exit $$ret)
	@docker-compose -f tests/docker-compose.yml down

install: compile
	@echo "=== $(INTEGRATION) === [ install ]: installing bin/$(BINARY_NAME)..."
	@sudo install -D --mode=755 --owner=root --strip $(ROOT)bin/$(BINARY_NAME) $(INTEGRATIONS_DIR)/bin/$(BINARY_NAME)
	@sudo install -D --mode=644 --owner=root $(ROOT)$(INTEGRATION)-definition.yml $(INTEGRATIONS_DIR)/$(INTEGRATION)-definition.yml
	@sudo install -D --mode=644 --owner=root $(ROOT)$(INTEGRATION)-config.yml.sample $(CONFIG_DIR)/$(INTEGRATION)-config.yml.sample

# Include thematic Makefiles
include $(CURDIR)/build/ci.mk
include $(CURDIR)/build/release.mk


check-version:
ifdef GOOS
ifneq "$(GOOS)" "$(NATIVEOS)"
	$(error GOOS is not $(NATIVEOS). Cross-compiling is only allowed for 'clean', 'deps-only' and 'compile-only' targets)
endif
endif
ifdef GOARCH
ifneq "$(GOARCH)" "$(NATIVEARCH)"
	$(error GOARCH variable is not $(NATIVEARCH). Cross-compiling is only allowed for 'clean', 'deps-only' and 'compile-only' targets)
endif
endif

.PHONY: all build clean validate compile test integration-test install