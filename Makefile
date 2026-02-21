.PHONY: test lint cover build examples vet clean

## Run all tests
test:
	go test ./... -count=1 -timeout 120s

## Run tests with verbose output
test-v:
	go test ./... -count=1 -timeout 120s -v

## Run tests with coverage report
cover:
	go test ./... -coverprofile=coverage.out -count=1 -timeout 120s
	go tool cover -func=coverage.out
	@echo ""
	@echo "To view HTML coverage report: go tool cover -html=coverage.out"

## Run go vet
vet:
	go vet ./...

## Run golangci-lint
lint:
	golangci-lint run ./...

## Build all packages (including examples)
build:
	go build ./...

## Build only the library (no examples)
build-lib:
	go build .

## Verify examples compile
examples:
	@for dir in examples/*/; do \
		echo "Building $$dir..."; \
		go build ./$$dir || exit 1; \
	done
	@echo "All examples compile successfully."

## Run all checks (vet, lint, test)
check: vet lint test

## Clean generated files
clean:
	rm -f coverage.out

## Show help
help:
	@echo "Available targets:"
	@echo "  make test       - Run all tests"
	@echo "  make test-v     - Run tests with verbose output"
	@echo "  make cover      - Run tests with coverage report"
	@echo "  make vet        - Run go vet"
	@echo "  make lint       - Run golangci-lint"
	@echo "  make build      - Build all packages"
	@echo "  make examples   - Verify examples compile"
	@echo "  make check      - Run vet + lint + test"
	@echo "  make clean      - Remove generated files"
