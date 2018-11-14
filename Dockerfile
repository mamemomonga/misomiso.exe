FROM golang:1.11.2-alpine3.8

RUN set -xe && \
	apk --update add tzdata && \
	cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
	apk del tzdata && \
	rm -rf /var/cache/apk/*

RUN set -xe && \
	apk --update add \
		make curl git && \
	rm -rf /var/cache/apk/*

ARG APPPATH

ADD . /go/src/${APPPATH}

RUN set -xe && \
	cd /go/src/${APPPATH} && \
	make deps

RUN set -xe && \
	cd /go/src/${APPPATH} && \
	make dcr-release-build
