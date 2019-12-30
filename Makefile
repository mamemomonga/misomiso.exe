
NAME=misomiso

VERSION  := v$(shell cat version)
REVISION := $(shell git rev-parse --short HEAD)
LDFLAGS   := -ldflags="-s -w -X \"main.Version=$(VERSION)\" -X \"main.Revision=$(REVISION)\" -extldflags \"-static\""
BUILDARGS := CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo $(LDFLAGS)


build: bin/$(NAME)

run:
	go run ./go/cmd/misomiso/ --help

bin/$(NAME):
	$(BUILDARGS) -o bin/$(NAME) ./go/cmd/misomiso/

clean:
	rm -rf bin

