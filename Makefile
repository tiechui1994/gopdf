GOFMT ?= gofmt "-s"
PACKAGES ?= $(shell go list -buildvcs=false ./... | grep -v /vendor/)
GOFILES := $(shell find . -name "*.go" -type f -not -path "./vendor/*")
GONOTESTFILES := $(shell find . -name "*.go" -type f -not -path "./vendor/*"|grep -v _test.go)
GONODOCFILES := $(shell find . -name "*.go" -type f -not -path "./vendor/*"|grep -v doc.go)
GOTESTFILES := $(shell find . -name "*_test.go" -type f -not -path "./vendor/*")

.PHONY: all
all: alltest

.PHONY: alltest
alltest: fmt vet misspell tidy test lint

.PHONY: allcheck
allcheck: fmt-check vet misspell-check test lint

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: test
test:
	@if [ ! -d "./sonar" ]; then \
		mkdir sonar; \
	fi;
	go test -buildvcs=false -covermode=atomic -coverprofile=sonar/cover.out $(PACKAGES)
	go tool cover -func=sonar/cover.out

.PHONY: testout
testout:test
	go tool cover -html=sonar/cover.out

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

vet:
	go vet $(PACKAGES)

.PHONY: misspell-check
misspell-check: tools
	misspell -error $(GOFILES)

.PHONY: misspell
misspell:
	misspell -w $(GOFILES)

.PHONY: tools
tools:
	@hash misspell > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install github.com/client9/misspell/cmd/misspell; \
	fi
	@hash golangci-lint > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint; \
	fi

.PHONY: lint
lint: tools
	# run golanci-lint
	golangci-lint run
