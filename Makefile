APP=bitnob

.PHONY: build test run fmt

build:
	go build ./cmd/$(APP)

test:
	go test ./...

run:
	go run ./cmd/$(APP)

fmt:
	gofmt -w cmd internal
