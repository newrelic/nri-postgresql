WORKDIR      := $(shell pwd)
TARGET       := target
TARGET_DIR    = $(WORKDIR)/$(TARGET)
NATIVEOS	 := $(shell go version | awk -F '[ /]' '{print $$4}')
NATIVEARCH	 := $(shell go version | awk -F '[ /]' '{print $$5}')
INTEGRATION  := postgresql
BINARY_NAME   = nr-$(INTEGRATION)
GO_PKGS      := $(shell go list ./... | grep -v "/vendor/")
GO_FILES     := ./src/
GOTOOLS       = github.com/kardianos/govendor \
		gopkg.in/alecthomas/gometalinter.v2 \
		github.com/axw/gocov/gocov \
		github.com/stretchr/testify/assert \
		github.com/AlekSi/gocov-xml \

all: build

build: check-version clean validate test compile

clean:
	@echo "=== $(INTEGRATION) === [ clean ]: Removing binaries and coverage file..."
	@rm -rfv bin coverage.xml $(TARGET)

tools: check-version
	@echo "=== $(INTEGRATION) === [ tools ]: Installing tools required by the project..."
	@go get $(GOTOOLS)
	@gometalinter.v2 --install

tools-update: check-version
	@echo "=== $(INTEGRATION) === [ tools-update ]: Updating tools required by the project..."
	@go get -u $(GOTOOLS)
	@gometalinter.v2 --install

deps: tools deps-only

deps-only:
	@echo "=== $(INTEGRATION) === [ deps ]: Installing package dependencies required by the project..."
	@govendor sync

validate: deps
	@echo "=== $(INTEGRATION) === [ validate ]: Validating source code running gometalinter..."
	@gometalinter.v2 --config=.gometalinter.json ./...

validate-all: deps
	@echo "=== $(INTEGRATION) === [ validate ]: Validating source code running gometalinter..."
	@gometalinter.v2 --config=.gometalinter.json --enable=interfacer --enable=gosimple ./...

compile: deps
	@echo "=== $(INTEGRATION) === [ compile ]: Building $(BINARY_NAME)..."
	@go build -o bin/$(BINARY_NAME) ./src

compile-only: deps-only
	@echo "=== $(INTEGRATION) === [ compile ]: Building $(BINARY_NAME)..."
	@go build -o bin/$(BINARY_NAME) ./src

test: deps
	@echo "=== $(INTEGRATION) === [ test ]: Running unit tests..."
	@gocov test $(GO_PKGS) | gocov-xml > coverage.xml

# Include thematic Makefiles
include Makefile-*.mk

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

.PHONY: all build clean tools tools-update deps validate compile test check-version
