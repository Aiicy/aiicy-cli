DIST := dist
IMPORT := github.com/aiicy/aiicy-cli

GO ?= go
SED_INPLACE := sed -i
EXTRA_GOFLAGS ?=
EXTRA_GOENVS ?=
# TAGS ?=
LDFLAGS ?=

ifeq ($(OS), Windows_NT)
	EXECUTABLE_CLI := aiicy-cli.exe
	EXTRA_GOENVS = CGO_ENABLED=0 GOOS=windows GOARCH=amd64
else
	EXECUTABLE_CLI := aiicy-cli
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Darwin)
		EXTRA_GOENVS = CGO_ENABLED=0 GOOS=darwin GOARCH=amd64
	else
		EXTRA_GOENVS = CGO_ENABLED=0 GOOS=linux GOARCH=amd64
	endif
endif

GOFILES := $(shell find . -name "*.go" -type f ! -path "./vendor/*")
GOBINS := ${GOPATH}/bin
GOFMT ?= gofmt -s

GOFLAGS := -mod=vendor -v

PACKAGES ?= $(shell GO111MODULE=on $(GO) list -mod=vendor ./... | grep -v /vendor/)
SOURCES ?= $(shell find . -name "*.go" -type f)

.PHONY: all
all: build

.PHONY: clean
clean:
	$(GO) clean -i ./...
	rm -f $(EXECUTABLE_CLI)

.PHONY: fmt
fmt:
	$(GOFMT) -w $(GOFILES)

.PHONY: fmt-check
fmt-check:
	# get all go files and run go fmt on them
	@diff=$$($(GOFMT) -d $(GOFILES)); \
	if [ -n "$$diff" ]; then \
		echo "Please run 'make fmt' and commit the result:"; \
		echo "$${diff}"; \
		exit 1; \
	fi;

.PHONY: errcheck
errcheck:
	@hash errcheck > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/kisielk/errcheck; \
	fi
	errcheck $(PACKAGES)

.PHONY: revive
revive:
	@hash revive > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/mgechev/revive; \
	fi
	revive -config .revive.toml -exclude=./vendor/... ./... || exit 1

.PHONY: misspell-check
misspell-check:
	@hash misspell > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/client9/misspell/cmd/misspell; \
	fi
	misspell -error -i unknwon,destory $(GOFILES)

.PHONY: misspell
misspell:
	@hash misspell > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/client9/misspell/cmd/misspell; \
	fi
	misspell -w -i unknwon $(GOFILES)

.PHONY: test
test:
	$(GO) test -tags='sqlite sqlite_unlock_notify' $(PACKAGES)

.PHONY: vendor
vendor:
	GO111MODULE=on $(GO) mod tidy && GO111MODULE=on $(GO) mod vendor

.PHONY: test-vendor
test-vendor: vendor
	@diff=$$(git diff vendor/); \
	if [ -n "$$diff" ]; then \
		echo "Please run 'make vendor' and commit the result:"; \
		echo "$${diff}"; \
		exit 1; \
	fi;

.PHONY: build
build: $(EXECUTABLE_CLI)

$(EXECUTABLE_CLI): $(SOURCES)
	GO111MODULE=on $(EXTRA_GOENVS) $(GO) build $(GOFLAGS) $(EXTRA_GOFLAGS) -o $@;
