MOCKGEN=$(GOPATH)/bin/mockgen

mocks/mocks.go: $(MOCKGEN)
	$(MOCKGEN) -source=scanner.go -destination=mocks/mocks.go -package=mocks

$(MOCKGEN):
	go get github.com/golang/mock/mockgen
