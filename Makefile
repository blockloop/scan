MOCKGEN=$(GOPATH)/bin/mockgen
GOFILES=$(shell find . -type f -iname '*.go')

build: $(GOFILES)
	go build -o /dev/null *.go
.PHONY: build

test:
	go test -tags=integration ./...
.PHONY: test

internal/mocks/mocks.go: interface.go | $(MOCKGEN)
	$(MOCKGEN) -source=$< -destination=$@ -package=mocks

$(MOCKGEN):
	go get github.com/golang/mock/mockgen
