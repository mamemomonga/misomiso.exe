#!/bin/bash
set -eu
BASEDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && cd .. && pwd )"
source $BASEDIR/build.env
DCR_IMAGE="builder-$APPNAME"

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
	if [ "$GOOS" == "windows" ]; then
		mv /go/app/dist/$GOOS-$GOARCH/$APPNAME /go/app/dist/$GOOS-$GOARCH/$APPNAME.exe
	fi
	set +x
}

# ビルドターゲットはここで設定

# GOOS=darwin GOARCH=amd64 build
GOOS=windows GOARCH=386  build

upx -9 /go/app/dist/windows-386/misomiso.exe

EOS
