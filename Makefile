NAME    := misomiso
APPPATH := github.com/mamemomonga/misomiso.exe

SRCS     := $(shell find go -type f -name '*.go')
VERSION  := v$(shell cat version)
REVISION := $(shell git rev-parse --short HEAD)

LDFLAGS   := -ldflags="-s -w -X \"main.Version=$(VERSION)\" -X \"main.Revision=$(REVISION)\" -extldflags \"-static\""
BUILDARGS := CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo $(LDFLAGS)

BUILDER_DOCKER_IMAGE := $(NAME)-builder
export GOBIN := $(shell if [ -z "$$GOBIN" ]; then echo "$$GOPATH/bin"; else echo "$$GOBIN"; fi)

# -----------------------

.PHONY: build deps clean run release

build: bin/$(NAME)

deps: $(GOBIN)/dep
	$(GOBIN)/dep ensure -v

clean:
	rm -rf bin vendor release

run:
	cd go/cmd/$(NAME); go run *.go -config "$(CURDIR)/etc/config.yaml" -target "みそみそ" -regexp "みそ"

# -----------------------

release:
	mkdir -p release
	docker build --build-arg APPPATH=$(APPPATH) -t $(BUILDER_DOCKER_IMAGE) .
	docker run --rm $(BUILDER_DOCKER_IMAGE) tar cC /go/src/$(APPPATH)/release . | tar xC release

dcr-release-build: $(PACKR_FILES)
	mkdir -p release
	GOOS=linux   GOARCH=arm   $(MAKE) dcr-release-build-os-arch
	GOOS=linux   GOARCH=amd64 $(MAKE) dcr-release-build-os-arch
	GOOS=darwin  GOARCH=amd64 $(MAKE) dcr-release-build-os-arch
	GOOS=windows GOARCH=amd64 $(MAKE) dcr-release-build-os-arch
	cd release; mv $(NAME)-windows-amd64 $(NAME)-windows-amd64.exe
	chmod 755 release/*

dcr-release-build-os-arch:
	cd go/cmd/$(NAME); $(BUILDARGS) -o ../../../release/$(NAME)-$(GOOS)-$(GOARCH)

# -----------------------

bin/$(NAME): $(SRCS)
	mkdir -p bin
	cd go/cmd/$(NAME); $(BUILDARGS) -o ../../../bin/$(NAME) 

# -----------------------

$(GOBIN)/dep:
	@if [ "$(shell go env GOARCH)" = "arm" ]; then \
		go get -v -u github.com/golang/dep/cmd/dep ;\
	else \
		curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh ;\
	fi


