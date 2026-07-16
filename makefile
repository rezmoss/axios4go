# Makefile for axios4go 

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet
GOCOVER=$(GOCMD) tool cover

# Test flags
TEST_FLAGS=-v
RACE_FLAGS=-race
COVERAGE_FLAGS=-coverprofile=coverage.out

all: test build

build:
	$(GOBUILD) ./...

test:
	CGO_ENABLED=0 $(GOTEST) $(TEST_FLAGS) ./...

test-race:
	$(GOTEST) $(TEST_FLAGS) $(RACE_FLAGS) ./...

test-coverage:
	CGO_ENABLED=0 $(GOTEST) $(TEST_FLAGS) $(COVERAGE_FLAGS) ./...
	$(GOCOVER) -func=coverage.out

test-all: test test-race test-coverage

benchmark:
	$(GOTEST) -run=^$$ -bench=. -benchmem $(shell go list ./... | grep -v /examples)

clean:
	$(GOCLEAN) ./...
	rm -f coverage.out

deps:
	$(GOMOD) download

# Format all Go files
fmt:
	$(GOFMT) -s -w .

fmt-check:
	test -z "$$(gofmt -l .)"

# Run go vet
vet:
	$(GOVET) ./...

lint:
	golangci-lint run ./...

vuln:
	govulncheck ./...

# Run gocyclo
cyclo:
	@which gocyclo > /dev/null || go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	gocyclo -over 15 .

# Check examples
check-examples:
	@echo "Checking Go files in examples/ ..."
	@find examples -type f -name '*.go' | while read file; do \
		echo "  Checking $$file"; \
		$(GOCMD) build -o /dev/null "$$file" || exit 1; \
	done

# Run all checks and tests
check: fmt-check vet lint test-all check-examples

# Install gocyclo if not present
install-gocyclo:
	@which gocyclo > /dev/null || go install github.com/fzipp/gocyclo/cmd/gocyclo@latest

.PHONY: all build test test-race test-coverage test-all benchmark clean deps fmt fmt-check vet lint vuln cyclo check-examples check install-gocyclo
