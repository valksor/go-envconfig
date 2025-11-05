.PHONY: lint test coverage coverage-html vendor clean

export GO111MODULE=on

default: lint test coverage

lint:
	golangci-lint run

test:
	go test -v -cover ./...

coverage:
	go test -race -covermode atomic -coverprofile=covprofile ./...

coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html

vendor:
	go mod vendor

clean:
	rm -rf ./vendor
