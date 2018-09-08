#!/bin/bash
set -eu
BASEDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source $BASEDIR/build.env
DCR_IMAGE="builder-$APPNAME"

COMMANDS="build run"

function do_build {
	cd $BASEDIR
	rm -rf var/dist
	mkdir -p var/dist
	
	docker build -t $DCR_IMAGE .
	
	docker run --rm -i \
		-e APPNAME=$APPNAME \
		-e VERSION=$VERSION \
		-e REVISION=$REVISION \
		-v $PWD/var/dist:/go/app/dist \
		-v $PWD/src:/go/app/src:ro \
		$DCR_IMAGE bash -e << 'EOS'
cd /go/app/src
function build {
	set -x
	mkdir -p /go/app/dist/$GOOS-$GOARCH

	go build -ldflags '-w -s -X "main.Version='$VERSION'" -X "main.Revision='$REVISION'" -extldflags "-static"' \
	   	-o /go/app/dist/$GOOS-$GOARCH/$APPNAME

#	upx -9 /go/app/dist/$GOOS-$GOARCH/$APPNAME

	if [ "$GOOS" == "windows" ]; then
		mv /go/app/dist/$GOOS-$GOARCH/$APPNAME /go/app/dist/$GOOS-$GOARCH/$APPNAME.exe
	fi
	set +x
}

# ビルドターゲットはここで設定

GOOS=darwin  GOARCH=amd64 build
# GOOS=linux   GOARCH=amd64 build
# GOOS=linux   GOARCH=arm   build
# GOOS=windows GOARCH=amd64 build
GOOS=windows GOARCH=386   build

EOS

	mkdir -p var/runenv
	cp -f config.yaml var/runenv/
	cp -f var/dist/darwin-amd64/$APPNAME var/runenv/
}

function do_run {
	mkdir -p var/runenv
	cp -f config.yaml var/runenv/
	go build -o var/runenv/$APPNAME src/main.go 
	cd var/runenv
	exec ./$APPNAME
}

function run {
    for i in $COMMANDS; do
    if [ "$i" == "${1:-}" ]; then
        shift
        do_$i $@
        exit 0
    fi
    done
    echo "USAGE: $( basename $0 ) COMMAND"
    echo "COMMANDS:"
    for i in $COMMANDS; do
    echo "   $i"
    done
    exit 1
}

run $@

