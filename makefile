.PHONY: lint format

lint:
	golangci-lint run --verbose *.go

format:
	goimports -w *.go