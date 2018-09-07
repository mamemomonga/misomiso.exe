NAME     := misomiso.exe
VERSION  := v0.0.1
REVISION := $(shell git rev-parse --short HEAD)
SRCS     := $(shell find src -type f -name '*.go')
LDFLAGS  := -ldflags="-s -w -X \"main.Version=$(VERSION)\" -X \"main.Revision=$(REVISION)\" -extldflags \"-static\""
GOBUILD  = go build -a -tags netgo -installsuffix netgo $(LDFLAGS) -o ../$@/$(NAME)-$$GOOS-$$GOARCH

dist: $(SRCS)
	mkdir -p dist
	GOOS=windows GOARCH=386 sh -ec 'cd src; $(GOBUILD)'
	mv $@/$(NAME)-windows-386 $@/misomiso.exe
	cp ./src/misomiso.cmd dist/
	cp ./README.txt dist/
	cd dist; 7za a -tzip ../misomiso.zip .

clean:
	rm -rf dist

get:
	cd src; go get -v

fmt:
	find src -name '*.go' -exec go fmt {} \;

.PHONY: clean get fmt

