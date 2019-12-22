GOFILES=$(shell find . -type f -iname '*.go')

.PHONY: build
build: $(GOFILES)
	go build -o /dev/null *.go

.PHONY: test
test:
	go test -tags=integration ./...

.PHONY: lint
lint:
	golangci-lint run
