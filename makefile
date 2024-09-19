# Makefile for axios4go 

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet
GOCOVER=$(GOCMD) tool cover

# Binary name
BINARY_NAME=axios4go

# Test flags
TEST_FLAGS=-v
RACE_FLAGS=-race
COVERAGE_FLAGS=-coverprofile=coverage.out

all: test build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v

test:
	$(GOTEST) $(TEST_FLAGS) $(shell go list ./... | grep -v /examples)

test-race:
	$(GOTEST) $(TEST_FLAGS) $(RACE_FLAGS) $(shell go list ./... | grep -v /examples)

test-coverage:
	$(GOTEST) $(TEST_FLAGS) $(COVERAGE_FLAGS) $(shell go list ./... | grep -v /examples)
	$(GOCOVER) -func=coverage.out

test-all: test test-race test-coverage

benchmark:
	$(GOTEST) -run=^$$ -bench=. -benchmem $(shell go list ./... | grep -v /examples)

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME) coverage.out

run:
	$(GOBUILD) -o $(BINARY_NAME) -v
	./$(BINARY_NAME)

deps:
	$(GOGET) -v -t -d ./...
	$(GOMOD) tidy

# Format all Go files
fmt:
	$(GOFMT) -s -w .

# Run go vet
vet:
	$(GOVET) ./...

# Run gocyclo
cyclo:
	@which gocyclo > /dev/null || go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	gocyclo -over 15 .

# Check examples
check-examples:
	@for file in examples/*.go; do \
		echo "Checking $$file"; \
		$(GOCMD) build -o /dev/null $$file || exit 1; \
	done

# Run all checks and tests
check: fmt vet cyclo test-all check-examples

# Install gocyclo if not present
install-gocyclo:
	@which gocyclo > /dev/null || go install github.com/fzipp/gocyclo/cmd/gocyclo@latest

.PHONY: all build test test-race test-coverage test-all benchmark clean run deps fmt vet cyclo check-examples check install-gocyclo