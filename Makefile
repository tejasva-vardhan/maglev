.PHONY: build clean coverage test run lint

include .env

run: build
	bin/maglev \
		-data-path=./gtfs.db \
    	-gtfs-url=https://unitrans.ucdavis.edu/media/gtfs/Unitrans_GTFS.zip \
    	-trip-updates-url=https://webservices.umoiq.com/api/gtfs-rt/v1/trip-updates/unitrans \
    	-vehicle-positions-url=https://webservices.umoiq.com/api/gtfs-rt/v1/vehicle-positions/unitrans \
    	-realtime-auth-header-name=x-umo-iq-api-key \
    	-realtime-auth-header-value=$(REALTIME_AUTH_HEADER_VALUE)

build:
	go build -gcflags "all=-N -l" -o bin/maglev ./cmd/api

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
