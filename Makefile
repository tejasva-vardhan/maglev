.PHONY: build clean coverage test run lint

run: build
	./bin/maglev -api-keys=test

build:
	go build -o bin/maglev ./cmd/api

clean:
	go clean
	rm -f maglev
	rm -f coverage.out

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out


check-golangci-lint:
	@which golangci-lint > /dev/null 2>&1 || (echo "Error: golangci-lint is not installed. Please install it by running: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)

lint: check-golangci-lint
	golangci-lint run

test:
	go test ./...
