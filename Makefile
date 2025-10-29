.PHONY: lint test coverage coverage-html vendor clean

export GO111MODULE=on

default: lint test coverage

lint:
	golangci-lint run

test:
	go test -v -cover ./...

coverage:
	@go test -coverprofile=coverage.out ./...
	@echo "Coverage report generated (excluding test_utils.go): coverage.out"

coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

vendor:
	go mod vendor

clean:
	rm -rf ./vendor
